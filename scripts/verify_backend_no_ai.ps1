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
$script:ServerLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_backend_no_ai_server.log'
$script:ServerErrLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_backend_no_ai_server.err.log'
$script:ListenerLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_backend_no_ai_listener.log'
$script:ListenerErrLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_backend_no_ai_listener.err.log'
$script:RunLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_backend_no_ai_run.log'
$script:TempHashFile = $null
$script:TempListenerFile = $null

$runId = 'verify_no_ai_{0}' -f ([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())
$userName = "${runId}_admin"
$password = 'VerifyNoAI!123'
$sourceName = "${runId}_source"
$channelName = "${runId}_channel"
$routeName = "${runId}_route"
$silenceName = "Quick Silence - ${runId} Alert"

$inputTemplate = '{"alert_id":"{{.event_id}}","alert_name":"{{.alert_name}}","severity":"{{.severity}}","message":"{{.message}}","source":"{{.source_name}}","status":"{{.status}}","trigger_time":"{{.timestamp}}"}'.Replace('\"', '"')
$outputTemplate = '{"title":"[{{.severity}}] {{.alert_name}}","content":"{{.message}}"}'.Replace('\"', '"')
$sampleData = '{"event_id":"evt-template","alert_name":"Template Alert","severity":"P1","message":"template works","source_name":"template-source","status":"firing","timestamp":"2026-01-01T00:00:00Z"}'.Replace('\"', '"')

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
      # retry
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
  param([string]$Sql, [switch]$Raw)

  $result = docker compose -f $script:ComposeFile exec -T $script:ComposePostgresService psql -U postgres -d $script:VerificationDbName -v ON_ERROR_STOP=1 -At -c $Sql
  if ($LASTEXITCODE -ne 0) {
    throw "psql failed: $Sql"
  }

  if ($Raw) {
    return $result
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

function Clear-TestData {
  $sql = @"
DELETE FROM silence_rules WHERE name = '$silenceName';
DELETE FROM alerts WHERE source = '$sourceName';
DELETE FROM route_rules WHERE name = '$routeName';
DELETE FROM channels WHERE name = '$channelName';
DELETE FROM data_sources WHERE name = '$sourceName';
DELETE FROM users WHERE username = '$userName';
"@
  Invoke-Psql -Sql $sql | Out-Null
}

function New-PasswordHash {
  $script:TempHashFile = Join-Path ([System.IO.Path]::GetTempPath()) ("verify_backend_no_ai_hash_{0}.go" -f $PID)
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
  if (Test-Path $script:ListenerErrLog) {
    Remove-Item -LiteralPath $script:ListenerErrLog -Force
  }

  $script:TempListenerFile = Join-Path ([System.IO.Path]::GetTempPath()) ("verify_backend_no_ai_listener_{0}.ps1" -f $PID)
  $listenerCode = @'
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$port = [int]$env:VERIFY_NO_AI_LISTENER_PORT
$output = $env:VERIFY_NO_AI_LISTENER_LOG

$listener = [System.Net.HttpListener]::new()
$listener.Prefixes.Add("http://127.0.0.1:$port/")
$listener.Start()

try {
  while ($true) {
    $context = $listener.GetContext()
    try {
      $reader = [System.IO.StreamReader]::new($context.Request.InputStream, $context.Request.ContentEncoding)
      try {
        $body = $reader.ReadToEnd()
      } finally {
        $reader.Dispose()
      }

      $payload = @{
        method = $context.Request.HttpMethod
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
    } catch {
      $context.Response.StatusCode = 500
      $buffer = [System.Text.Encoding]::UTF8.GetBytes($_.Exception.Message)
      $context.Response.OutputStream.Write($buffer, 0, $buffer.Length)
      $context.Response.Close()
    }
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
  $startInfo.RedirectStandardOutput = $false
  $startInfo.RedirectStandardError = $false
  $startInfo.EnvironmentVariables['VERIFY_NO_AI_LISTENER_PORT'] = [string]$script:ListenerPort
  $startInfo.EnvironmentVariables['VERIFY_NO_AI_LISTENER_LOG'] = $script:ListenerLog

  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $startInfo
  $process.Start() | Out-Null
  $script:ListenerProcess = $process

  Wait-ForTcpPort -HostName '127.0.0.1' -Port $script:ListenerPort
}

function Start-Server {
  if (Test-Path $script:ServerLog) {
    Remove-Item -LiteralPath $script:ServerLog -Force
  }
  if (Test-Path $script:ServerErrLog) {
    Remove-Item -LiteralPath $script:ServerErrLog -Force
  }

  $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
  $startInfo.FileName = 'go'
  $startInfo.Arguments = 'run cmd/server/main.go'
  $startInfo.WorkingDirectory = $script:RepoRoot
  $startInfo.UseShellExecute = $false
  $startInfo.RedirectStandardOutput = $false
  $startInfo.RedirectStandardError = $false
  $startInfo.EnvironmentVariables['JWT_SECRET'] = 'verify-no-ai-secret'
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
      if ($Body -is [string]) {
        $json = $Body
      } else {
        $json = $Body | ConvertTo-Json -Depth 10 -Compress
      }
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
    [int]$TimeoutSeconds = 60,
    [int]$IntervalMilliseconds = 1000
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    $result = & $Action
    if ($result.StatusCode -eq $ExpectedStatus) {
      return $result
    }

    if ($script:ServerProcess -and $script:ServerProcess.HasExited) {
      $log = ''
      if (Test-Path $script:ServerLog) {
        $log = Get-Content -Raw $script:ServerLog
      }
      throw "server exited unexpectedly: $log"
    }

    Start-Sleep -Milliseconds $IntervalMilliseconds
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

  $health = Wait-ForStatus -ExpectedStatus 200 -Action {
    Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/health"
  }
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
  '$sourceName', 'Verify No AI Source', '$runId-api-key', true, 3600, false, 300,
  '$inputTemplate', '$outputTemplate', '[]'::jsonb, true, NOW(), NOW()
);

INSERT INTO channels (name, type, config, enabled, created_at, updated_at)
VALUES ('$channelName', 'webhook', '$channelConfig'::jsonb, true, NOW(), NOW());

INSERT INTO route_rules (
  name, priority, severities, sources, label_matchers, channel_ids, time_ranges, enabled, created_at, updated_at
)
VALUES (
  '$routeName', 1, '["P1"]'::jsonb, '["$sourceName"]'::jsonb, '[]'::jsonb,
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
  $loginBody = $login.Content | ConvertFrom-Json
  $token = $loginBody.token
  Assert (-not [string]::IsNullOrWhiteSpace($token)) 'login response missing token'
  $authHeaders = @{ Authorization = "Bearer $token" }

  $statsBefore = Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts/stats" -Headers $authHeaders
  Assert ($statsBefore.StatusCode -eq 200) 'initial stats failed'
  $statsBeforeBody = $statsBefore.Content | ConvertFrom-Json

  $templateTest = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/webhook/test-template" -Body @{
    template    = $inputTemplate
    sample_data = $sampleData
  }
  Assert ($templateTest.StatusCode -eq 200) 'template test failed'
  Write-Step '/webhook/test-template=200'

  $webhookPayload = @{
    event_id    = "${runId}-evt-1"
    alert_name  = "${runId} Alert"
    severity    = 'P1'
    message     = 'backend no ai verification'
    source_name = $sourceName
    status      = 'firing'
    timestamp   = (Get-Date).ToUniversalTime().ToString('o')
  }
  $webhook = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/webhook/$sourceName" -Headers @{
    'X-API-KEY' = "$runId-api-key"
  } -Body $webhookPayload
  Assert ($webhook.StatusCode -eq 200) 'webhook ingestion failed'
  Write-Step '/webhook/{source}=200'
  $webhookBody = $webhook.Content | ConvertFrom-Json
  Assert ($webhookBody.processed -ge 1) 'webhook did not process any alert'
  $alertId = $webhookBody.alerts[0].alert_id
  Assert (-not [string]::IsNullOrWhiteSpace($alertId)) 'webhook response missing alert id'

  $alerts = Wait-ForStatus -ExpectedStatus 200 -Action {
    Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts" -Headers $authHeaders
  }
  Write-Step '/api/v1/alerts=200'
  $alertsBody = $alerts.Content | ConvertFrom-Json
  $matchedAlert = $alertsBody.list | Where-Object { $_.alert_id -eq $alertId }
  Assert ($null -ne $matchedAlert) 'alerts list missing webhook alert'

  $statsAfterWebhook = Wait-ForStatus -ExpectedStatus 200 -Action {
    Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts/stats" -Headers $authHeaders
  }
  Write-Step '/api/v1/alerts/stats=200'
  $statsAfterWebhookBody = $statsAfterWebhook.Content | ConvertFrom-Json
  Assert (($statsAfterWebhookBody.total -ge ($statsBeforeBody.total + 1)) -or ($statsAfterWebhookBody.firing -ge ($statsBeforeBody.firing + 1))) 'stats did not change after webhook'

  $deadline = (Get-Date).AddSeconds(20)
  $dispatchOk = $false
  while ((Get-Date) -lt $deadline) {
    if (Test-Path $script:ListenerLog) {
      $lines = @(Get-Content $script:ListenerLog)
      if ($lines.Count -gt 0) {
        $dispatchOk = $true
        break
      }
    }
    Start-Sleep -Milliseconds 500
  }
  Assert $dispatchOk 'notification dispatch was not observed'
  Write-Step 'notification_dispatch=ok'

  $ack = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts/$alertId/ack" -Headers $authHeaders -Body @{
    comment = 'verified by script'
  }
  Assert ($ack.StatusCode -eq 200) 'ack failed'
  Write-Step '/api/v1/alerts/{id}/ack=200'

  $quickSilence = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts/$alertId/quick-silence" -Headers $authHeaders -Body @{
    duration = 300
  }
  Assert ($quickSilence.StatusCode -eq 200) 'quick silence failed'
  Write-Step '/api/v1/alerts/{id}/quick-silence=200'

  $statsAfterActions = Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts/stats" -Headers $authHeaders
  Assert ($statsAfterActions.StatusCode -eq 200) 'stats after actions failed'
  $statsAfterActionsBody = $statsAfterActions.Content | ConvertFrom-Json
  Assert (($statsAfterActionsBody.silenced -ge ($statsAfterWebhookBody.silenced + 1)) -or ($statsAfterActionsBody.acked -ge ($statsBeforeBody.acked + 1))) 'stats did not reflect ack or quick silence'

  $alertsAfterActions = Invoke-JsonRequest -Method Get -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/alerts" -Headers $authHeaders
  Assert ($alertsAfterActions.StatusCode -eq 200) 'alerts list after actions failed'
  $alertsAfterActionsBody = $alertsAfterActions.Content | ConvertFrom-Json
  $matchedAfterActions = $alertsAfterActionsBody.list | Where-Object { $_.alert_id -eq $alertId }
  Assert ($null -ne $matchedAfterActions) 'alert disappeared after ack/quick silence'

  $aiChat = Invoke-JsonRequest -Method Post -Uri "http://127.0.0.1:$($script:ServerPort)/api/v1/ai/chat" -Headers $authHeaders -Body @{
    message = 'should 404'
  }
  Assert ($aiChat.StatusCode -eq 404) 'ai route did not return 404'
  Write-Step '/api/v1/ai/chat=404'
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
