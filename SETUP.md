# LottoSmash 개발 환경 설정 가이드

다른 컴퓨터에서 이 프로젝트를 클론하고 개발 환경을 구축하는 방법입니다.

## 필수 요구사항

다음 프로그램들이 설치되어 있어야 합니다:

### 1. Go 1.24 이상

**설치 확인:**
```bash
go version
```

**설치 방법:**
- [Go 공식 다운로드](https://go.dev/dl/)
- Windows: MSI 설치 파일 다운로드
- Linux/Mac: 패키지 매니저 사용
  ```bash
  # Ubuntu/Debian
  sudo apt install golang-go

  # macOS (Homebrew)
  brew install go
  ```

### 2. Docker Desktop (선택사항, Docker 사용 시)

**설치 확인:**
```bash
docker --version
docker-compose --version
```

**설치 방법:**
- [Docker Desktop 다운로드](https://www.docker.com/products/docker-desktop/)
- Windows/Mac: Docker Desktop 설치
- Linux: Docker Engine + Docker Compose 설치

### 3. Git

**설치 확인:**
```bash
git --version
```

## 빠른 시작

### 1. 프로젝트 클론

```bash
git clone https://github.com/ethan509/LottoSmash.git
cd LottoSmash
```

### 2. 의존성 다운로드

```bash
go mod download
```

### 3. 로컬 실행

```bash
# 직접 실행
go run ./cmd/server

# 또는 빌드 후 실행
go build -o server ./cmd/server
./server
```

서버가 시작되면 http://localhost:8080 에서 접근 가능합니다.

### 4. Docker로 실행 (권장)

```bash
# Docker Compose 사용 (Make 없을 경우)
docker-compose -f docker/docker-compose.yml up -d

# Make 사용 (Linux/Mac/Git Bash)
make docker-up
```

## 개발 환경별 설정

### Windows

#### Make 설치 (선택사항)

**옵션 1: Chocolatey**
```powershell
choco install make
```

**옵션 2: Scoop**
```powershell
scoop install make
```

**옵션 3: Git Bash 사용**
Git Bash를 사용하면 make가 포함되어 있습니다.

#### PowerShell에서 Docker 사용

Make 없이 직접 사용:
```powershell
# 빌드
docker-compose -f docker/docker-compose.yml build

# 실행
docker-compose -f docker/docker-compose.yml up -d

# 로그
docker-compose -f docker/docker-compose.yml logs -f app

# 중지
docker-compose -f docker/docker-compose.yml down
```

### Linux/Mac

Make가 기본적으로 설치되어 있습니다.

```bash
# 빌드
make build

# 실행
make run

# Docker
make docker-up
make docker-logs
make docker-down
```

## VSCode 설정 (권장)

### 추천 확장 프로그램

1. **Go** (golang.go) - Go 언어 지원
2. **Docker** (ms-azuretools.vscode-docker) - Docker 파일 지원
3. **GitLens** (eamodio.gitlens) - Git 기능 강화

### 디버깅 설정

프로젝트에 이미 `.vscode/launch.json`이 포함되어 있습니다.

F5를 눌러 디버깅을 시작할 수 있습니다.

## 디렉토리 구조 확인

```bash
LottoSmash/
├── cmd/server/           # 메인 애플리케이션
├── internal/             # 내부 패키지
├── config/               # 설정 파일
├── docker/               # Docker 관련
├── scripts/              # 유틸리티 스크립트
├── logs/                 # 로그 파일 (자동 생성)
├── Makefile              # Make 명령어
└── README.md             # 프로젝트 개요
```

**주의:** `logs/` 디렉토리는 서버 실행 시 자동으로 생성됩니다.

## 테스트 실행

```bash
# 모든 테스트 실행
go test -v ./...

# 또는 Make 사용
make test
```

## Docker 테스트

### Windows (PowerShell)
```powershell
.\scripts\test-docker.ps1
```

### Linux/Mac (Bash)
```bash
chmod +x scripts/test-docker.sh
./scripts/test-docker.sh
```

## 설정 파일 수정

`config/config.json` 파일을 수정하여 서버 동작을 변경할 수 있습니다:

```json
{
  "server": {
    "address": ":8080",
    "readTimeoutSec": 10,
    "writeTimeoutSec": 10,
    "idleTimeoutSec": 60
  },
  "logging": {
    "level": "debug",
    "dir": "logs"
  },
  "concurrency": {
    "maxConcurrentRequests": 5000,
    "mainLogicWorkerCount": 8,
    "dbWorkerCount": 4,
    "externalWorkerCount": 4
  }
}
```

## API 테스트

서버가 실행 중이면 다음 명령어로 테스트할 수 있습니다:

```bash
# Health check
curl http://localhost:8080/healthz

# Readiness check
curl http://localhost:8080/readyz

# Metrics
curl http://localhost:8080/metrics

# Ping API
curl http://localhost:8080/api/v1/ping
```

## 문제 해결

### 포트 이미 사용 중

```bash
# Windows
netstat -ano | findstr :8080

# Linux/Mac
lsof -i :8080

# 포트 변경 (config/config.json)
"address": ":9090"
```

### Go 모듈 오류

```bash
# 캐시 정리
go clean -modcache

# 의존성 재다운로드
go mod download

# go.sum 재생성
go mod tidy
```

### Docker 빌드 실패

```bash
# 캐시 없이 재빌드
docker-compose -f docker/docker-compose.yml build --no-cache

# 또는 Make 사용
make docker-clean
make docker-build
```

## 개발 워크플로우

### 1. 새 기능 개발

```bash
# 1. 기능 브랜치 생성
git checkout -b feature/new-feature

# 2. 코드 작성
# ... 개발 ...

# 3. 테스트
go test -v ./...

# 4. 로컬 실행 확인
go run ./cmd/server

# 5. Docker 테스트
make docker-up
make docker-logs

# 6. 커밋
git add .
git commit -m "feat: add new feature"

# 7. 푸시
git push origin feature/new-feature
```

### 2. Hot Reload (개발 모드)

설정 파일 변경은 10분마다 자동으로 재로드됩니다.

코드 변경 시 자동 재시작이 필요하면 [Air](https://github.com/cosmtrek/air) 같은 도구를 사용할 수 있습니다.

## 추가 리소스

- [README.md](./README.md) - 프로젝트 개요
- [STRUCTURE.md](./STRUCTURE.md) - 프로젝트 구조 상세
- [docker/DOCKER.md](./docker/DOCKER.md) - Docker 상세 가이드
- [Go 공식 문서](https://go.dev/doc/)
- [chi 라우터 문서](https://github.com/go-chi/chi)

## 지원

문제가 발생하면 GitHub Issues에 등록해주세요:
https://github.com/ethan509/LottoSmash/issues
