#!/usr/bin/env bash
# Heretic installer.
#
# Installs the latest release binary for your platform from GitHub Releases,
# verifying its checksum. If no prebuilt archive is available for this
# platform, falls back to `go install` (requires Go).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/gedwolmen/heretic/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/gedwolmen/heretic/main/install.sh | bash -s -- --version v1.2.3
#   curl -fsSL https://raw.githubusercontent.com/gedwolmen/heretic/main/install.sh | bash -s -- --install-dir "$HOME/.local/bin"
#
set -euo pipefail

OWNER="gedwolmen"
REPO="heretic"
MODULE="github.com/gedwolmen/heretic"
BINARY="heretic"

VERSION=""
INSTALL_DIR=""
WORK=""

cleanup() { [ -n "${WORK:-}" ] && rm -rf "$WORK"; }
trap cleanup EXIT

err() { echo "install: error: $*" >&2; exit 1; }

usage() {
  cat <<EOF
Heretic installer.

Usage: install.sh [options]

Options:
  --version <ver>      Install a specific version (e.g. v1.2.3). Default: latest.
  --install-dir <dir>  Directory to install the binary into.
                       Default: /usr/local/bin (falls back to \$HOME/.local/bin).
  -h, --help           Show this help.
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:-}"; shift 2 ;;
    --install-dir) INSTALL_DIR="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) err "unknown option: $1" ;;
  esac
done

have() { command -v "$1" >/dev/null 2>&1; }

# --- platform detection ------------------------------------------------------
detect_os() {
  case "$(uname -s)" in
    Darwin)  echo "Darwin" ;;
    Linux)   echo "Linux" ;;
    FreeBSD) echo "FreeBSD" ;;
    OpenBSD) echo "OpenBSD" ;;
    NetBSD)  echo "NetBSD" ;;
    MINGW*|MSYS*|CYGWIN*) echo "Windows" ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "x86_64" ;;
    i386|i686)     echo "i386" ;;
    aarch64|arm64) echo "arm64" ;;
    armv7l)        echo "armv7" ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

# --- install directory -------------------------------------------------------
choose_install_dir() {
  [ -n "$INSTALL_DIR" ] && { echo "$INSTALL_DIR"; return; }
  if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then echo "/usr/local/bin"; return; fi
  if [ -d /usr/local/bin ] && have sudo; then echo "/usr/local/bin"; return; fi
  echo "${HOME}/.local/bin"
}

INSTALL_DIR="$(choose_install_dir)"
mkdir -p "$INSTALL_DIR" 2>/dev/null || true

# --- github release helpers --------------------------------------------------
fetch_release_json() {
  local url
  if [ -n "$VERSION" ]; then
    url="https://api.github.com/repos/${OWNER}/${REPO}/releases/tags/${VERSION}"
  else
    url="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
  fi
  curl -fsSL "$url"
}

asset_url() {
  # $1 = release json. Prints the browser_download_url for the matching archive.
  echo "$1" | grep -oE "https://[^\"]*heretic_[^\"]*_${OS}_${ARCH}\.(tar\.gz|zip)" | head -n1
}

checksums_url() {
  echo "$1" | grep -oE "https://[^\"]*checksums\.txt" | head -n1
}

release_tag() {
  echo "$1" | grep '"tag_name"' | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]*)".*/\1/'
}

# --- go install fallback -----------------------------------------------------
go_install_fallback() {
  if ! have go; then
    err "no prebuilt release for ${OS}/${ARCH}, and Go is not installed.
Install Go from https://go.dev/dl/ and re-run, or download an archive from
https://github.com/${OWNER}/${REPO}/releases"
  fi
  local target="$MODULE"
  if [ -n "$VERSION" ]; then
    target="${target}@${VERSION}"
  else
    target="${target}@latest"
  fi
  echo "Falling back to: go install ${target}"
  if ! go install "$target"; then
    err "go install ${target} failed. No release or semver tag found?
Cut a release (see 'task release') or download manually from
https://github.com/${OWNER}/${REPO}/releases"
  fi
  local gobin
  gobin="$(go env GOBIN)"
  [ -z "$gobin" ] && gobin="$(go env GOPATH)/bin"
  echo
  echo "Installed ${BINARY} via go to: ${gobin}"
  echo "Ensure ${gobin} is on your PATH."
  echo "  export PATH=\"${gobin:-\$HOME/go/bin}:\$PATH\""
}

# --- main --------------------------------------------------------------------
main() {
  echo "Resolving latest release for ${OS}/${ARCH}..."
  local json
  if ! json="$(fetch_release_json 2>/dev/null || true)"; then
    :
  fi

  local url
  url="$(asset_url "${json:-}")"
  if [ -z "$url" ] || [ -z "${json:-}" ]; then
    echo "No prebuilt archive found for ${OS}/${ARCH}." >&2
    go_install_fallback
    return
  fi

  local tag
  tag="$(release_tag "$json")"
  [ -z "$tag" ] && tag="latest"

  WORK="$(mktemp -d)"
  local archive="${WORK}/heretic.archive"
  local ext="tar.gz"
  case "$url" in
    *.zip) ext="zip" ;;
  esac

  echo "Downloading ${BINARY} ${tag} (${OS}/${ARCH})..."
  curl -fSL "$url" -o "$archive"

  # checksum verification (best-effort)
  local csum_url
  csum_url="$(checksums_url "$json")"
  if [ -n "$csum_url" ] && { have sha256sum || have shasum; }; then
    local expected
    expected="$(curl -fsSL "$csum_url" | grep -E "heretic_.*_${OS}_${ARCH}\.${ext}$" | awk '{print $1}' || true)"
    if [ -n "$expected" ]; then
      local actual
      if have sha256sum; then
        actual="$(sha256sum "$archive" | awk '{print $1}')"
      else
        actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
      fi
      [ "$expected" = "$actual" ] || err "checksum mismatch: expected ${expected}, got ${actual}"
      echo "Checksum verified."
    fi
  fi

  # extract
  local extractdir="${WORK}/extract"
  mkdir -p "$extractdir"
  if [ "$ext" = "zip" ]; then
    have unzip || err "unzip is required to extract the Windows .zip archive"
    (cd "$extractdir" && unzip -q "$archive")
  else
    tar -xzf "$archive" -C "$extractdir"
  fi

  # locate the binary (archive wraps in a directory)
  local binpath
  binpath="$(find "$extractdir" -type f -name "$BINARY" | head -n1)"
  [ -n "$binpath" ] || err "could not find the ${BINARY} binary in the archive"

  # install
  local dest="${INSTALL_DIR}/${BINARY}"
  local need_sudo=0
  if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ ! -w "$INSTALL_DIR" ]; then
    need_sudo=1
  fi

  echo "Installing to ${dest}..."
  if [ "$need_sudo" = "1" ]; then
    sudo install -m 0755 "$binpath" "$dest"
  else
    install -m 0755 "$binpath" "$dest"
  fi

  echo
  echo "Installed ${BINARY} ${tag} -> ${dest}"
  if ! have "$BINARY"; then
    echo "Note: ${INSTALL_DIR} is not on your PATH. Add it:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
  echo
  echo "Shell completions: ${BINARY} completion bash|zsh|fish"
  echo "Manpage:           ${BINARY} man"
  echo "Update:            re-run this installer."
}

main "$@"
