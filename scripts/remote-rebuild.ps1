param(
  [string]$Context = "remote",
  [string[]]$Services = @("backend", "worker", "frontend_01")
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

docker --context $Context compose @composeFiles up -d --build --remove-orphans @serviceList
