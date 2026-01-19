#!/bin/bash

# Docker 설정 테스트 스크립트

echo "=== Docker 설정 테스트 시작 ==="
echo ""

# 1. Docker 빌드
echo "1. Docker 이미지 빌드 중..."
docker-compose -f docker/docker-compose.yml build
if [ $? -ne 0 ]; then
    echo "❌ Docker 빌드 실패"
    exit 1
fi
echo "✅ Docker 빌드 성공"
echo ""

# 2. 컨테이너 시작
echo "2. 컨테이너 시작 중..."
docker-compose -f docker/docker-compose.yml up -d
if [ $? -ne 0 ]; then
    echo "❌ 컨테이너 시작 실패"
    exit 1
fi
echo "✅ 컨테이너 시작 성공"
echo ""

# 3. 컨테이너 상태 확인
echo "3. 컨테이너 상태 확인 중..."
sleep 10
docker-compose -f docker/docker-compose.yml ps
echo ""

# 4. Health check
echo "4. Health check 테스트..."
sleep 5
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/healthz)
if [ "$HEALTH_RESPONSE" -eq 200 ]; then
    echo "✅ /healthz - HTTP $HEALTH_RESPONSE"
else
    echo "❌ /healthz - HTTP $HEALTH_RESPONSE (expected 200)"
fi
echo ""

# 5. Readiness check
echo "5. Readiness check 테스트..."
READY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/readyz)
if [ "$READY_RESPONSE" -eq 200 ]; then
    echo "✅ /readyz - HTTP $READY_RESPONSE"
else
    echo "❌ /readyz - HTTP $READY_RESPONSE (expected 200)"
fi
echo ""

# 6. Metrics check
echo "6. Metrics 엔드포인트 테스트..."
METRICS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/metrics)
if [ "$METRICS_RESPONSE" -eq 200 ]; then
    echo "✅ /metrics - HTTP $METRICS_RESPONSE"
else
    echo "❌ /metrics - HTTP $METRICS_RESPONSE (expected 200)"
fi
echo ""

# 7. Ping API 테스트
echo "7. Ping API 테스트..."
PING_RESPONSE=$(curl -s http://localhost:8080/api/v1/ping)
echo "Response: $PING_RESPONSE"
echo ""

# 8. Echo API 테스트
echo "8. Echo API 테스트..."
ECHO_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello Docker!"}')
echo "Response: $ECHO_RESPONSE"
echo ""

# 9. 로그 확인
echo "9. 최근 로그 확인..."
docker-compose -f docker/docker-compose.yml logs --tail=20 app
echo ""

echo "=== 테스트 완료 ==="
echo ""
echo "컨테이너를 중지하려면: make docker-down"
echo "로그를 계속 보려면: make docker-logs"