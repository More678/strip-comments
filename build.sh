#!/usr/bin/env bash
# Cross-compile static strip-comments binaries for Linux and Windows into dist/.
# VERSION defaults to the current git tag (or "dev" outside a tagged checkout).
set -euo pipefail
cd "$(dirname "$0")"

version="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"

targets=(
	"linux amd64"
	"linux arm64"
	"windows amd64"
	"windows arm64"
)

rm -rf dist
mkdir -p dist
for t in "${targets[@]}"; do
	read -r goos goarch <<<"$t"
	out="dist/strip-comments-${goos}-${goarch}"
	[[ $goos == windows ]] && out+=".exe"
	echo "building $out ($version)"
	CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
		go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$out" .
done

# The `*` binary-mode marker that Git Bash emits is stripped so the file looks
# the same whether it was produced on Windows or on the Linux CI runner.
(cd dist && sha256sum strip-comments-* | sed 's/ \*/  /' >SHA256SUMS)
ls -l dist
