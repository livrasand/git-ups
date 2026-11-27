# Installs git-ups from the latest GitHub release on Windows.
#   pwsh -Command "iwr https://git-ups.livrasand.com/install.ps1 -useb | iex"
#
# Optional environment variable:
#   GIT_UPS_VERSION  install a specific tag instead of the latest release (e.g. v0.1.1)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repo = 'livrasand/git-ups'
if ($env:GIT_UPS_VERSION) {
    $base = "https://github.com/$repo/releases/download/$($env:GIT_UPS_VERSION)"
} else {
    $base = "https://github.com/$repo/releases/latest/download"
}

switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    default { throw "unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)" }
}

$asset = "git-ups-windows-$arch.exe"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    $binPath = Join-Path $tmp $asset
    $sumsPath = Join-Path $tmp 'SHA256SUMS'

    Write-Host "Downloading $asset..."
    Invoke-WebRequest -UseBasicParsing -Uri "$base/$asset" -OutFile $binPath
    Invoke-WebRequest -UseBasicParsing -Uri "$base/SHA256SUMS" -OutFile $sumsPath

    $line = Get-Content $sumsPath | Where-Object { $_ -match "\s\*?$([regex]::Escape($asset))$" } | Select-Object -First 1
    if (-not $line) { throw "$asset not listed in SHA256SUMS" }
    $expected = ($line -split '\s+')[0].ToLower()
    $actual = (Get-FileHash -Algorithm SHA256 -Path $binPath).Hash.ToLower()
    if ($actual -ne $expected) { throw "checksum mismatch for $asset (expected $expected, got $actual)" }
    Write-Host 'Checksum verified.'

    # `install` copies the binary to %USERPROFILE%\.local\bin as git-ups and git-gost
    # and tells you how to add that folder to PATH.
    & $binPath install
    if ($LASTEXITCODE -ne 0) { throw "git-ups install failed with exit code $LASTEXITCODE" }

    $binDir = Join-Path $env:USERPROFILE '.local\bin'
    Write-Host ''
    Write-Host "git-ups installed in $binDir."
    Write-Host 'Try: git ups version'
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
