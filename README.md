>[!IMPORTANT]
>Bump has come to an end.
>It was a nice experiment and playground for different AI agents.
>After almost a year, I've decided to scrap it and rebuild from scratch, based on the experience I've gained and my actual use cases.
>So, please welcome https://semtag.dev

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
bump          # Bumps patch version (e.g., v1.2.3 -> v1.2.4)
bump major    # Bumps major version (e.g., v1.2.3 -> v2.0.0)
bump minor    # Bumps minor version (e.g., v1.2.3 -> v1.3.0)
bump patch    # Bumps patch version (e.g., v1.2.3 -> v1.2.4)

# Monorepo: specify package name as last argument
bump pkg/x           # Bumps patch for pkg/x (e.g., pkg/x/v1.2.3 -> pkg/x/v1.2.4)
bump major pkg/x     # Bumps major for pkg/x (e.g., pkg/x/v1.2.3 -> pkg/x/v2.0.0)
bump minor services/api   # Works with any prefix structure

# Other commands
bump undo     # Removes the latest semver git tag
```

## Options

```
--repo, -r       Path to the repository (if not current directory)
--verbose, -v    Print verbose output
--local, -l      If local is set, bump will not error if no remotes are found
--brave, -b      If brave is set, bump will not ask any questions (default: false)
--no-color       Disable colorful output (default: false)
--json           Output a single JSON object to stdout
--prefix         Tag prefix to use for monorepo support (e.g., 'pkg/x')
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

## Monorepo Support

`bump` supports monorepos where each package has its own version tags with prefixes. Simply specify the package name as the last argument:

```bash
# Bump version for a specific package in a monorepo
bump pkg/x              # Bumps patch version (e.g., pkg/x/v1.2.3 -> pkg/x/v1.2.4)
bump major pkg/x        # Bumps major version (e.g., pkg/x/v1.2.3 -> pkg/x/v2.0.0)
bump minor pkg/x        # Bumps minor version (e.g., pkg/x/v1.2.3 -> pkg/x/v1.3.0)

# Different packages in the same repo can have different versions
bump pkg/y              # Works independently of pkg/x tags
bump services/api       # Can use any prefix structure
bump major libs/utils   # Works with all version parts

# You can also use the --prefix flag for backward compatibility
bump --prefix pkg/x     # Same as: bump pkg/x
```

The package name is used exactly as provided - if you want `pkg/x/foo/v1.0.0`, use `bump pkg/x/foo`.

Example output:
```bash
$ bump major pkg/x
• on default branch: main
• no uncommitted changes
• no remote changes
• no unpushed changes
• no new remote tags
• bump tag pkg/x/v1.2.3 => pkg/x/v2.0.0
• tag pkg/x/v2.0.0 created
• tag pkg/x/v2.0.0 pushed
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
- Monorepo support with custom tag prefixes (e.g., `pkg/name/vX.X.X`)
