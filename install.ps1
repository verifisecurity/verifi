# Verifi CLI installer for Windows (PowerShell).
#
#   irm https://raw.githubusercontent.com/verifisecurity/verifi/main/install.ps1 | iex
#
# Downloads the right prebuilt verifi.exe for your architecture from GitHub
# Releases, verifies its SHA-256 against the published checksums, installs it to
# a per-user location, and adds that location to your PATH. It prints every step:
# piping a script to a shell should never be a black box, least of all for a
# security tool.
#
# Overrides (env vars):
#   VERIFI_VERSION       tag to install (default: latest), e.g. v0.1.0
#   VERIFI_INSTALL_DIR   where to install (default: %LOCALAPPDATA%\Programs\verifi)
#   VERIFI_NO_MODIFY_PATH set to 1 to skip updating PATH (only prints a hint)

$ErrorActionPreference = 'Stop'

function Say  ($m) { Write-Host "  $m" }
function Warn ($m) { Write-Host "  ! $m" -ForegroundColor Yellow }
function Die  ($m) { Write-Host "`n  error: $m`n" -ForegroundColor Red; exit 1 }

$Repo    = 'verifisecurity/verifi'
$Version = if ($env:VERIFI_VERSION) { $env:VERIFI_VERSION } else { 'latest' }

# --- detect architecture -----------------------------------------------------
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { 'amd64' }
  'ARM64' { 'arm64' }
  'x86'   { Die 'unsupported architecture x86. Verifi ships 64-bit builds only.' }
  default { Die "unsupported architecture '$($env:PROCESSOR_ARCHITECTURE)'." }
}

$asset = "verifi_windows_${arch}.zip"
$base  = if ($Version -eq 'latest') {
  "https://github.com/$Repo/releases/latest/download"
} else {
  "https://github.com/$Repo/releases/download/$Version"
}

Write-Host "`n  Verifi CLI installer`n"
Say "platform: windows/$arch"
Say "release:  $Version"

# --- download into a temp dir ------------------------------------------------
$tmp = Join-Path $env:TEMP ("verifi_" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
try {
  $zip = Join-Path $tmp $asset
  $sumFile = Join-Path $tmp 'checksums.txt'

  Say "downloading $asset ..."
  try { Invoke-WebRequest -Uri "$base/$asset" -OutFile $zip -UseBasicParsing }
  catch { Die "download failed. Is there a published release yet? See https://github.com/$Repo/releases" }

  Say 'downloading checksums.txt ...'
  try { Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $sumFile -UseBasicParsing }
  catch { Die 'could not fetch checksums.txt for verification.' }

  # --- verify the SHA-256 ----------------------------------------------------
  $got  = (Get-FileHash -Algorithm SHA256 -Path $zip).Hash.ToLower()
  $want = $null
  foreach ($ln in Get-Content $sumFile) {
    if ($ln -match "^\s*([0-9a-fA-F]{64})\s+\*?$([regex]::Escape($asset))\s*$") {
      $want = $Matches[1].ToLower(); break
    }
  }
  if (-not $want)      { Die "no checksum listed for $asset." }
  if ($got -ne $want)  { Die "checksum mismatch for $asset (got $got, expected $want)." }
  Say 'checksum ok'

  # --- extract ---------------------------------------------------------------
  Expand-Archive -Path $zip -DestinationPath $tmp -Force
  $exe = Join-Path $tmp 'verifi.exe'
  if (-not (Test-Path $exe)) { Die "archive did not contain verifi.exe." }

  # --- choose an install dir -------------------------------------------------
  $dir = if ($env:VERIFI_INSTALL_DIR) {
    $env:VERIFI_INSTALL_DIR
  } else {
    Join-Path $env:LOCALAPPDATA 'Programs\verifi'
  }
  New-Item -ItemType Directory -Path $dir -Force | Out-Null
  Copy-Item -Path $exe -Destination (Join-Path $dir 'verifi.exe') -Force
  Say "installed: $(Join-Path $dir 'verifi.exe')"

  # --- PATH ------------------------------------------------------------------
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  $onPath = ($userPath -split ';') -contains $dir
  if (-not $onPath) {
    if ($env:VERIFI_NO_MODIFY_PATH -eq '1') {
      Warn "$dir is not on your PATH."
      Say  "Add it:  `$env:Path = `"$dir;`$env:Path`""
    } else {
      $newPath = if ([string]::IsNullOrEmpty($userPath)) { $dir } else { "$userPath;$dir" }
      [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
      $env:Path = "$dir;$env:Path"   # current session too
      Say "added $dir to your user PATH"
      Say 'open a new terminal for other apps to see it'
    }
  }

  Write-Host ''
  & (Join-Path $dir 'verifi.exe') version
  Write-Host "`n  Done. Run ``verifi`` to say hello.`n"
}
finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
