# Contributing to Globus Go CLI

Thank you for your interest in contributing to the Globus Go CLI! This document provides guidelines and instructions for contributing.

## Development Setup

1. Fork and clone the repository
2. Install dependencies: `go mod download`
3. Build the project: `make build`

## Development Workflow

1. Create a branch for your feature or bugfix: `git checkout -b feature/your-feature-name`
2. Make your changes
3. Run tests: `make test`
4. Run linters: `make lint`
5. Commit your changes following the [conventional commits](https://www.conventionalcommits.org/) format
6. Push your branch and create a pull request

## Coding Standards

- Follow standard Go coding conventions and idioms
- Use `go fmt` and `golangci-lint` to format and lint your code
- Write tests for new functionality
- Update documentation as needed
- Include SPDX license headers in all new files

## Pull Request Process

1. Ensure your code passes all tests and linting
2. Update the README.md with details of changes if appropriate
3. The PR should work for Linux, macOS, and Windows

## License

By contributing, you agree that your contributions will be licensed under the project's Apache 2.0 license.