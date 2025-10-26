# Release Process

This document describes how to create a new release of the Globus Go CLI.

## Automated Releases with GoReleaser

The project uses [GoReleaser](https://goreleaser.com/) for automated releases. When you push a new tag, GitHub Actions automatically:

1. Builds binaries for multiple platforms (Linux, macOS, Windows)
2. Creates GitHub release with changelog
3. Generates checksums and signs them with Cosign
4. Builds and pushes Docker images (optional)
5. Updates Homebrew tap (optional)

## Prerequisites

### For Manual Testing
```bash
# Install GoReleaser
brew install goreleaser

# Or using Go
go install github.com/goreleaser/goreleaser@latest
```

### GitHub Secrets Required

Configure these secrets in your repository settings (Settings > Secrets and variables > Actions):

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `DOCKER_USERNAME` - Docker Hub username (optional, for Docker images)
- `DOCKER_TOKEN` - Docker Hub access token (optional, for Docker images)
- `HOMEBREW_TAP_GITHUB_TOKEN` - Personal access token for Homebrew tap updates (optional)
- `SCOOP_BUCKET_GITHUB_TOKEN` - Personal access token for Scoop bucket updates (optional)

## Release Checklist

### 1. Prepare the Release

- [ ] Ensure all tests pass: `go test ./...`
- [ ] Update `CHANGELOG.md` with release notes
- [ ] Update version in documentation if needed
- [ ] Commit all changes
- [ ] Push to main branch

### 2. Test Locally (Optional)

```bash
# Test the build without releasing
goreleaser build --snapshot --clean

# Test the full release process without publishing
goreleaser release --snapshot --clean --skip=publish
```

### 3. Create and Push Tag

```bash
# Create a new tag (semantic versioning)
git tag -a v3.39.0-3 -m "Release v3.39.0-3"

# Push the tag to trigger the release
git push origin v3.39.0-3
```

### 4. Monitor the Release

1. Go to the Actions tab in GitHub
2. Watch the "Release with GoReleaser" workflow
3. Once complete, check the Releases page

### 5. Verify the Release

- [ ] Download and test binaries for your platform
- [ ] Verify checksums
- [ ] Check Docker images (if enabled)
- [ ] Test Homebrew installation (if tap is configured)
- [ ] Test Scoop installation on Windows (if bucket is configured)

## Release Artifacts

Each release includes:

- **Binaries**: Pre-compiled for Linux, macOS, and Windows (amd64 and arm64)
- **Archives**: `.tar.gz` for Unix, `.zip` for Windows
- **Checksums**: `checksums.txt` with SHA256 hashes
- **Signatures**: Cosign signatures for verification
- **SBOM**: Software Bill of Materials
- **Docker Images**: Multi-arch images (if configured)

## Version Numbering

This project follows semantic versioning with an additional suffix:

```
v[MAJOR].[MINOR].[PATCH]-[BUILD]
```

Where:
- **MAJOR**: Aligns with upstream Globus CLI major version (3)
- **MINOR**: Aligns with upstream Globus CLI minor version (39)
- **PATCH**: Aligns with upstream Globus CLI patch version (0)
- **BUILD**: Our build number (1, 2, 3, etc.)

Example: `v3.39.0-2` means:
- Upstream CLI v3.39.0
- Our 2nd build for this version

## Manual Release Process

If you need to release manually without GitHub Actions:

```bash
# Set up environment
export GITHUB_TOKEN="your_github_token"

# Run GoReleaser
goreleaser release --clean
```

## Troubleshooting

### Build Fails

1. Check Go version matches `.goreleaser.yaml`
2. Ensure all dependencies are available
3. Run `go mod tidy` and `go test ./...`

### Docker Build Fails

1. Verify Docker is running
2. Check Docker Hub credentials
3. Test local Docker build: `docker build -t test .`

### Homebrew Tap Update Fails

1. Verify tap repository exists
2. Check `HOMEBREW_TAP_GITHUB_TOKEN` has correct permissions
3. Ensure tap repository has `Formula` directory

### Scoop Bucket Update Fails

1. Verify scoop-bucket repository exists
2. Check `SCOOP_BUCKET_GITHUB_TOKEN` has correct permissions
3. Ensure bucket repository structure is valid for Scoop

## Rolling Back a Release

If you need to delete a release:

```bash
# Delete the tag locally
git tag -d v3.39.0-3

# Delete the tag remotely
git push origin :refs/tags/v3.39.0-3

# Delete the release on GitHub
# Go to Releases page and delete manually
```

## Post-Release Tasks

After a successful release:

1. Announce the release (blog post, social media, etc.)
2. Update documentation sites if needed
3. Monitor issue tracker for bug reports
4. Plan next release

## Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
