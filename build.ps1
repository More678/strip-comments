# Cross-compile static strip-comments binaries for Linux and Windows into dist\.
# $env:VERSION defaults to the current git tag (or "dev" outside a tagged checkout).
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$version = $env:VERSION
if (-not $version) {
    $version = git describe --tags --always --dirty 2>$null
    if ($LASTEXITCODE -ne 0 -or -not $version) { $version = 'dev' }
}

$targets = @(
    @('linux', 'amd64'), @('linux', 'arm64'),
    @('windows', 'amd64'), @('windows', 'arm64')
)

Remove-Item -Recurse -Force dist -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force dist | Out-Null
$env:CGO_ENABLED = '0'
foreach ($t in $targets) {
    $env:GOOS, $env:GOARCH = $t
    $out = "dist/strip-comments-$($t[0])-$($t[1])"
    if ($t[0] -eq 'windows') { $out += '.exe' }
    Write-Host "building $out ($version)"
    go build -trimpath -ldflags "-s -w -X main.version=$version" -o $out .
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
Remove-Item Env:GOOS, Env:GOARCH

Get-ChildItem dist/strip-comments-* | ForEach-Object {
    "$((Get-FileHash $_ -Algorithm SHA256).Hash.ToLower())  $($_.Name)"
} | Set-Content -Encoding ascii dist/SHA256SUMS
Get-ChildItem dist
