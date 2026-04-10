Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$script:RepoRoot = Split-Path -Parent $PSScriptRoot
$script:VerificationDbName = 'game_ops_alert_system'
$script:ServerPort = 0
$script:ListenerPort = 0
$script:ComposePostgresService = 'postgres'
$script:ComposeFile = Join-Path $script:RepoRoot 'docker-compose.yml'
$script:ServerProcess = $null
$script:ListenerProcess = $null
$script:ListenerLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_template_passthrough_listener.log'
$script:RunLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_template_passthrough_run.log'
$script:TempHashFile = $null
$script:TempListenerFile = $null

$runId = 'verify_passthrough_{0}' -f ([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())
$userName = "${runId}_admin"
$password = 'VerifyPassThrough!123'
$legacySourceName = "${runId}_legacy"
$rawSourceName = "${runId}_raw"
$channelName = "${runId}_channel"
$routeName = "${runId}_route"

$legacyInputTemplate = '{"alert_id":"{{.event_id}}","alert_name":"{{.alert_name}}","severity":"{{.severity}}","message":"{{.message}}","source":"{{.source_name}}","status":"{{.status}}","trigger_time":"{{.timestamp}}"}'
$legacyOutputTemplate = '{"title":"[{{.severity}}] {{.alert_name}}","content":"{{.message}}"}'
$rawInputTemplate = '{"alert_id":"{{.labels.alertname}}-{{.startsAt}}","alert_name":"{{.labels.alertname}}","severity":"{{.labels.severity}}","message":"{{default .annotations.summary .summary}}","source":"' + $rawSourceName + '","status":"{{.status}}","trigger_time":"{{.startsAt}}"}'
$rawOutputTemplate = '{"title":"{{default .event.summary .alert_name}}","content":"runbook={{default .event.annotations.runbook "none"}}|ticket={{default .event.custom.ticket "none"}}|message={{.message}}"}'

function Write-Step {
  param([string]$Message)
  Add-Content -LiteralPath $script:RunLog -Value $Message
  Write-Host $Message
}

function Get-FreeTcpPort {
  $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
  $listener.Start()
  try {
    return $listener.LocalEndpoint.Port
  } finally {
    $listener.Stop()
  }
}

function Wait-ForTcpPort {
  param(
    [string]$HostName,
    [int]$Port,
    [int]$TimeoutSeconds = 20
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    $tcpClient = [System.Net.Sockets.TcpClient]::new()
    try {
      $async = $tcpClient.ConnectAsync($HostName, $Port)
      if ($async.Wait(500) -and $tcpClient.Connected) {
        return
      }
    } catch {
    } finally {
      $tcpClient.Dispose()
    }
    Start-Sleep -Milliseconds 250
  }

  throw "timed out waiting for TCP port ${HostName}:${Port}"
}

function Stop-ProcessTree {
  param([System.Diagnostics.Process]$Process)

  if (-not $Process) {
    return
  }

  try {
    $null = cmd /c "taskkill /PID $($Process.Id) /T /F"
  } catch {
  }
}

function Invoke-Psql {
  param([string]$Sql)

  $result = docker compose -f $script:ComposeFile exec -T $script:ComposePostgresService psql -U postgres -d $script:VerificationDbName -v ON_ERROR_STOP=1 -At -c $Sql
  if ($LASTEXITCODE -ne 0) {
    throw "psql failed: $Sql"
  }

  return ($result | Where-Object { $_ -ne '' }) -join "`n"
}

function Ensure-VerificationDatabase {
  $exists = docker compose -f $script:ComposeFile exec -T $script:ComposePostgresService psql -U postgres -d postgres -At -c "SELECT 1 FROM pg_database WHERE datname = '$($script:VerificationDbName)'"
  if ($LASTEXITCODE -ne 0) {
    throw 'failed to query verification database presence'
  }

  if (-not ($exists | Where-Object { $_ -eq '1' })) {
    docker compose -f $script:ComposeFile exec -T $script:ComposePostgresService psql -U postgres -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE $($script:VerificationDbName)"
    if ($LASTEXITCODE -ne 0) {
      throw 'failed to create verification database'
    }
  }
}

function Escape-SqlLiteral {
  param([string]$Value)
  return $Value.Replace("'", "''")
}

function Clear-TestData {
  $sql = @"
DELETE FROM alerts WHERE source IN ('$legacySourceName', '$rawSourceName', 'raw-webhook');
DELETE FROM route_rules WHERE name = '$routeName';
DELETE FROM channels WHERE name = '$channelName';
DELETE FROM data_sources WHERE name IN ('$legacySourceName', '$rawSourceName');
DELETE FROM users WHERE username = '$userName';
"@
  Invoke-Psql -Sql $sql | Out-Null
}

function New-PasswordHash {
  $script:TempHashFile = Join-Path ([System.IO.Path]::GetTempPath()) ("verify_template_passthrough_hash_{0}.go" -f $PID)
  $goCode = @"
package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("$password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(hash))
}
"@
  Set-Content -LiteralPath $script:TempHashFile -Value $goCode -Encoding ASCII
  $hash = go run $script:TempHashFile
  if ($LASTEXITCODE -ne 0) {
    throw 'failed to generate bcrypt hash'
  }
  return ($hash | Out-String).Trim()
}

function Start-NotificationListener {
  if (Test-Path $script:ListenerLog) {
    Remove-Item -LiteralPath $script:ListenerLog -Force
  }

  $script:TempListenerFile = Join-Path ([System.IO.Path]::GetTempPath()) ("verify_template_passthrough_listener_{0}.ps1" -f $PID)
  $listenerCode = @'
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$port = [int]$env:VERIFY_TEMPLATE_LISTENER_PORT
$output = $env:VERIFY_TEMPLATE_LISTENER_LOG

$listener = [System.Net.HttpListener]::new()
$listener.Prefixes.Add("http://127.0.0.1:$port/")
$listener.Start()

try {
  while ($true) {
    $context = $listener.GetContext()
    $reader = [System.IO.StreamReader]::new($context.Request.InputStream, $context.Request.ContentEncoding)
    try {
      $body = $reader.ReadToEnd()
    } finally {
      $reader.Dispose()
    }

    $payload = @{
      path = $context.Request.RawUrl
      headers = @{
        'X-Test-Channel' = $context.Request.Headers['X-Test-Channel']
      }
      body = $body
    } | ConvertTo-Json -Compress
    Add-Content -LiteralPath $output -Value $payload

    $buffer = [System.Text.Encoding]::UTF8.GetBytes('ok')
    $context.Response.StatusCode = 200
    $context.Response.ContentType = 'text/plain'
    $context.Response.OutputStream.Write($buffer, 0, $buffer.Length)
    $context.Response.Close()
  }
} finally {
  $listener.Stop()
  $listener.Close()
}
'@
  Set-Content -LiteralPath $script:TempListenerFile -Value $listenerCode -Encoding ASCII

  $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
  $startInfo.FileName = 'pwsh'
  $startInfo.Arguments = "-NoProfile -ExecutionPolicy Bypass -File `"$script:TempListenerFile`""
  $startInfo.WorkingDirectory = $script:RepoRoot
  $startInfo.UseShellExecute = $false
  $startInfo.EnvironmentVariables['VERIFY_TEMPLATE_LISTENER_PORT'] = [string]$script:ListenerPort
  $startInfo.EnvironmentVariables['VERIFY_TEMPLATE_LISTENER_LOG'] = $script:ListenerLog

  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $startInfo
  $process.Start() | Out-Null
  $script:ListenerProcess = $process

  Wait-ForTcpPort -HostName '127.0.0.1' -Port $script:ListenerPort
}

function Start-Server {
  $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
  $startInfo.FileName = 'go'
  $startInfo.Arguments = 'run cmd/server/main.go'
  $startInfo.WorkingDirectory = $script:RepoRoot
  $startInfo.UseShellExecute = $false
  $startInfo.EnvironmentVariables['JWT_SECRET'] = 'verify-passthrough-secret'
  $startInfo.EnvironmentVariables['DB_HOST'] = '127.0.0.1'
  $startInfo.EnvironmentVariables['DB_PORT'] = '5432'
  $startInfo.EnvironmentVariables['DB_USER'] = 'postgres'
  $startInfo.EnvironmentVariables['DB_PASSWORD'] = 'postgres'
  $startInfo.EnvironmentVariables['DB_NAME'] = $script:VerificationDbName
  $startInfo.EnvironmentVariables['DB_SSLMODE'] = 'disable'
  $startInfo.EnvironmentVariables['REDIS_HOST'] = '127.0.0.1'
  $startInfo.EnvironmentVariables['REDIS_PORT'] = '6379'
  $startInfo.EnvironmentVariables['REDIS_PASSWORD'] = ''
  $startInfo.EnvironmentVariables['REDIS_DB'] = '0'
  $startInfo.EnvironmentVariables['SERVER_PORT'] = [string]$script:ServerPort
  $startInfo.EnvironmentVariables['SERVER_MODE'] = 'release'
  $startInfo.EnvironmentVariables['TOKEN_EXPIRY'] = '2h'

  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $startInfo
  $process.Start() | Out-Null
  $script:ServerProcess = $process
}

function Stop-Resources {
  Stop-ProcessTree -Process $script:ServerProcess
  Stop-ProcessTree -Process $script:ListenerProcess

  if ($script:TempHashFile -and (Test-Path $script:TempHashFile)) {
    Remove-Item -LiteralPath $script:TempHashFile -Force
  }

  if ($script:TempListenerFile -and (Test-Path $script:TempListenerFile)) {
    Remove-Item -LiteralPath $script:TempListenerFile -Force
  }
}

function Invoke-JsonRequest {
  param(
    [string]$Method,
    [string]$Uri,
    [hashtable]$Headers,
    [object]$Body
  )

  $handler = [System.Net.Http.HttpClientHandler]::new()
  $client = [System.Net.Http.HttpClient]::new($handler)
  try {
    $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::$Method, $Uri)
    if ($Headers) {
      foreach ($entry in $Headers.GetEnumerator()) {
        $request.Headers.TryAddWithoutValidation($entry.Key, [string]$entry.Value) | Out-Null
      }
    }

    if ($null -ne $Body) {
      $json = if ($Body -is [string]) { $Body } else { $Body | ConvertTo-Json -Depth 10 -Compress }
      $request.Content = [System.Net.Http.StringContent]::new($json, [System.Text.Encoding]::UTF8, 'application/json')
    }

    $response = $client.Send($request)
    $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
    return [pscustomobject]@{
      StatusCode = [int]$response.StatusCode
      Content    = $content
    }
  } catch {
    return [pscustomobject]@{
      StatusCode = 0
      Content    = $_.Exception.Message
    }
  } finally {
    $client.Dispose()
    $handler.Dispose()
  }
}

function Wait-ForStatus {
  param(
    [scriptblock]$Action,
    [int]$ExpectedStatus,
    [int]$TimeoutSeconds = 60
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    $result = & $Action
    if ($result.StatusCode -eq $ExpectedStatus) {
      return $result
    }

    if ($script:ServerProcess -and $script:ServerProcess.HasExited) {
      throw 'server exited unexpectedly'
    }

    Start-Sleep -Milliseconds 500
  }

  throw "timed out waiting for status $ExpectedStatus"
}

function Assert {
  param(
    [bool]$Condition,
    [string]$Message
  )

  if (-not $Condition) {
    throw $Message
  }
}

function Get-NotificationEvents {
  if (-not (Test-Path $script:ListenerLog)) {
    return @()
  }

  $events = @()
  foreach ($line in Get-Content -LiteralPath $script:ListenerLog) {
    if ([string]::IsNullOrWhiteSpace($line)) {
      continue
    }

    $entry = $line | ConvertFrom-Json
    $body = $entry.body | ConvertFrom-Json
    $events += [pscustomobject]@{
      Header  = $entry.headers.'X-Test-Channel'
      Title   = $body.title
      Content = $body.content
    }
  }

  return $events
}

function Wait-ForNotificationCount {
  param([int]$ExpectedCount)

  $deadline = (Get-Date).AddSeconds(20)
  while ((Get-Date) -lt $deadline) {
    $events = @(Get-NotificationEvents)
    if ($events.Count -ge $ExpectedCount) {
      return $events
    }
    Start-Sleep -Milliseconds 500
  }

  throw "expected at least $ExpectedCount notification events"
}

if (Test-Path $script:RunLog) {
  Remove-Item -LiteralPath $script:RunLog -Force
}

try {
  $script:ServerPort = Get-FreeTcpPort
  $script:ListenerPort = Get-FreeTcpPort

  Write-Step 'docker_compose_up=running'
  docker compose -f $script:ComposeFile up -d postgres redis | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw 'failed to start docker compose dependencies'
  }
  Ensure-VerificationDatabase

  Write-Step 'notification_listener=starting'
  Start-NotificationListener
  Start-Sleep -Seconds 1

  Write-Step 'server=starting'
  Start-Server
  Wait-ForStatus -ExpectedStatus 200 -Action {
    Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/health"
  } | Out-Null
  Write-Step '/health=200'

  Clear-TestData

  $passwordHash = New-PasswordHash
  $channelConfig = '{"url":"http://127.0.0.1:' + $script:ListenerPort + '/notify","method":"POST","headers":{"X-Test-Channel":"' + $runId + '"}}'

  $seedSql = @"
INSERT INTO users (username, password_hash, name, email, role, created_at, updated_at)
VALUES ('$userName', '$passwordHash', 'Verify Admin', '$userName@example.com', 'admin', NOW(), NOW());

INSERT INTO data_sources (
  name, display_name, api_key, deduplicate_enabled, deduplicate_window, group_enabled, group_window,
  input_template, output_template, group_by_labels, enabled, created_at, updated_at
)
VALUES (
  '$legacySourceName', 'Legacy Template Source', '${runId}-legacy-key', true, 3600, false, 300,
  '$(Escape-SqlLiteral $legacyInputTemplate)', '$(Escape-SqlLiteral $legacyOutputTemplate)', '[]'::jsonb, true, NOW(), NOW()
), (
  '$rawSourceName', 'Raw Template Source', '${runId}-raw-key', true, 3600, false, 300,
  '$(Escape-SqlLiteral $rawInputTemplate)', '$(Escape-SqlLiteral $rawOutputTemplate)', '[]'::jsonb, true, NOW(), NOW()
);

INSERT INTO channels (name, type, config, enabled, created_at, updated_at)
VALUES ('$channelName', 'webhook', '$(Escape-SqlLiteral $channelConfig)'::jsonb, true, NOW(), NOW());

INSERT INTO route_rules (
  name, priority, severities, sources, label_matchers, channel_ids, time_ranges, enabled, created_at, updated_at
)
VALUES (
  '$routeName', 1, '["P1"]'::jsonb, '["$legacySourceName","$rawSourceName"]'::jsonb, '[]'::jsonb,
  ('[' || (SELECT id::text FROM channels WHERE name = '$channelName') || ']')::jsonb,
  '[]'::jsonb, true, NOW(), NOW()
);
"@
  Invoke-Psql -Sql $seedSql | Out-Null

  $login = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/auth/login" -Body @{
    username = $userName
    password = $password
  }
  Assert ($login.StatusCode -eq 200) 'login failed'
  Write-Step '/api/v1/auth/login=200'

  $legacyWebhook = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/webhook/$legacySourceName" -Headers @{
    'X-API-KEY' = "${runId}-legacy-key"
  } -Body @{
    event_id    = "${runId}-legacy-evt"
    alert_name  = 'Legacy Alert'
    severity    = 'P1'
    message     = 'legacy template body'
    source_name = $legacySourceName
    status      = 'firing'
    timestamp   = (Get-Date).ToUniversalTime().ToString('o')
  }
  Assert ($legacyWebhook.StatusCode -eq 200) 'legacy webhook ingestion failed'
  Write-Step 'legacy_webhook=200'

  $legacyEvents = Wait-ForNotificationCount -ExpectedCount 1
  $legacyEvent = $legacyEvents[-1]
  Assert ($legacyEvent.Title -eq '[P1] Legacy Alert') 'legacy notification title mismatch'
  Assert ($legacyEvent.Content -eq 'legacy template body') 'legacy notification content mismatch'
  Write-Step 'legacy_notification=ok'

  $rawWebhook = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/webhook/$rawSourceName" -Headers @{
    'X-API-KEY' = "${runId}-raw-key"
  } -Body @{
    status      = 'firing'
    startsAt    = (Get-Date).ToUniversalTime().ToString('o')
    summary     = 'raw summary from webhook'
    value       = 187
    labels      = @{
      alertname = 'Raw Payload Alert'
      severity  = 'warning'
      instance  = 'game-raw-01'
    }
    annotations = @{
      summary = 'Latency above threshold'
      runbook = 'https://runbook.internal/game-latency'
    }
    custom      = @{
      ticket = 'OPS-404'
    }
  }
  Assert ($rawWebhook.StatusCode -eq 200) 'raw webhook ingestion failed'
  Write-Step 'raw_webhook=200'

  $rawEvents = Wait-ForNotificationCount -ExpectedCount 2
  $rawEvent = $rawEvents[-1]
  Assert ($rawEvent.Title -eq 'raw summary from webhook') 'raw notification title mismatch'
  Assert ($rawEvent.Content -eq 'runbook=https://runbook.internal/game-latency|ticket=OPS-404|message=Latency above threshold') 'raw notification content mismatch'
  Write-Step 'raw_notification=ok'

  Write-Step 'phase04_passthrough_verification=passed'
} catch {
  Add-Content -LiteralPath $script:RunLog -Value ("ERROR: " + $_.Exception.Message)
  throw
} finally {
  try {
    if ($script:ServerProcess -and -not $script:ServerProcess.HasExited) {
      Clear-TestData
    }
  } catch {
    Add-Content -LiteralPath $script:RunLog -Value ("CLEANUP_ERROR: " + $_.Exception.Message)
  }

  Stop-Resources
}
