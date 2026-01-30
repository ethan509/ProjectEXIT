# LottoSmash Project Rules

## Git Branch Strategy

**항상 다음 순서로 머지할 것:**

```
feature/* → develop → main
```

- 새 기능 개발 시 `feature/` 브랜치 생성
- 작업 완료 후 `develop`에 먼저 머지
- `develop`에서 `main`으로 머지
- feature 브랜치 삭제금지
- BRANCH_POLICY.md 내용 참조

## Project Structure

- Go 백엔드 서버 (Echo 프레임워크 아님, 표준 net/http)
- PostgreSQL 데이터베이스
- 설정 파일: `config/config.json`
- 마이그레이션 파일: `migrations/`

## Database Migration

- `autoMigrate: true` 설정 시 서버 시작 시 자동 실행
- golang-migrate 라이브러리 사용
