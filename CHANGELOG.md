# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.6] - 2025-03-27

### Added
- Improved help documentation with examples and better descriptions
- Version increment arguments (`major`, `minor`, `patch`) now visible in help text
- Added CHANGELOG.md to track version history
- Added Changelog section to documentation website

### Changed
- Better command usage descriptions in help text
- Example commands now more clearly show their purpose
- Updated undo command documentation with usage examples

## [0.0.5] - 2025-03-20

### Added
- New `HasRemoteUnfetchedTags()` function to detect and fetch new tags from remote repositories
- Automatic tag fetching from remote before performing operations
- "Brave mode" flag (`--brave`, `-b`) to bypass warnings and continue operations
- Improved error handling with consistent exit behavior

### Changed
- Updated output format to use bullet points (•) instead of symbols (✓, ⚠, ✗)
- Revised documentation with new output examples
- Updated README and HTML documentation to reflect the new features and output format
- Improved messaging for tag operations

### Fixed
- Better handling of repositories without remotes
- Consistent error messaging

## [0.0.4] - 2025-03-15

### Added
- Initial public release
- Basic semantic versioning command-line functionality
- Support for major, minor, and patch version bumping
- Tag creation and pushing to remote repositories
- Undo command for removing the latest tag
- Validation for repository state before operations

[Unreleased]: https://github.com/flaticols/bump/compare/v0.0.6...HEAD
[0.0.6]: https://github.com/flaticols/bump/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/flaticols/bump/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/flaticols/bump/releases/tag/v0.0.4
