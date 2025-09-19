# Release Notes - v3.37.0-1

**Release Date:** September 18, 2025
**Upstream Alignment:** Globus CLI v3.37.0
**SDK Version:** Globus Go SDK v3.63.0-1

## 🚀 What's New

This release updates the Globus Go CLI to align with upstream Globus CLI v3.37.0 and incorporates the latest Globus Go SDK v3.63.0-1 improvements.

### 🔄 Version Alignment

- **Updated to v3.37.0-1** following established versioning pattern
- **Aligned with upstream Globus CLI v3.37.0** (released 18 Sep 2025)
- **Maintains backward compatibility** with existing functionality

### 📦 SDK Updates

- **Updated to Globus Go SDK v3.63.0-1** (latest release)
- Resolved SDK release issues and incorporated latest improvements
- Enhanced stability and performance

### 🔧 Technical Improvements

- All unit tests continue to pass with new SDK version
- Integration tests compile successfully with SDK v3.63.0-1
- Code quality checks continue to pass (`go vet`, `go fmt`)
- Clean build and functional CLI maintained

## ✅ Quality Assurance

- ✅ **All unit tests passing**
- ✅ **Integration tests compile successfully**
- ✅ **Code quality checks** (`go vet`, `go fmt`) passing
- ✅ **Clean build and functional CLI**
- ✅ **Backward compatibility** maintained

## 📚 Documentation

- Updated README with SDK v3.63.0-1 reference and upstream CLI v3.37.0 alignment
- Comprehensive CHANGELOG entry documenting version updates
- All version references updated throughout codebase
- Added proper Apache-2.0 LICENSE file for open source compliance

## 🔗 Links

- [GitHub Release](https://github.com/scttfrdmn/globus-go-cli/releases/tag/v3.37.0-1)
- [Changelog](https://github.com/scttfrdmn/globus-go-cli/blob/main/CHANGELOG.md)
- [Globus Go SDK v3.63.0-1](https://github.com/scttfrdmn/globus-go-sdk)
- [Upstream Globus CLI](https://github.com/globus/globus-cli)

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

This release supersedes [v3.36.0-1](RELEASE_NOTES_V3.36.0-1.md).

---

This release represents a **stable, well-tested version** that maintains full compatibility while staying current with both upstream CLI releases and SDK improvements.