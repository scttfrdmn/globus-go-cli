# Release Process

This document describes the process for creating a new release of the Globus Go CLI.

## Pre-Release Checklist

Before creating a new release, ensure the following **REQUIRED** prerequisites:

1. All targeted features are implemented and working correctly
2. **ALL** tests are passing on all platforms, including tests with real credentials:
   ```bash
   # Set up test credentials
   cp .env.test.example .env.test
   # Edit .env.test with real test credentials

   # Run all unit tests
   make test

   # Run integration tests with real credentials
   make test-integration

   # Run all linting checks
   make lint

   # Verify GitHub Actions CI checks are passing
   # Check the GitHub Actions tab in the repository
   ```
3. All linting checks must pass with no errors or warnings
4. Security scans must pass with no critical issues
5. Cross-platform testing must pass on Linux, macOS, and Windows

Additional requirements:
1. Documentation is up-to-date
2. CHANGELOG.md is updated with all notable changes since the last release
3. Version number is updated in `cmd/root.go`

**IMPORTANT**: No release should be tagged or published until all tests and checks are passing. This is a strict requirement to maintain quality and stability.

## Creating a Release

1. Decide on the next version number following [Semantic Versioning](https://semver.org/)
2. Update the CHANGELOG.md file with the new version and release date
3. Commit the changes:
   ```
   git add CHANGELOG.md
   git commit -m "chore: prepare for release v0.x.y"
   ```
4. Tag the release:
   ```
   git tag -a v0.x.y -m "Release v0.x.y"
   ```
5. Push the changes and tags:
   ```
   git push origin main
   git push origin v0.x.y
   ```

Once the tag is pushed, the CI/CD pipeline will automatically:
1. Run tests to verify the build
2. Build binaries for all supported platforms
3. Create a GitHub release with the built binaries
4. Update the Homebrew formula
5. Push container images to Docker Hub

## Post-Release

After the release is complete:

1. Verify that the GitHub release was created correctly
2. Check that the Homebrew formula was updated
3. Verify that the Docker images were pushed to Docker Hub
4. Announce the release on appropriate channels

## Release Automation

The release process is automated using GitHub Actions and GoReleaser. The configuration for these tools can be found in:

- `.github/workflows/release.yml` - GitHub Action workflow for releases
- `.goreleaser.yml` - GoReleaser configuration for building and publishing releases

## Versioning

This project follows [Semantic Versioning](https://semver.org/). In summary:

- MAJOR version (X.0.0) for incompatible API changes
- MINOR version (0.X.0) for added functionality in a backward compatible manner
- PATCH version (0.0.X) for backward compatible bug fixes