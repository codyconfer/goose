#!/bin/sh
set -eu

base_version="${BASE_VERSION:-0.1.0}"

describe_semver_tag() {
	git describe --tags --match 'v[0-9]*.[0-9]*.[0-9]*' --match '[0-9]*.[0-9]*.[0-9]*' "$@"
}

short_commit() {
	git rev-parse --short HEAD 2>/dev/null || printf 'nogit'
}

commit_count() {
	git rev-list --count HEAD 2>/dev/null || printf '0'
}

is_dirty() {
	test -n "$(git status --porcelain 2>/dev/null || true)"
}

strip_v() {
	printf '%s' "$1" | sed 's/^v//'
}

add_dirty_metadata() {
	version="$1"
	if is_dirty; then
		case "$version" in
			*+*) printf '%s.dirty\n' "$version" ;;
			*) printf '%s+dirty\n' "$version" ;;
		esac
	else
		printf '%s\n' "$version"
	fi
}

latest_tag="$(describe_semver_tag --abbrev=0 2>/dev/null || true)"

if [ -z "$latest_tag" ]; then
	version="${base_version}-dev.$(commit_count)+g$(short_commit)"
	add_dirty_metadata "$version"
	exit 0
fi

semver="$(strip_v "$latest_tag")"

if describe_semver_tag --exact-match HEAD >/dev/null 2>&1; then
	add_dirty_metadata "$semver"
	exit 0
fi

major="$(printf '%s' "$semver" | cut -d. -f1)"
minor="$(printf '%s' "$semver" | cut -d. -f2)"
patch="$(printf '%s' "$semver" | cut -d. -f3)"
next_patch=$((patch + 1))
commits_since_tag="$(git rev-list --count "${latest_tag}..HEAD" 2>/dev/null || printf '0')"

version="${major}.${minor}.${next_patch}-dev.${commits_since_tag}+g$(short_commit)"
add_dirty_metadata "$version"
