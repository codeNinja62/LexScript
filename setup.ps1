# LexScript — First-time setup script (Windows / PowerShell)
# Run this once after cloning the repository.
#
#   .\setup.ps1
#
# What it does:
#   1. Builds the lexs.exe binary
#   2. Installs the LexScript VS Code extension

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "LexScript Setup" -ForegroundColor Cyan
Write-Host "===============" -ForegroundColor Cyan
Write-Host ""

# ── Step 1: Build the Go binary ──────────────────────────────────────────────
Write-Host "[1/2] Building lexs.exe..." -ForegroundColor Yellow
Push-Location $Root
try {
    & go build -o bin\lexs.exe .
    if ($LASTEXITCODE -ne 0) { throw "go build failed (exit $LASTEXITCODE)" }
    Write-Host "      OK: bin\lexs.exe" -ForegroundColor Green
} finally {
    Pop-Location
}

# ── Step 2: Install the VS Code extension ────────────────────────────────────
Write-Host "[2/2] Installing LexScript VS Code extension..." -ForegroundColor Yellow

$vsix = Join-Path $Root "vscode-extension\lexscript-0.4.0.vsix"
if (-not (Test-Path $vsix)) {
    Write-Host "      VSIX not found at $vsix" -ForegroundColor Red
    Write-Host "      Rebuild it with: cd vscode-extension && npm install && npm run package" -ForegroundColor Red
    exit 1
}

if (-not (Get-Command code -ErrorAction SilentlyContinue)) {
    Write-Host "      'code' command not found. Add VS Code to PATH, then run:" -ForegroundColor Red
    Write-Host "      code --install-extension `"$vsix`"" -ForegroundColor Yellow
    exit 1
}

& code --install-extension $vsix
if ($LASTEXITCODE -ne 0) { throw "Extension install failed (exit $LASTEXITCODE)" }
Write-Host "      OK: LexScript extension installed" -ForegroundColor Green

Write-Host ""
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "Reload VS Code (Ctrl+Shift+P -> 'Developer: Reload Window') then open a .lxs file." -ForegroundColor Cyan
Write-Host ""
