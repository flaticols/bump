package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

type cliOpts struct {
	repoPath string
	version  bool
	verbose  bool
}

func main() {
	opts := cliOpts{}

	flag.BoolVar(&opts.version, "version", false, "Print version information")
	flag.BoolVar(&opts.verbose, "verbose", false, "Print verbose information")
	flag.Parse()

	if opts.version {
		fmt.Println(version)
	}

	if opts.repoPath != "" {
		if err := setBumpWd(opts.repoPath); err != nil {
			color.Red("Failed to change to repository directory")
			os.Exit(1)
		}
	}

	// if ok, err := isDefaultBranch(); err != nil {
	// 	color.Red("Failed to check branch")
	// 	os.Exit(1)
	// } else if !ok {
	// 	color.Red("Not on default branch")
	// 	fmt.Printf("Switch to default branch (%s) first before run bump tool\n", strings.Join(defaultBranches, ","))
	// 	os.Exit(1)
	// }

	// dirt, err := checkLocalChanges()
	// if err != nil {
	// 	color.Red("Failed to check local changes")
	// 	os.Exit(1)
	// }
	// if dirt {
	// 	color.Red("Local changes detected")
	// 	fmt.Println("Commit your changes first before run bump tool")
	// }

	// needFetch, err := checkRemoteChanges()
	// if err != nil {
	// 	color.Red("Failed to check remote changes")
	// 	if opts.verbose {
	// 		fmt.Println(err)
	// 	}
	// 	os.Exit(1)
	// }

	// if needFetch {
	// 	color.Red("Remote changes detected")
	// 	fmt.Println("Fetch remote changes first before run bump tool")
	// }
	//

	ltag, initVer, err := getLatestGitTag()
	if err != nil {
		color.Red("Failed to get latest git tag")
		if opts.verbose {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	var nextVer *semver.Version

	green := color.New(color.FgGreen).SprintFunc()
	if !initVer {
		curVer, err := parseTag(ltag)
		if err != nil {
			color.Red("Failed to parse latest git tag")
			if opts.verbose {
				fmt.Println(err)
			}
			os.Exit(1)
		}
		nv := curVer.IncPatch()
		nextVer = &nv
		blue := color.New(color.FgBlue).SprintFunc()
		fmt.Printf("Version bump %s => %s\n", blue(fmt.Sprintf("v%s", curVer.String())), green(fmt.Sprintf("v%s", nextVer.String())))
	} else {
		nextVer = semver.MustParse("0.0.1")
		fmt.Printf("First version %s\n", green(fmt.Sprintf("v%s", nextVer.String())))
	}

	fmt.Println("Git tag created")
	fmt.Println("Git tag pushed")
}
