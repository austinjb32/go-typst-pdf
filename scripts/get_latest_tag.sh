#!/bin/bash
set -e

# Get the bump type from input
BUMP_TYPE="$1"

# Use gem-release to bump the version
echo "Bumping version with type: $BUMP_TYPE"
gem-release bump --version "$BUMP_TYPE" -p

# Get the latest version from the gemspec file
# This assumes there's a .gemspec file in the current directory
GEMSPEC_FILE=$(find . -name "*.gemspec" | head -n 1)
if [ -z "$GEMSPEC_FILE" ]; then
  echo "Error: No gemspec file found"
  exit 1
fi

# Extract version from gemspec
VERSION=$(grep -oE "version\s*=\s*['\"](.*?)['\"]" "$GEMSPEC_FILE" | grep -oE "[0-9]+\.[0-9]+\.[0-9]+")
if [ -z "$VERSION" ]; then
  echo "Error: Could not extract version from gemspec"
  exit 1
fi

# Format with 'v' prefix as is common for Git tags
TAG="v$VERSION"

# Output the tag for GitHub Actions
echo "tag=$TAG" >> $GITHUB_OUTPUT
echo "Found tag: $TAG"
