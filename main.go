package main

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/jessevdk/go-flags"
	"os"
	"runtime/debug"
	"strings"

	"github.com/fatih/color"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

type options struct {
	RepoPath   string `short:"r" long:"repo" description:"Path to the repository"`
	Version    bool   `long:"version" description:"Print version information"`
	Force      bool   `long:"force" description:"Force version bump"`
	Verbose    bool   `long:"verbose" description:"Print verbose output"`
	Prerelease string `long:"pre" description:"Bump prerelease version"`

	Major struct {
	} `command:"major" description:"Bump major version"`
	Minor struct {
	} `command:"minor" description:"Bump minor version"`
	Patch struct {
	} `command:"patch" description:"Bump patch version"`

	Set struct {
		Args struct {
			Version string `positional-arg-name:"version" description:"Version to set"`
		} `positional-args:"yes"`
	} `command:"set" description:"Set version"`
}

func main() {
	var opts options

	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if !errors.Is(err.(*flags.Error).Type, flags.ErrHelp) {
			// cli error, not help
			fmt.Printf("%v", err)
			os.Exit(1)
		}
		os.Exit(0) // help printed
	}

	if opts.Version {
		info, _ := debug.ReadBuildInfo()
		fmt.Printf("v%s\n", info.Main.Version)
		os.Exit(0)
	}

	if opts.RepoPath != "" {
		if err := setBumpWd(opts.RepoPath); err != nil {
			color.Red("Failed to change to repository directory")
			os.Exit(1)
		}
	}

	cVer, isInitial, err := getNextVer(opts)
	if err != nil {
		color.Red("Failed to get next version")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if opts.Prerelease != "" {
		v, _ := ver.SetPrerelease(opts.Prerelease)
		ver = &v
	}

	gtxt := color.New(color.FgGreen).SprintFunc()
	if isInitial {
		fmt.Printf("Initializing version %s\n", gtxt(ver.String()))
	} else {
		btxt := color.New(color.FgBlue).SprintFunc()
		fmt.Printf("Bumping version %s => %s\n", btxt(cVer.String()), gtxt(ver.String()))
	}
}

func detectCommand(p *flags.Parser, opts options) {
	ver := cVer
	if p.Command.Active != nil {
		switch p.Command.Active.Name {
		case "major":
			v := cVer.IncMajor()
			ver = &v
		case "minor":
			v := cVer.IncMinor()
			ver = &v
		case "patch":
			v := cVer.IncPatch()
			ver = &v
		case "set":
			ver, err = semver.NewVersion(opts.Set.Args.Version)
			isInitial = false
			if err != nil {
				color.Red("Failed to parse version")
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	} else {
		v := cVer.IncPatch()
		ver = &v
	}
}

func getGitState() (*semver.Version, error) {
	if ok, err := isDefaultBranch(); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("not on default branch, expected one of: %s", strings.Join(defaultBranches, ","))
	}

	dirt, err := checkLocalChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check local changes: %w", err)
	}
	if dirt {
		return nil, fmt.Errorf("local changes detected, commit your changes first")
	}

	needFetch, err := checkRemoteChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check remote changes: %w", err)
	}

	if needFetch {
		return nil, fmt.Errorf("remote changes detected, fetch changes first")
	}

	LatestTag, err := getLatestGitTag()
	if err != nil {
		if errors.Is(err, ErrNoTagsFound) {
			LatestTag = "0.0.0"
		} else {
			return nil, fmt.Errorf("failed to get latest git tag: %w", err)
		}
	}

	return semver.NewVersion(LatestTag)
}
