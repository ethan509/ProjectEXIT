# Docker 설정 테스트 스크립트 (PowerShell)

Write-Host "=== Docker 설정 테스트 시작 ===" -ForegroundColor Cyan
Write-Host ""

# 1. Docker 빌드
Write-Host "1. Docker 이미지 빌드 중..." -ForegroundColor Yellow
docker-compose -f docker/docker-compose.yml build
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker 빌드 실패" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Docker 빌드 성공" -ForegroundColor Green
Write-Host ""

# 2. 컨테이너 시작
Write-Host "2. 컨테이너 시작 중..." -ForegroundColor Yellow
docker-compose -f docker/docker-compose.yml up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 컨테이너 시작 실패" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 컨테이너 시작 성공" -ForegroundColor Green
Write-Host ""

# 3. 컨테이너 상태 확인
Write-Host "3. 컨테이너 상태 확인 중..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
docker-compose -f docker/docker-compose.yml ps
Write-Host ""

# 4. Health check
Write-Host "4. Health check 테스트..." -ForegroundColor Yellow
Start-Sleep -Seconds 5
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/healthz" -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ /healthz - HTTP $($response.StatusCode)" -ForegroundColor Green
    } else {
        Write-Host "❌ /healthz - HTTP $($response.StatusCode) (expected 200)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ /healthz - Failed to connect" -ForegroundColor Red
}
Write-Host ""

# 5. Readiness check
Write-Host "5. Readiness check 테스트..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/readyz" -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ /readyz - HTTP $($response.StatusCode)" -ForegroundColor Green
    } else {
        Write-Host "❌ /readyz - HTTP $($response.StatusCode) (expected 200)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ /readyz - Failed to connect" -ForegroundColor Red
}
Write-Host ""

# 6. Metrics check
Write-Host "6. Metrics 엔드포인트 테스트..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/metrics" -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ /metrics - HTTP $($response.StatusCode)" -ForegroundColor Green
    } else {
        Write-Host "❌ /metrics - HTTP $($response.StatusCode) (expected 200)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ /metrics - Failed to connect" -ForegroundColor Red
}
Write-Host ""

# 7. Ping API 테스트
Write-Host "7. Ping API 테스트..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/ping" -Method Get
    Write-Host "Response: $($response | ConvertTo-Json -Compress)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Ping API Failed" -ForegroundColor Red
}
Write-Host ""

# 8. Echo API 테스트
Write-Host "8. Echo API 테스트..." -ForegroundColor Yellow
try {
    $body = @{
        message = "Hello Docker!"
    } | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/echo" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Response: $($response | ConvertTo-Json -Compress)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Echo API Failed" -ForegroundColor Red
}
Write-Host ""

# 9. 로그 확인
Write-Host "9. 최근 로그 확인..." -ForegroundColor Yellow
docker-compose -f docker/docker-compose.yml logs --tail=20 app
Write-Host ""

Write-Host "=== 테스트 완료 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "컨테이너를 중지하려면: docker-compose -f docker/docker-compose.yml down" -ForegroundColor Yellow
Write-Host "로그를 계속 보려면: docker-compose -f docker/docker-compose.yml logs -f app" -ForegroundColor Yellow