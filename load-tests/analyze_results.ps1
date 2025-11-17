# Script para analizar results.json de k6
param(
    [string]$JsonFile = "load-tests/results.json"
)

Write-Host "`n$('='*70)" -ForegroundColor Cyan
Write-Host "  Análisis de resultados k6: $JsonFile" -ForegroundColor Cyan
Write-Host "$('='*70)`n" -ForegroundColor Cyan

$metrics = @{}
$durations = @()
$failed = @()
$checks = @{}

Write-Host "Procesando archivo (esto puede tardar unos segundos)..." -ForegroundColor Yellow

Get-Content $JsonFile | ForEach-Object {
    try {
        $data = $_ | ConvertFrom-Json
        
        if ($data.type -eq "Point") {
            $metric = $data.metric
            $value = $data.data.value
            
            # Recolectar duraciones HTTP
            if ($metric -eq "http_req_duration") {
                $durations += $value
            }
            
            # Recolectar fallos
            if ($metric -eq "http_req_failed") {
                $failed += $value
            }
            
            # Recolectar checks
            if ($metric -eq "checks") {
                $checkName = $data.data.tags.check
                if (-not $checks.ContainsKey($checkName)) {
                    $checks[$checkName] = @{passed=0; failed=0}
                }
                if ($value -eq 1) {
                    $checks[$checkName].passed++
                } else {
                    $checks[$checkName].failed++
                }
            }
            
            # Guardar todas las métricas
            if (-not $metrics.ContainsKey($metric)) {
                $metrics[$metric] = @()
            }
            $metrics[$metric] += $value
        }
    } catch {
        # Ignorar líneas con errores de JSON
    }
}

# Calcular estadísticas de latencia
if ($durations.Count -gt 0) {
    $sorted = $durations | Sort-Object
    $avg = ($durations | Measure-Object -Average).Average
    $p50 = $sorted[[Math]::Floor($sorted.Count * 0.50)]
    $p95 = $sorted[[Math]::Floor($sorted.Count * 0.95)]
    $p99 = $sorted[[Math]::Floor($sorted.Count * 0.99)]
    $max = ($durations | Measure-Object -Maximum).Maximum
    
    Write-Host "📊 REQUESTS TOTALES: $($durations.Count)" -ForegroundColor Green
    Write-Host "`n⏱️  LATENCIAS (ms):" -ForegroundColor Yellow
    Write-Host "   Promedio: $([Math]::Round($avg, 2)) ms"
    Write-Host "   p(50):    $([Math]::Round($p50, 2)) ms"
    
    $p95Status = if ($p95 -lt 800) { "✅" } else { "⚠️" }
    Write-Host "   p(95):    $([Math]::Round($p95, 2)) ms  $p95Status"
    Write-Host "   p(99):    $([Math]::Round($p99, 2)) ms"
    Write-Host "   Max:      $([Math]::Round($max, 2)) ms"
}

# Calcular tasa de fallos
if ($failed.Count -gt 0) {
    $totalFailed = ($failed | Measure-Object -Sum).Sum
    $failRate = ($totalFailed / $failed.Count) * 100
    $status = if ($failRate -lt 2) { "✅" } else { "❌" }
    
    Write-Host "`n🚨 TASA DE FALLOS: $([Math]::Round($failRate, 2))% $status" -ForegroundColor $(if ($failRate -lt 2) { "Green" } else { "Red" })
    Write-Host "   Fallidos: $([Math]::Floor($totalFailed)) de $($failed.Count)"
}

# Mostrar checks
if ($checks.Count -gt 0) {
    $totalPassed = ($checks.Values | ForEach-Object { $_.passed } | Measure-Object -Sum).Sum
    $totalFailed = ($checks.Values | ForEach-Object { $_.failed } | Measure-Object -Sum).Sum
    $totalChecks = $totalPassed + $totalFailed
    $successRate = if ($totalChecks -gt 0) { ($totalPassed / $totalChecks) * 100 } else { 0 }
    
    Write-Host "`n✓ CHECKS:" -ForegroundColor Yellow
    Write-Host "   Total: $totalChecks ($([Math]::Round($successRate, 1))% exitosos)"
    Write-Host "   ✅ Pasados: $totalPassed" -ForegroundColor Green
    Write-Host "   ❌ Fallidos: $totalFailed" -ForegroundColor $(if ($totalFailed -gt 0) { "Red" } else { "Gray" })
    
    if ($totalFailed -gt 0) {
        Write-Host "`n   Checks con fallos:" -ForegroundColor Red
        foreach ($check in $checks.GetEnumerator()) {
            if ($check.Value.failed -gt 0) {
                $total = $check.Value.passed + $check.Value.failed
                $failPct = ($check.Value.failed / $total) * 100
                Write-Host "      • $($check.Key): $($check.Value.failed) fallos ($([Math]::Round($failPct, 1))%)" -ForegroundColor Red
            }
        }
    }
}

# Métricas personalizadas
$customMetrics = @("upload_duration", "s3_asset_duration", "vote_failures")
foreach ($metricName in $customMetrics) {
    if ($metrics.ContainsKey($metricName)) {
        $values = $metrics[$metricName] | Where-Object { $_ -gt 0 }
        if ($values.Count -gt 0) {
            $avg = ($values | Measure-Object -Average).Average
            $sorted = $values | Sort-Object
            $p95 = $sorted[[Math]::Floor($sorted.Count * 0.95)]
            
            $displayName = $metricName.ToUpper().Replace("_", " ")
            Write-Host "`n📈 $displayName`:" -ForegroundColor Yellow
            Write-Host "   Promedio: $([Math]::Round($avg, 2))"
            Write-Host "   p(95):    $([Math]::Round($p95, 2))"
            
            if ($metricName -eq "vote_failures") {
                $total = ($values | Measure-Object -Sum).Sum
                Write-Host "   Total:    $([Math]::Floor($total))"
            }
        }
    }
}

# VUs máximos
if ($metrics.ContainsKey("vus")) {
    $maxVus = ($metrics["vus"] | Measure-Object -Maximum).Maximum
    Write-Host "`n👥 VUs MÁXIMOS: $([Math]::Floor($maxVus))" -ForegroundColor Cyan
}

# Data transferida
if ($metrics.ContainsKey("data_received")) {
    $totalReceived = ($metrics["data_received"] | Measure-Object -Sum).Sum
    $totalMB = $totalReceived / (1024 * 1024)
    Write-Host "`n📥 DATA RECIBIDA: $([Math]::Round($totalMB, 2)) MB" -ForegroundColor Cyan
}

if ($metrics.ContainsKey("data_sent")) {
    $totalSent = ($metrics["data_sent"] | Measure-Object -Sum).Sum
    $totalMB = $totalSent / (1024 * 1024)
    Write-Host "📤 DATA ENVIADA: $([Math]::Round($totalMB, 2)) MB" -ForegroundColor Cyan
}

Write-Host "`n$('='*70)`n" -ForegroundColor Cyan


