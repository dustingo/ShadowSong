Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$script:RepoRoot = Split-Path -Parent $PSScriptRoot
$script:FrontendDir = Join-Path $script:RepoRoot 'frontend'
$script:BuildLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_frontend_console_baseline_build.log'
$script:BuildErrLog = Join-Path ([System.IO.Path]::GetTempPath()) 'verify_frontend_console_baseline_build.err.log'

function Write-Step {
  param([string]$Message)
  Write-Host $Message
}

function Assert-Command {
  param([string]$Name)

  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "required command not found: $Name"
  }
}

function Invoke-ResidualScan {
  $patterns = @(
    'AIAssistant',
    'aiApi',
    '/api/v1/ai',
    'ai_summary',
    'ai_root_cause',
    'ai_suggestions',
    'ai_tags',
    'ai_severity',
    '问 AI',
    'AI 分析',
    'AI 助手',
    'RobotOutlined'
  )

  $args = @(
    '--hidden',
    '--glob', '!frontend/dist/**',
    '--glob', '!frontend/node_modules/**',
    '--glob', '!frontend/pnpm-lock.yaml',
    '-n',
    ($patterns -join '|'),
    'frontend/src',
    'frontend/index.html'
  )

  & rg @args
  if ($LASTEXITCODE -eq 0) {
    throw 'frontend AI residual scan found unexpected matches'
  }

  if ($LASTEXITCODE -ne 1) {
    throw 'frontend AI residual scan failed'
  }
}

Assert-Command -Name 'pnpm'
Assert-Command -Name 'rg'

if (Test-Path $script:BuildLog) {
  Remove-Item -LiteralPath $script:BuildLog -Force
}
if (Test-Path $script:BuildErrLog) {
  Remove-Item -LiteralPath $script:BuildErrLog -Force
}

Write-Step 'frontend_build=starting'
$build = Start-Process -FilePath 'pnpm.cmd' -ArgumentList 'build' -WorkingDirectory $script:FrontendDir -RedirectStandardOutput $script:BuildLog -RedirectStandardError $script:BuildErrLog -Wait -PassThru -NoNewWindow
if ($build.ExitCode -ne 0) {
  $stderr = if (Test-Path $script:BuildErrLog) { Get-Content -Raw $script:BuildErrLog } else { '' }
  $stdout = if (Test-Path $script:BuildLog) { Get-Content -Raw $script:BuildLog } else { '' }
  throw "pnpm build failed.`nSTDOUT:`n$stdout`nSTDERR:`n$stderr"
}
Write-Step 'frontend_build=passed'

Write-Step 'frontend_ai_residual_scan=starting'
Invoke-ResidualScan
Write-Step 'frontend_ai_residual_scan=passed'
