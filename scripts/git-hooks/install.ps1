# Install git hooks for Axis project
$hooksDir = Join-Path (git rev-parse --git-dir) 'hooks'
Copy-Item scripts/git-hooks/commit-msg "$hooksDir/commit-msg" -Force
Copy-Item scripts/git-hooks/pre-commit "$hooksDir/pre-commit" -Force
Write-Host 'Git hooks installed successfully.'
