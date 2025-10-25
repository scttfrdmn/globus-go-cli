# Release Notes - v3.39.0-1

**Release Date:** October 25, 2025
**Upstream Alignment:** Globus CLI v3.39.0
**SDK Version:** Globus Go SDK v3.65.0-1

## 🚀 What's New

This release updates the Globus Go CLI to align with upstream Globus CLI v3.39.0 and incorporates the latest Globus Go SDK v3.65.0-1 improvements.

### 🔄 Version Alignment

- **Updated to v3.39.0-1** following established versioning pattern
- **Aligned with upstream Globus CLI v3.39.0** (released 15 Oct 2025)
- **Maintains backward compatibility** with existing functionality

### 📦 SDK Updates

- **Updated to Globus Go SDK v3.65.0-1** (latest v3.x release)
- **FlowTimer Helpers**: New SDK features for simplified timer creation
  - `CreateFlowTimerOnce()` for one-time flow executions
  - `CreateFlowTimerRecurring()` for ISO 8601 interval-based scheduling
  - `CreateFlowTimerCron()` for cron-based flow scheduling
- **Groups Status Filtering**: Enhanced Groups service with status-based filtering in `ListGroups()`
- Enhanced stability and performance

### 🎯 Current Implementation Status

This release maintains full implementation of:
- ✅ **Auth Commands** (complete) - login, logout, whoami, tokens, identities
- ✅ **Transfer Commands** (complete) - endpoints, ls, mkdir, cp, rm, tasks

Features available in upstream CLI but planned for future releases:
- 📋 **Groups** - Group management commands (planned)
- 📋 **Timers** - Timer and flow scheduling commands (planned)
- 📋 **Search** - Search service commands (planned)
- 📋 **Flows** - Flow management commands (planned)
- 📋 **Compute** - Compute service commands (planned)

> **Note**: While this release uses the latest SDK v3.65.0-1 with Groups and Timers service support, the CLI commands for these services remain in the planned stage. The SDK capabilities are available for future implementation.

### 🔧 Technical Improvements

- All unit tests continue to pass with new SDK version
- Integration tests compile successfully with SDK v3.65.0-1
- Code quality checks continue to pass (`go vet`, `go fmt`)
- Clean build and functional CLI maintained
- Full backward compatibility with v3.37.0-1

## 📝 Upstream Changes Summary

### From Globus CLI v3.38.0
- Group subscription admin verification features (Groups service - planned for CLI)
- Timer transfer include/exclude flags (Timers service - planned for CLI)

### From Globus CLI v3.39.0
- Timer activity status display (Timers service - planned for CLI)
- `timer create flow` command (Timers service - planned for CLI)

## ✅ Quality Assurance

- ✅ **All unit tests passing**
- ✅ **Integration tests compile successfully**
- ✅ **Code quality checks** (`go vet`, `go fmt`) passing
- ✅ **Clean build and functional CLI**
- ✅ **Backward compatibility** maintained

## 📚 Documentation

- Updated README with SDK v3.65.0-1 reference and upstream CLI v3.39.0 alignment
- Comprehensive CHANGELOG entry documenting version updates
- All version references updated throughout codebase
- Clear documentation of implemented vs. planned features

## 🔗 Links

- [GitHub Release](https://github.com/scttfrdmn/globus-go-cli/releases/tag/v3.39.0-1)
- [Changelog](https://github.com/scttfrdmn/globus-go-cli/blob/main/CHANGELOG.md)
- [Globus Go SDK v3.65.0-1](https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v3.65.0-1)
- [Upstream Globus CLI v3.39.0](https://github.com/globus/globus-cli/releases/tag/3.39.0)

## 📋 Installation

### Binary Downloads

Download the latest release for your platform:
- [Latest Release Page](https://github.com/scttfrdmn/globus-go-cli/releases/latest)

### Package Managers

```bash
# Homebrew (macOS/Linux)
brew tap scttfrdmn/globus
brew install globus-go-cli

# Docker
docker run --rm -it scttfrdmn/globus-go-cli:latest auth whoami
```

### From Source

```bash
git clone https://github.com/scttfrdmn/globus-go-cli.git
cd globus-go-cli
go build -o globus
```

## 🎯 Previous Version

This release supersedes [v3.37.0-1](RELEASE_NOTES_V3.37.0-1.md).

## 🔮 Future Roadmap

Upcoming planned features:
- Groups service commands (leveraging SDK v3.65.0-1 capabilities)
- Timers service commands (leveraging SDK v3.65.0-1 FlowTimer helpers)
- Search service commands
- Flows service commands
- Compute service commands

---

This release represents a **stable, well-tested version** that maintains full compatibility while staying current with upstream CLI versioning and using the latest stable SDK release. The focus remains on excellent Auth and Transfer service support, with additional services planned for future releases.
