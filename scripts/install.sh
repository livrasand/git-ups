#!/bin/sh
# Installs git-ups from the latest GitHub release.
#   curl -L https://git-ups.livrasand.com/install.sh | sh
#
# Optional environment variables:
#   GIT_UPS_VERSION  install a specific tag instead of the latest release (e.g. v0.1.1)
set -eu

REPO="livrasand/git-ups"
BIN_DIR="$HOME/.local/bin"

fail() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

need curl
need uname

case "$(uname -s)" in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	*) fail "unsupported operating system: $(uname -s). On Windows use install.ps1." ;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	armv7l | armv6l) [ "$os" = linux ] && arch=arm || fail "unsupported architecture: $(uname -m)" ;;
	i386 | i686) [ "$os" = linux ] && arch=386 || fail "unsupported architecture: $(uname -m)" ;;
	*) fail "unsupported architecture: $(uname -m)" ;;
esac

if [ -n "${GIT_UPS_VERSION:-}" ]; then
	base="https://github.com/$REPO/releases/download/$GIT_UPS_VERSION"
else
	base="https://github.com/$REPO/releases/latest/download"
fi

asset="git-ups-$os-$arch"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

printf 'Downloading %s...\n' "$asset"
curl -fsSL "$base/$asset" -o "$tmp/$asset" || fail "download failed: $base/$asset"
curl -fsSL "$base/SHA256SUMS" -o "$tmp/SHA256SUMS" || fail "download failed: $base/SHA256SUMS"

expected=$(awk -v f="$asset" '$2 == f { print $1 }' "$tmp/SHA256SUMS")
[ -n "$expected" ] || fail "$asset not listed in SHA256SUMS"

if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "$tmp/$asset" | awk '{ print $1 }')
else
	need shasum
	actual=$(shasum -a 256 "$tmp/$asset" | awk '{ print $1 }')
fi
[ "$actual" = "$expected" ] || fail "checksum mismatch for $asset (expected $expected, got $actual)"
echo "Checksum verified."

chmod 755 "$tmp/$asset"
# `install` copies the binary to $BIN_DIR as git-ups and git-gost, and adds $BIN_DIR to PATH when possible.
"$tmp/$asset" install

echo
echo "git-ups installed in $BIN_DIR."
echo "Try: git ups version"
case ":$PATH:" in
	*":$BIN_DIR:"*) ;;
	*) echo "Open a new shell, or add $BIN_DIR to your PATH." ;;
esac
