#!/bin/sh
# Install the latest cutroom release.
#
#   curl -fsSL https://raw.githubusercontent.com/NWylynko/cutroom/main/install.sh | sh
#
# Environment:
#   CUTROOM_VERSION      a release tag to install instead of the latest (e.g. v1.0.0)
#   CUTROOM_INSTALL_DIR  where to put the binary (default: ~/.local/bin)
set -eu

repo="NWylynko/cutroom"
install_dir="${CUTROOM_INSTALL_DIR:-$HOME/.local/bin}"

err() { echo "install.sh: $*" >&2; exit 1; }

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) err "unsupported OS $(uname -s); build from source: go install github.com/$repo/cmd/cutroom@latest" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) err "unsupported architecture $(uname -m); build from source: go install github.com/$repo/cmd/cutroom@latest" ;;
esac

if [ -n "${CUTROOM_VERSION:-}" ]; then
  base="https://github.com/$repo/releases/download/$CUTROOM_VERSION"
else
  base="https://github.com/$repo/releases/latest/download"
fi

if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO "$2" "$1"; }
else
  err "needs curl or wget"
fi

name="cutroom_${os}_${arch}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $name from $base ..."
fetch "$base/$name.tar.gz" "$tmp/$name.tar.gz" || err "download failed: $base/$name.tar.gz"
fetch "$base/checksums.txt" "$tmp/checksums.txt" || err "download failed: $base/checksums.txt"

expected="$(grep " $name.tar.gz\$" "$tmp/checksums.txt" | cut -d' ' -f1)"
[ -n "$expected" ] || err "$name.tar.gz is not listed in checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$name.tar.gz" | cut -d' ' -f1)"
else
  actual="$(shasum -a 256 "$tmp/$name.tar.gz" | cut -d' ' -f1)"
fi
[ "$expected" = "$actual" ] || err "checksum mismatch for $name.tar.gz"

tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
mkdir -p "$install_dir"
mv "$tmp/$name/cutroom" "$install_dir/cutroom"
chmod +x "$install_dir/cutroom"

echo "Installed $("$install_dir/cutroom" --version) to $install_dir/cutroom"

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    echo
    echo "$install_dir is not on your PATH. Add it, e.g. in ~/.zshrc or ~/.bashrc:"
    echo "  export PATH=\"$install_dir:\$PATH\""
    ;;
esac

echo
echo "Next: run 'cutroom doctor' to check ffmpeg, whisper.cpp and a speech model."
