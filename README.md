# bump
![Uploading 189E66B5-315F-4A40-A289-8D1A4DF39098.PNG…]()

A command-line tool to easily bump the git tag version of your project using semantic versioning.

> [!WARNING]
> In development, so bugs may occur.

## Installation
### Homebrew

#### Add tap
```bash
brew tap flaticols/apps
```

#### Install

```bash
brew install flaticols/apps/bump
```

### Go

```bash
go install github.com/flaticols/bump@latest
```




## Usage

```bash
# In your git repository
bump          # Bumps patch version (e.g., 1.2.3 -> 1.2.4)
bump major    # Bumps major version (e.g., 1.2.3 -> 2.0.0)
bump minor    # Bumps minor version (e.g., 1.2.3 -> 1.3.0)
bump patch    # Bumps patch version (e.g., 1.2.3 -> 1.2.4)
```

## Options

```
--repo, -r     Path to the repository (if not current directory)
--version      Print version information
--force        Force version bump (ignore repository state)
--verbose      Print verbose output
--pre=VALUE    Add prerelease suffix (e.g., 1.2.3-alpha)
```

## Features

- Automatically detects and increments from the latest git tag
- Validates that you're on a default branch (main, master, etc.)
- Checks for uncommitted local changes
- Ensures you're in sync with remote repository
- Creates and pushes git tags using semantic versioning
