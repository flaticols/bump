# bump
![bump-small](https://github.com/user-attachments/assets/fa47f507-24fd-4a7d-8972-6e77e11aa578)

A command-line tool to easily bump the git tag version of your project using semantic versioning.

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
bump undo     # Removes the latest semver git tag
```

## Options

```
--repo, -r       Path to the repository (if not current directory)
--verbose, -v    Print verbose output
--local, -l      If local is set, bump will not error if no remotes are found
--brave, -b      If brave is set, bump will not ask any questions (default: false)
--no-color       Disable colorful output (default: false)
--version        Print version information
```

## Commands

- `bump [major|minor|patch]` - Bump the version according to semantic versioning
- `bump undo` - Remove the latest semver git tag both locally and from the remote repository

## Example Output

```bash
$ bump
• on default branch: main
• no uncommitted changes
• no remote changes
• no unpushed changes
• no new remote tags
• bump tag v1.2.3 => v1.2.4
• tag v1.2.4 created
• tag v1.2.4 pushed
```

With brave mode:
```bash
$ bump --brave
• brave mode enabled, ignoring warnings and errors
• on default branch: main
• no uncommitted changes
• no remote changes
• no unpushed changes
• no new remote tags
• bump tag v1.2.3 => v1.2.4
• tag v1.2.4 created
• tag v1.2.4 pushed
```

## Features

- Automatically detects and increments from the latest git tag
- Validates that you're on a default branch (main, master, etc.)
- Checks for uncommitted local changes and ensures you're in sync with remote repository
- Detects and fetches new tags from the remote before bumping
- Creates and pushes git tags using semantic versioning
- Provides colorful terminal output with status indicators
- Support for brave mode to bypass warnings and continue operations
- Allows removing the latest tag with the `undo` command
