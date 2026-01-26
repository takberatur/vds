param(
  [string]$Context = "remote",
  [ValidateSet("all", "backend", "worker", "frontend")]
  [string]$Target = "all",
  [int]$Tail = 300,
  [string]$Since = "",
  [switch]$Follow,
  [switch]$Timestamps,
  [string]$Grep = ""
)

$ErrorActionPreference = "Stop"

docker --context $Context ps | Out-Null

$serviceMap = @{
  backend  = "backend"
  worker   = "worker"
  frontend = "frontend_01"
}

$containerMap = @{
  backend  = "video_downloader_api"
  worker   = "video_downloader_worker"
  frontend = "video_downloader_web_01"
}

$composeFiles = @(
  "-f", "docker-compose.yml",
  "-f", "docker-compose.remote.yml"
)

function Invoke-LogsCommand {
  param(
    [string[]]$DockerArgs
  )

  if ($Grep -and $Grep.Trim() -ne "") {
    docker --context $Context @DockerArgs 2>&1 | Select-String -Pattern $Grep
    return
  }

  docker --context $Context @DockerArgs
}

if ($Follow) {
  $composeArgs = @("compose") + $composeFiles + @("logs", "-f")
  if ($Timestamps) { $composeArgs += "--timestamps" }
  if ($Tail -ge 0) { $composeArgs += @("--tail", "$Tail") }
  if ($Since -and $Since.Trim() -ne "") { $composeArgs += @("--since", $Since) }

  if ($Target -eq "all") {
    Invoke-LogsCommand -DockerArgs $composeArgs
    exit 0
  }

  $composeArgs += $serviceMap[$Target]
  Invoke-LogsCommand -DockerArgs $composeArgs
  exit 0
}

$targets = @()
if ($Target -eq "all") {
  $targets = @("backend", "worker", "frontend")
} else {
  $targets = @($Target)
}

foreach ($t in $targets) {
  $container = $containerMap[$t]
  Write-Output ""
  Write-Output "===== $t ($container) ====="

  $args = @("logs")
  if ($Timestamps) { $args += "-t" }
  if ($Tail -ge 0) { $args += @("--tail", "$Tail") }
  if ($Since -and $Since.Trim() -ne "") { $args += @("--since", $Since) }
  $args += $container

  Invoke-LogsCommand -DockerArgs $args
}

