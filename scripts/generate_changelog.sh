#!/bin/bash
set -e
LAST_TAG="$1"

if [ -z "$LAST_TAG" ]; then
  echo "No previous tag found"
  exit 0
fi

PREV_DATE=$(git log -1 --format=%cI "$LAST_TAG")
gh pr list --state merged --base main --search "merged:>$PREV_DATE" --json number > prs.json

> changelog.txt
for pr in $(jq '.[].number' prs.json); do
  TITLE=$(gh pr view "$pr" --json title --jq '.title')
  AUTHOR=$(gh pr view "$pr" --json author --jq '.author.login')
  URL=$(gh pr view "$pr" --json url --jq '.url')
  BODY=$(gh pr view "$pr" --json body --jq '.body')
  SECTION=$(echo "$BODY" | awk 'BEGIN{IGNORECASE=1} /^##[ ]*Release Changelog[ ]*\(Required\)/{flag=1; next} /^## /{flag=0} flag')
  if [ -n "$SECTION" ]; then
    echo "**$TITLE ([#$pr]($URL)) by @$AUTHOR**" >> changelog.txt
    echo "$SECTION" | sed '/^\s*$/d' >> changelog.txt
    echo "" >> changelog.txt
  fi
done

echo "body<<EOF" >> $GITHUB_OUTPUT
cat changelog.txt >> $GITHUB_OUTPUT
echo "EOF" >> $GITHUB_OUTPUT
