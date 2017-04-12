# Target OS and ARCH
Set-Item Env:GOOS "linux"
Set-Item Env:GOARCH "amd64"

# Actually do the build
Invoke-Expression "go build .\instafetch.go"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Error when building executable"
    Return
} else {
    Write-Host "Build complete"
}

# Remove the explicitly set OS and ARC, go itself will reset these correctly
Remove-Item Env:\GOOS
Remove-Item Env:\GOARCH
