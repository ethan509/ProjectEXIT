# Docker 설정 가이드

## 개요

LottoSmash 프로젝트는 프로덕션 레디 Docker 설정을 포함하고 있습니다.

## 파일 구조

```
.
├── Dockerfile                 # 멀티스테이지 빌드 설정
├── .dockerignore             # Docker 빌드 최적화
├── docker-compose.yml        # 프로덕션 환경
├── docker-compose.dev.yml    # 개발 환경 오버라이드
├── Makefile                  # 편리한 명령어 모음
└── scripts/
    ├── test-docker.sh        # Docker 테스트 스크립트 (Linux/Mac)
    └── test-docker.ps1       # Docker 테스트 스크립트 (Windows)
```

## Dockerfile 설명

### 멀티스테이지 빌드

**Builder Stage:**
- Go 1.24 Alpine 이미지 사용
- 의존성 다운로드 (go mod download)
- 정적 바이너리 빌드 (CGO_ENABLED=0)
- 바이너리 크기 최적화 (-ldflags="-w -s")

**Runtime Stage:**
- 최소한의 Alpine 이미지
- 비루트 유저로 실행 (appuser:1000)
- 헬스체크 내장
- 포트 8080 노출

### 이미지 크기 최적화

- Builder 이미지: ~500MB (빌드 도구 포함)
- Runtime 이미지: ~20MB (바이너리만)
- 압축률: 96% 감소

## docker-compose.yml 설명

### Services

#### 1. db (PostgreSQL)

```yaml
services:
  db:
    image: postgres:15-alpine
    ports:
      - "5434:5434"
    environment:
      PGPORT: 5434
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U lottosmash -d lottosmash_db"]
      interval: 10s
      timeout: 5s
      retries: 5
```

**특징:**
- PostgreSQL 15 Alpine 버전
- 헬스체크로 DB 준비 상태 확인
- 데이터 영구 저장 (Volume)

#### 2. app (LottoSmash Server)

```yaml
services:
  app:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ../logs:/app/logs
      - ../config:/app/config
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
```

**특징:**
- DB가 준비된 후 시작 (depends_on + condition)
- 로그/설정 파일 호스트와 공유
- 헬스체크로 서버 상태 모니터링
- 자동 재시작 (restart: unless-stopped)

## 사용법

### 1. 프로덕션 환경

```bash
# 빌드 및 실행
make docker-up

# 로그 확인
make docker-logs

# 상태 확인
make docker-ps

# 중지
make docker-down

# 완전 삭제 (볼륨 포함)
make docker-clean
```

### 2. 개발 환경

```bash
# 개발 모드로 실행 (소스 코드 마운트)
make dev-up

# 중지
make dev-down
```

### 3. 수동 Docker 명령어

```bash
# 이미지 빌드
docker-compose build

# 백그라운드 실행
docker-compose up -d

# 로그 실시간 확인
docker-compose logs -f app

# 특정 서비스 재시작
docker-compose restart app

# 컨테이너 내부 접속
docker-compose exec app sh

# DB 접속
docker-compose exec db psql -U lottosmash -d lottosmash_db

# 중지 및 삭제
docker-compose down

# 볼륨까지 삭제
docker-compose down -v
```

## 테스트

### 자동 테스트 스크립트

**Windows (PowerShell):**
```powershell
.\scripts\test-docker.ps1
```

**Linux/Mac (Bash):**
```bash
chmod +x scripts/test-docker.sh
./scripts/test-docker.sh
```

### 수동 테스트

```bash
# 서버 시작 후
docker-compose up -d

# Health check
curl http://localhost:8080/healthz

# Readiness check
curl http://localhost:8080/readyz

# Metrics
curl http://localhost:8080/metrics

# Ping API
curl http://localhost:8080/api/v1/ping

# Echo API
curl -X POST http://localhost:8080/api/v1/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello Docker!"}'
```

## 헬스체크

### Docker 헬스체크 상태 확인

```bash
docker inspect lottosmash-app | jq '.[0].State.Health'
```

### 출력 예시

