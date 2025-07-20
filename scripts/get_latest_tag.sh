#!/bin/bash
set -e

BUMP_TYPE="$1"
BUMP_OUTPUT=$(gem bump --version "$BUMP_TYPE" -p)

VERSION=$(echo "$BUMP_OUTPUT" | grep -oE 'to [0-9]+\.[0-9]+\.[0-9]+' | awk '{print $2}')

TAG="v$VERSION"

echo "tag=$TAG" >> $GITHUB_OUTPUT
echo "Found tag: $TAG"
