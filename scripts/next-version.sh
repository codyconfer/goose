#!/bin/sh
set -eu

# Computes the next release version by bumping the latest semver tag.
#
#   BUMP=patch|minor|major  (default: patch)
#   BASE_VERSION=X.Y.Z       first release when no tag exists (default: 0.1.0)
#
# Prints the next version prefixed with `v` (e.g. v0.1.0).

bump="${BUMP:-patch}"
base_version="${BASE_VERSION:-0.1.0}"

describe_semver_tag() {
	git describe --tags --match 'v[0-9]*.[0-9]*.[0-9]*' --match '[0-9]*.[0-9]*.[0-9]*' "$@"
}

strip_v() {
	printf '%s' "$1" | sed 's/^v//'
}

latest_tag="$(describe_semver_tag --abbrev=0 2>/dev/null || true)"

if [ -z "$latest_tag" ]; then
	printf 'v%s\n' "$base_version"
	exit 0
fi

semver="$(strip_v "$latest_tag")"
major="$(printf '%s' "$semver" | cut -d. -f1)"
minor="$(printf '%s' "$semver" | cut -d. -f2)"
patch="$(printf '%s' "$semver" | cut -d. -f3)"

case "$bump" in
	major) major=$((major + 1)); minor=0; patch=0 ;;
	minor) minor=$((minor + 1)); patch=0 ;;
	patch) patch=$((patch + 1)) ;;
	*) echo "unknown BUMP '$bump' (want major|minor|patch)" >&2; exit 1 ;;
esac

printf 'v%s.%s.%s\n' "$major" "$minor" "$patch"