```json
{
  "Status": "healthy",
  "FailingStreak": 0,
  "Log": [
    {
      "Start": "2024-01-17T12:00:00.000000000Z",
      "End": "2024-01-17T12:00:00.500000000Z",
      "ExitCode": 0,
      "Output": ""
    }
  ]
}
```

## 환경변수

### docker-compose.yml에서 설정 가능

```yaml
environment:
  # Database
  DB_HOST: db
  DB_USER: lottosmash
  DB_PASSWORD: lottosmash_pw
  DB_NAME: lottosmash_db
  DB_PORT: 5434

  # Application (필요시 추가)
  LOG_LEVEL: debug
  SERVER_PORT: 8080
```

### .env 파일 사용

`.env` 파일을 생성하여 환경변수 관리:

```bash
# .env
DB_USER=myuser
DB_PASSWORD=mypassword
LOG_LEVEL=info
```

docker-compose.yml에서 참조:

```yaml
environment:
  DB_USER: ${DB_USER}
  DB_PASSWORD: ${DB_PASSWORD}
  LOG_LEVEL: ${LOG_LEVEL:-debug}  # 기본값: debug
```

## 트러블슈팅

### 1. 포트 이미 사용 중

```bash
# 포트 사용 확인 (Windows)
netstat -ano | findstr :8080

# 포트 사용 확인 (Linux/Mac)
lsof -i :8080

# docker-compose.yml에서 포트 변경
ports:
  - "9090:8080"  # 호스트:컨테이너
```

### 2. 컨테이너가 시작하지 않음

```bash
# 로그 확인
docker-compose logs app

# 상세 로그
docker-compose logs --tail=100 app

# 빌드 로그
docker-compose build --no-cache app
```

### 3. DB 연결 실패

```bash
# DB 헬스체크 상태 확인
docker-compose ps

# DB 로그 확인
docker-compose logs db

# DB 연결 테스트
docker-compose exec app ping db
```

### 4. 볼륨 권한 문제

```bash
# 로그 디렉토리 권한 설정
chmod -R 777 logs/

# 또는 소유자 변경
chown -R 1000:1000 logs/
```

### 5. 이미지 빌드 실패

```bash
# 캐시 없이 재빌드
docker-compose build --no-cache

# 모든 컨테이너/이미지 삭제 후 재빌드
docker-compose down -v --rmi all
docker-compose build
```

## 보안 고려사항

### 1. 비루트 유저 실행

Dockerfile에서 appuser(UID:1000)로 실행:

```dockerfile
USER appuser
```

### 2. 읽기 전용 루트 파일시스템 (선택)

docker-compose.yml에 추가:

```yaml
services:
  app:
    read_only: true
    tmpfs:
      - /tmp
      - /app/logs
```

### 3. 리소스 제한

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '1'
          memory: 512M
```

### 4. 네트워크 격리

```yaml
networks:
  frontend:
  backend:

services:
  app:
    networks:
      - frontend
      - backend

  db:
    networks:
      - backend  # DB는 백엔드 네트워크만
```

## 프로덕션 배포

### Docker Swarm

```bash
# Swarm 초기화
docker swarm init

# 스택 배포
docker stack deploy -c docker-compose.yml lottosmash

# 서비스 확인
docker service ls

# 로그 확인
docker service logs lottosmash_app
```

### Kubernetes

```bash
# Deployment YAML 생성 (kompose 사용)
kompose convert

# 배포
kubectl apply -f .

# 상태 확인
kubectl get pods
kubectl logs -f deployment/lottosmash-app
```

## 모니터링

### Docker 통계

```bash
# 실시간 리소스 사용량
docker stats lottosmash-app lottosmash-db

# 단일 컨테이너
docker stats lottosmash-app --no-stream
```

### Prometheus + Grafana

docker-compose.yml에 추가:

```yaml
services:
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```

## 참고 자료

- [Docker 공식 문서](https://docs.docker.com/)
- [Docker Compose 문서](https://docs.docker.com/compose/)
- [Go Docker 베스트 프랙티스](https://docs.docker.com/language/golang/)
- [Alpine Linux](https://alpinelinux.org/)