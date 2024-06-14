#!/bin/bash

# Helper script to release the Go SDK

set -e

# Read the contents of the files into variables
version=$(<internal/release/version)
build=$(<internal/release/version-build)
changelog=$(<internal/release/changelogs/"${version}"-"${build}")

# Check if Github CLI is installed
if ! command -v gh &> /dev/null; then
	echo "gh is not installed";\
	exit 1;\
fi

# Ensure GITHUB_TOKEN env var is set
if [ -z "${GITHUB_TOKEN}" ]; then
  echo "GITHUB_TOKEN environment variable is not set."
  exit 1
fi

git tag -a -s  "v${version}" -m "${version}"

# Get Current Branch Name
branch="$(git rev-parse --abbrev-ref HEAD)"

# Add changes and commit/push to branch
git add .
git commit -S -m "Release v${version}"
git push --set-upstream origin ${branch}

gh release create "v${version}" --title "Release ${version}" --notes "${changelog}" --repo github.com/1Password/onepassword-sdk-go

