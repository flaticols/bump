package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/flaticols/bump/internal/git"
	"github.com/flaticols/bump/internal/tui"
	semver "github.com/flaticols/server"
)

type semVerPart = string

const (
	major semVerPart = "major"
	minor semVerPart = "minor"
	patch semVerPart = "patch"
)

var defaultBranches = []string{"latest", "main", "master", "develop"}

func runBump(cfg *Config, args []string) error {
	incPart := getIncPart(args)
	cfg.Prefix = normalizePrefix(getPackageName(args), cfg.Prefix)

	if err := gitStateChecks(cfg); err != nil {
		return err
	}

	ver, noTags, err := currentVersion(cfg)
	if err != nil {
		return err
	}

	nextVer := incrementVersion(incPart, ver)
	newTag := formatTag(nextVer.Stringv(), cfg.Prefix)

	if noTags {
		slog.Info(fmt.Sprintf("set tag %s", newTag))
	} else {
		oldTag := formatTag(ver.Stringv(), cfg.Prefix)
		slog.Info(fmt.Sprintf("bump tag %s => %s", oldTag, newTag))
	}

	if err := git.CreateTag(newTag); err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("tag %s created", newTag))

	if !cfg.Local {
		sp := tui.NewSpinner(nil, "pushing tag...", cfg.Interactive)
		sp.Start()
		err := git.PushTag(newTag)
		sp.Stop()
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("tag %s pushed", newTag))
	}

	return nil
}

func currentVersion(cfg *Config) (semver.Version, bool, error) {
	ver, err := git.LatestTag(cfg.Prefix)
	if err != nil {
		var tagErr git.SemVerTagError
		if errors.As(err, &tagErr) {
			if tagErr.NoTags {
				slog.Info(fmt.Sprintf("no tags found, starting from %s", git.DefaultVersion))
				v, _ := semver.Parse("0.0.0")
				return v, true, nil
			}
			return semver.Version{}, false, fmt.Errorf("invalid semver tag: %s", tagErr.Tag)
		}
		return semver.Version{}, false, err
	}
	return ver, false, nil
}

func gitStateChecks(cfg *Config) error {
	branch, err := git.CurrentBranch()
	if err != nil {
		return handleErr(cfg, err)
	}

	if !slices.Contains(defaultBranches, branch) {
		if err := handleFail(cfg, fmt.Sprintf("not on default branch (%s)", branch)); err != nil {
			return err
		}
	} else {
		slog.Info(fmt.Sprintf("on default branch (%s)", branch))
	}

	if err := checkBool(cfg, git.HasLocalChanges, "uncommitted changes", "no uncommitted changes"); err != nil {
		return err
	}

	if cfg.Local {
		return nil
	}

	if err := checkBool(cfg, git.HasRemoteChanges, "remote changes, pull first", "no remote changes"); err != nil {
		return err
	}

	if err := checkBool(cfg, func() (bool, error) { return git.HasUnpushedChanges(branch) },
		"unpushed changes", "no unpushed changes"); err != nil {
		return err
	}

	return handleRemoteTags(cfg)
}

func handleErr(cfg *Config, err error) error {
	if err == nil {
		return nil
	}
	slog.Error(err.Error())
	if cfg.Brave {
		return nil
	}
	return err
}

func handleFail(cfg *Config, msg string) error {
	slog.Error(msg)
	if cfg.Brave {
		return nil
	}
	return errors.New(msg)
}

func checkBool(cfg *Config, check func() (bool, error), failMsg, okMsg string) error {
	result, err := check()
	if err != nil {
		return handleErr(cfg, err)
	}
	if result {
		return handleFail(cfg, failMsg)
	}
	slog.Info(okMsg)
	return nil
}

func handleRemoteTags(cfg *Config) error {
	yes, err := git.HasUnfetchedTags()
	if err != nil {
		slog.Warn(err.Error())
		return nil
	}
	if !yes {
		slog.Info("no new remote tags")
		return nil
	}

	slog.Warn("remote has new tags, fetching...")
	if err := git.FetchTags(); err != nil {
		slog.Error(fmt.Sprintf("failed to fetch tags: %s", err))
		if !cfg.Brave {
			return err
		}
	}
	slog.Info("tags fetched successfully")
	return nil
}

func normalizePrefix(pkgName, prefixFlag string) string {
	prefix := prefixFlag
	if pkgName != "" {
		prefix = pkgName
	}
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		return prefix + "/"
	}
	return prefix
}

func formatTag(version, prefix string) string {
	if prefix != "" {
		return prefix + version
	}
	return version
}

func getIncPart(args []string) semVerPart {
	if len(args) > 0 {
		switch args[0] {
		case major, minor, patch:
			return args[0]
		}
	}
	return patch
}

func getPackageName(args []string) string {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 2 {
		return args[1]
	}
	if args[0] != major && args[0] != minor && args[0] != patch {
		return args[0]
	}
	return ""
}

func incrementVersion(part semVerPart, ver semver.Version) semver.Version {
	switch part {
	case major:
		return ver.IncrementMajor()
	case minor:
		return ver.IncrementMinor()
	default:
		return ver.IncrementPatch()
	}
}
