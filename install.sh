#!/bin/sh
# Verifi CLI installer.
#
#   curl -fsSL https://raw.githubusercontent.com/verifi-security-platform/verifi-cli/main/install.sh | sh
#
# Downloads the right prebuilt `verifi` binary for your OS/arch from GitHub
# Releases, verifies its SHA-256 against the published checksums, and installs
# it. It prints every step: piping a script to a shell should never be a black
# box, least of all for a security tool.
#
# Overrides (env vars):
#   VERIFI_VERSION       tag to install (default: latest), e.g. v0.1.0
#   VERIFI_INSTALL_DIR   where to install (default: /usr/local/bin or ~/.local/bin)
set -eu

REPO="verifi-security-platform/verifi-cli"
VERSION="${VERIFI_VERSION:-latest}"

say()  { printf '  %s\n' "$1"; }
warn() { printf '  ! %s\n' "$1" >&2; }
die()  { printf '\n  error: %s\n\n' "$1" >&2; exit 1; }

# --- pick a downloader -------------------------------------------------------
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1" -o "$2"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO "$2" "$1"; }
else
  die "need curl or wget to download."
fi

# --- detect OS / arch --------------------------------------------------------
os=$(uname -s)
case "$os" in
  Linux)  os=linux ;;
  Darwin) os=darwin ;;
  *) die "unsupported OS '$os'. On Windows, grab the binary from https://github.com/$REPO/releases" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) die "unsupported architecture '$arch'." ;;
esac

asset="verifi_${os}_${arch}.tar.gz"

if [ "$VERSION" = "latest" ]; then
  base="https://github.com/$REPO/releases/latest/download"
else
  base="https://github.com/$REPO/releases/download/$VERSION"
fi

printf '\n  Verifi CLI installer\n\n'
say "platform: ${os}/${arch}"
say "release:  ${VERSION}"

# --- download into a temp dir ------------------------------------------------
tmp=$(mktemp -d 2>/dev/null || mktemp -d -t verifi)
trap 'rm -rf "$tmp"' EXIT INT TERM

say "downloading ${asset} ..."
fetch "$base/$asset" "$tmp/$asset" \
  || die "download failed. Is there a published release yet? See https://github.com/$REPO/releases"

say "downloading checksums.txt ..."
fetch "$base/checksums.txt" "$tmp/checksums.txt" \
  || die "could not fetch checksums.txt for verification."

# --- verify the SHA-256 ------------------------------------------------------
if command -v sha256sum >/dev/null 2>&1; then
  sum=$(sha256sum "$tmp/$asset" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  sum=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')
else
  die "no sha256sum/shasum available to verify the download."
fi

want=$(grep " ${asset}\$" "$tmp/checksums.txt" | awk '{print $1}')
[ -n "$want" ] || die "no checksum listed for ${asset}."
[ "$sum" = "$want" ] || die "checksum mismatch for ${asset} (got $sum, expected $want)."
say "checksum ok"

# --- extract -----------------------------------------------------------------
tar -xzf "$tmp/$asset" -C "$tmp" || die "failed to extract ${asset}."
[ -f "$tmp/verifi" ] || die "archive did not contain a 'verifi' binary."
chmod +x "$tmp/verifi"

# --- choose an install dir ---------------------------------------------------
if [ -n "${VERIFI_INSTALL_DIR:-}" ]; then
  dir="$VERIFI_INSTALL_DIR"
elif [ -w /usr/local/bin ] 2>/dev/null; then
  dir="/usr/local/bin"
else
  dir="$HOME/.local/bin"
fi
mkdir -p "$dir" || die "could not create install dir '$dir'."

if mv "$tmp/verifi" "$dir/verifi" 2>/dev/null; then
  :
elif command -v sudo >/dev/null 2>&1 && [ "$dir" = "/usr/local/bin" ]; then
  say "installing to $dir (needs sudo) ..."
  sudo mv "$tmp/verifi" "$dir/verifi" || die "could not install to '$dir'."
else
  die "could not write to '$dir'. Set VERIFI_INSTALL_DIR to a writable path and retry."
fi

say "installed: $dir/verifi"

# --- PATH hint + first run ---------------------------------------------------
case ":$PATH:" in
  *":$dir:"*) : ;;
  *) warn "$dir is not on your PATH. Add it, e.g.:  export PATH=\"$dir:\$PATH\"" ;;
esac

printf '\n'
if command -v verifi >/dev/null 2>&1; then
  verifi version || true
fi
printf '\n  Done. Run `verifi` to say hello.\n\n'
