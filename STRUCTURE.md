# LottoSmash 프로젝트 구조

```
LottoSmash/
├── cmd/
│   └── server/
│       └── main.go              # 애플리케이션 진입점
│
├── internal/
│   ├── config/                  # 설정 관리 및 Hot Reload
│   ├── logger/                  # 구조화 로깅 시스템
│   ├── middleware/              # HTTP 미들웨어 (TxID, Logging, Timeout 등)
│   ├── metrics/                 # Prometheus 메트릭
│   ├── response/                # 공통 응답 포맷
│   ├── scheduler/               # 스케줄러 (Daily/Weekly/Monthly/Yearly)
│   ├── server/                  # HTTP 핸들러 및 라우터
│   ├── txid/                    # 트랜잭션 ID 관리
│   └── worker/                  # 워커 풀
│
├── config/
│   └── config.json              # 애플리케이션 설정 파일
│
├── docker/                      # Docker 관련 파일
│   ├── Dockerfile               # 멀티스테이지 빌드
│   ├── docker-compose.yml       # 프로덕션 환경
│   ├── docker-compose.dev.yml   # 개발 환경
│   ├── .dockerignore            # Docker 빌드 최적화
│   └── DOCKER.md                # Docker 상세 가이드
│
├── scripts/                     # 유틸리티 스크립트
│   ├── test-docker.sh           # Docker 테스트 (Linux/Mac)
│   └── test-docker.ps1          # Docker 테스트 (Windows)
│
├── logs/                        # 로그 파일 디렉토리
│
├── .vscode/                     # VSCode 설정
│   └── launch.json              # 디버깅 설정
│
├── Makefile                     # Make 명령어
├── README.md                    # 프로젝트 개요
├── STRUCTURE.md                 # 이 파일
├── go.mod                       # Go 모듈 정의
└── go.sum                       # Go 의존성 체크섬
```

## 디렉토리 설명

### cmd/
애플리케이션 진입점. `main.go`에서 모든 컴포넌트를 초기화하고 서버를 시작합니다.

### internal/
외부에 노출되지 않는 내부 패키지들. Go 프로젝트의 표준 레이아웃을 따릅니다.

- **config**: JSON 설정 파일 로드 및 Hot Reload
- **logger**: 레벨별 로그 파일, 일일 로테이션, 1GB 분할
- **middleware**: HTTP 미들웨어 체인
- **metrics**: Prometheus 형식 메트릭
- **response**: 표준 JSON 응답 포맷
- **scheduler**: 시간 기반 작업 스케줄러
- **server**: HTTP 핸들러 및 chi 라우터
- **txid**: 요청 추적용 트랜잭션 ID
- **worker**: 채널 기반 워커 풀 (Main/DB/External)

### config/
설정 파일 저장 위치. `config.json`에서 서버 동작을 제어합니다.

### docker/
모든 Docker 관련 파일을 한 곳에 모았습니다.

- **Dockerfile**: 멀티스테이지 빌드 (builder + runtime)
- **docker-compose.yml**: App + PostgreSQL
- **docker-compose.dev.yml**: 개발 환경 오버라이드
- **.dockerignore**: 불필요한 파일 제외
- **DOCKER.md**: 상세한 사용 가이드

### scripts/
프로젝트 관리 및 테스트 스크립트

### logs/
애플리케이션 로그가 저장되는 위치. Git에는 포함되지 않습니다.

## 주요 기능

1. **Graceful Shutdown**: 최대 1분 타임아웃
2. **구조화 로깅**: 레벨별 파일, 일일 로테이션
3. **워커 풀**: 채널 기반 비동기 처리
4. **스케줄러**: 시간 기반 작업 실행
5. **Hot Reload**: 설정 파일 자동 재로드
6. **헬스체크**: /healthz, /readyz 엔드포인트
7. **메트릭**: Prometheus 형식 /metrics
8. **Docker 지원**: 멀티스테이지 빌드, Compose

## 파일 명명 규칙

- 설정: `config.json`
- 로그: `YYYYMMDD.{level}.log` (예: `20260117.info.log`)
- 로그 분할: `YYYYMMDD.{level}.log_{n}` (예: `20260117.info.log_1`)
- 테스트: `*_test.go`
- Mock: `*_mock.go`

## 의존성

- **go-chi/chi/v5**: HTTP 라우터
- **표준 라이브러리**: 나머지는 모두 표준 라이브러리 사용
