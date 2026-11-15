# Human-agent scheduling demo (US1 decision loop).
# Requires: axis-dev.exe built (go build -o axis-dev.exe ./cmd/axis)
#
# Terminal 1: .\axis-dev.exe start
# Terminal 2: run this script
param(
    [string]$AxisExe = ".\axis-dev.exe"
)

$ErrorActionPreference = "Stop"

Write-Host "== human-agent-scheduling demo ==" -ForegroundColor Cyan
Write-Host "Prereq: a runtime is running in another terminal: $AxisExe start`n"

# 1. Submit a task that requires human confirmation (explicit autonomy = ask path)
$task = @{
    task_id     = "demo-ask-" + (Get-Random -Maximum 9999)
    contract_id = "default"
    input       = @{ message = "hello"; goal = "refactor demo module" }
    metadata    = @{ executor = "agent"; "agent.autonomy_level" = "low" }
} | ConvertTo-Json -Depth 5

Write-Host "[1] submitting agent task with autonomy=low (Ask path)..."
$body = @{ task = ($task | ConvertFrom-Json) } | ConvertTo-Json -Depth 8
$tmp = New-TemporaryFile
Set-Content -Path $tmp -Value $body -Encoding utf8
$resp = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:9377/v1/tasks" -ContentType "application/json" -InFile $tmp
$taskId = $resp.task_id
Write-Host "    task $taskId submitted"

# 2. Poll until the request lands in the inbox (waiting_human)
Write-Host "[2] waiting for the decision request to appear in the inbox..."
$callId = $null
for ($i = 0; $i -lt 50; $i++) {
    & $AxisExe inbox list --json | Out-Null
    $inboxJson = & $AxisExe inbox list --json | Out-String
    $inbox = $inboxJson | ConvertFrom-Json
    $pending = @($inbox.records | Where-Object { $_.status -eq "pending" })
    if ($pending.Count -gt 0) { $callId = $pending[0].message.id; break }
    Start-Sleep -Milliseconds 200
}
if (-not $callId) { throw "no pending request appeared" }
Write-Host "    request $callId is pending"

# 3. Show it, then approve with the first option
& $AxisExe inbox show $callId
Write-Host "`n[3] responding 'proceed'..."
& $AxisExe inbox respond --call-id $callId --choice proceed

# 4. Watch the task finish
Write-Host "[4] task status after respond:"
for ($i = 0; $i -lt 50; $i++) {
    $statusJson = & $AxisExe status $taskId --json 2>$null | Out-String
    try { $st = ($statusJson | ConvertFrom-Json).status } catch { $st = "" }
    if ($st -in @("completed", "failed")) { Write-Host "    final status: $st"; break }
    Start-Sleep -Milliseconds 200
}

Write-Host "`nDemo complete. Event trail: .axis/events/tasks.jsonl (human.request.*)"
