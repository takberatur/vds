param(
  [string]$Context = "remote",
  [string[]]$Services = @("backend", "worker")
)

$ErrorActionPreference = "Stop"

docker --context $Context ps | Out-Null

$composeFiles = @(
  "-f", "docker-compose.yml",
  "-f", "docker-compose.remote.yml"
)

$serviceList = @()
foreach ($s in $Services) {
  $serviceList += ($s -split '[,\s]+') | Where-Object { $_ -and $_.Trim() -ne "" }
}

$containersToRemove = @("video_downloader_api", "video_downloader_worker", "video_downloader_setup")
try {
  docker --context $Context rm -f @containersToRemove | Out-Null
} catch {
}

docker --context $Context compose @composeFiles up -d --build --remove-orphans @serviceList

$localCookiesPath = Join-Path $PSScriptRoot "..\\backend\\cookies.txt"
if (Test-Path $localCookiesPath) {
  try {
    docker --context $Context cp $localCookiesPath "video_downloader_api:/app/cookies.txt" | Out-Null
  } catch {
  }
  try {
    docker --context $Context cp $localCookiesPath "video_downloader_worker:/app/cookies.txt" | Out-Null
  } catch {
  }
}