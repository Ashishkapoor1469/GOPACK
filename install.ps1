# GoPack Installer for Windows

Write-Host "Installing GoPack CLI..." -ForegroundColor Cyan

# Check if Go is installed
$goInstalled = Get-Command go -ErrorAction SilentlyContinue
if (-not $goInstalled) {
    Write-Host "Error: Go is not installed. Please install Go from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

# Compile the binary
Write-Host "Building GoPack executable..." -ForegroundColor Yellow
go build -o gp.exe main.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}

# Ensure destination directory exists
$goBin = Join-Path (Join-Path $HOME "go") "bin"
if (-not (Test-Path $goBin)) {
    New-Item -ItemType Directory -Path $goBin | Out-Null
}

# Copy binary to go bin
$targetPath = Join-Path $goBin "gp.exe"
Move-Item -Path "gp.exe" -Destination $targetPath -Force

Write-Host "Installed gp.exe to $targetPath" -ForegroundColor Green

# Check if path is in environment
$envPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $envPath.Contains($goBin)) {
    Write-Host "Warning: $goBin is not in your PATH. You might need to add it to use 'gp' anywhere." -ForegroundColor Yellow
} else {
    Write-Host "GoPack is ready! Type 'gp' to open the interactive TUI." -ForegroundColor Green
}
