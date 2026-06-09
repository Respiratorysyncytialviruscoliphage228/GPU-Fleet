param(
  [Parameter(Mandatory=$true)]
  [string]$AccountId,
  [Parameter(Mandatory=$true)]
  [string]$ApiToken
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$workerDir = Join-Path $root "telemetry-worker"
$workerName = "gpufleet-telemetry"
$databaseName = "gpufleet_telemetry"
$apiBase = "https://api.cloudflare.com/client/v4/accounts/$AccountId"
$headers = @{
  Authorization = "Bearer $ApiToken"
}
$jsonHeaders = @{
  Authorization = "Bearer $ApiToken"
  "Content-Type" = "application/json"
}

function Invoke-CF {
  param(
    [Parameter(Mandatory=$true)][string]$Method,
    [Parameter(Mandatory=$true)][string]$Uri,
    $Body = $null,
    [hashtable]$Headers = $jsonHeaders,
    [string]$ContentType = "application/json"
  )
  $params = @{
    Method = $Method
    Uri = $Uri
    Headers = $Headers
  }
  if ($null -ne $Body) {
    $params.Body = $Body
    $params.ContentType = $ContentType
  }
  $response = Invoke-RestMethod @params
  if ($response.success -eq $false) {
    $messages = @($response.errors | ForEach-Object { $_.message }) -join "; "
    throw "Cloudflare API failed: $messages"
  }
  return $response
}

$dbList = Invoke-CF -Method GET -Uri "$apiBase/d1/database?name=$databaseName&page=1&per_page=50" -Headers $headers
$database = @($dbList.result | Where-Object { $_.name -eq $databaseName } | Select-Object -First 1)
if (-not $database) {
  $createDb = Invoke-CF -Method POST -Uri "$apiBase/d1/database" -Body (@{ name = $databaseName } | ConvertTo-Json -Depth 4)
  $database = $createDb.result
}
$databaseId = $database.uuid
if (-not $databaseId) {
  $databaseId = $database.id
}
if (-not $databaseId) {
  throw "Unable to determine D1 database id"
}

$schema = Get-Content (Join-Path $workerDir "schema.sql") -Raw
$statements = @($schema -split ";\s*(?:\r?\n|$)" | ForEach-Object { $_.Trim() } | Where-Object { $_ })
foreach ($statement in $statements) {
  $body = @{ sql = $statement } | ConvertTo-Json -Depth 8
  Invoke-CF -Method POST -Uri "$apiBase/d1/database/$databaseId/query" -Body $body | Out-Null
}

$wranglerPath = Join-Path $workerDir "wrangler.toml"
$wrangler = Get-Content $wranglerPath -Raw
$wrangler = $wrangler -replace 'database_id = ".*"', "database_id = `"$databaseId`""
Set-Content -LiteralPath $wranglerPath -Value $wrangler -Encoding UTF8

$metadata = @{
  main_module = "index.js"
  compatibility_date = "2026-06-01"
  bindings = @(
    @{
      type = "d1"
      name = "DB"
      database_id = $databaseId
      id = $databaseId
    }
  )
} | ConvertTo-Json -Depth 8 -Compress
$script = Get-Content (Join-Path $workerDir "src\index.js") -Raw

Add-Type -AssemblyName System.Net.Http
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$client = New-Object System.Net.Http.HttpClient
$client.DefaultRequestHeaders.Authorization = New-Object System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", $ApiToken)
$multipart = New-Object System.Net.Http.MultipartFormDataContent
$metadataContent = New-Object System.Net.Http.ByteArrayContent(,$utf8NoBom.GetBytes($metadata))
$metadataContent.Headers.ContentType = [System.Net.Http.Headers.MediaTypeHeaderValue]::Parse("application/json")
$multipart.Add($metadataContent, "metadata")
$scriptContent = New-Object System.Net.Http.ByteArrayContent(,$utf8NoBom.GetBytes($script))
$scriptContent.Headers.ContentType = [System.Net.Http.Headers.MediaTypeHeaderValue]::Parse("application/javascript+module")
$multipart.Add($scriptContent, "index.js", "index.js")
$uploadResponse = $client.PutAsync("$apiBase/workers/scripts/$workerName", $multipart).GetAwaiter().GetResult()
$uploadRaw = $uploadResponse.Content.ReadAsStringAsync().GetAwaiter().GetResult()
$upload = $uploadRaw | ConvertFrom-Json
if (-not $uploadResponse.IsSuccessStatusCode -or $upload.success -eq $false) {
  $messages = @($upload.errors | ForEach-Object { $_.message }) -join "; "
  if (-not $messages) {
    $messages = $uploadRaw
  }
  throw "Worker upload failed: $messages"
}

$subdomainEnabled = $false
try {
  $subdomainBody = @{ enabled = $true } | ConvertTo-Json -Depth 4
  Invoke-CF -Method POST -Uri "$apiBase/workers/scripts/$workerName/subdomain" -Body $subdomainBody | Out-Null
  $subdomainEnabled = $true
} catch {
  $subdomainEnabled = $false
}

$workerURL = "https://$workerName.stlin256.workers.dev"
$summary = @{
  worker_name = $workerName
  database_name = $databaseName
  database_id = $databaseId
  worker_url = $workerURL
  report_url = "$workerURL/v1/report"
  badge_url = "$workerURL/badge"
  workers_dev_enabled = $subdomainEnabled
}
$summary | ConvertTo-Json -Depth 8
