# Go Environment Setup for F Drive
Write-Host "Setting up Go environment on F drive..." -ForegroundColor Green

# Set GOROOT
[System.Environment]::SetEnvironmentVariable('GOROOT', 'F:\Go', 'User')
Write-Host "GOROOT set to F:\Go" -ForegroundColor Green

# Set GOPATH
[System.Environment]::SetEnvironmentVariable('GOPATH', 'F:\GoPath', 'User')
Write-Host "GOPATH set to F:\GoPath" -ForegroundColor Green

# Create GOPATH directories
$dirs = @('F:\GoPath', 'F:\GoPath\src', 'F:\GoPath\pkg', 'F:\GoPath\bin')
foreach ($dir in $dirs) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "Created: $dir" -ForegroundColor Green
    }
}

# Add to PATH
$currentPath = [System.Environment]::GetEnvironmentVariable('Path', 'User')
$goBin = 'F:\Go\bin'
$goPathBin = 'F:\GoPath\bin'

if ($currentPath -notlike "*$goBin*") {
    $newPath = $currentPath + ";$goBin"
    [System.Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "Added F:\Go\bin to PATH" -ForegroundColor Green
}

if ($currentPath -notlike "*$goPathBin*") {
    $currentPath = [System.Environment]::GetEnvironmentVariable('Path', 'User')
    $newPath = $currentPath + ";$goPathBin"
    [System.Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "Added F:\GoPath\bin to PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "Please restart PowerShell and run: go version" -ForegroundColor Cyan
