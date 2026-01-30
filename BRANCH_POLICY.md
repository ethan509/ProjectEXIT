# LottoSmash Branch Policy

이 문서는 LottoSmash 프로젝트의 Git branch 전략과 정책을 정의합니다.

## Branch Strategy: Git Flow

```
main (프로덕션, 항상 배포 가능)
  ↑ merge from develop
develop (개발 통합 브랜치)
  ↑ merge from feature/*
feature/* (기능 개발)
bugfix/* (버그 수정)
hotfix/* (긴급 수정 - main에서 분기)
```

## Branch 종류

### 1. main
- **목적**: 프로덕션 배포 가능 상태 유지
- **규칙**: develop에서만 머지
- **보호**: Branch Protection 활성화

### 2. develop
- **목적**: 개발 통합 브랜치
- **규칙**: feature/* 브랜치에서 머지
- **생성**: main에서 분기

### 3. feature/* (기능 개발)
- **명명규칙**: `feature/기능명`
- **생성**: develop에서 분기
- **머지**: develop으로 머지
- **삭제**: 삭제 금지
- **예시**:
  - `feature/user-login`
  - `feature/auto-migrate`
  - `feature/email-verification`

### 4. bugfix/* (버그 수정)
- **명명규칙**: `bugfix/버그명`
- **생성**: develop에서 분기
- **머지**: develop으로 머지
- **예시**:
  - `bugfix/login-error`
  - `bugfix/connection-timeout`

### 5. hotfix/* (긴급 수정)
- **명명규칙**: `hotfix/버그명`
- **생성**: main에서 분기
- **머지**: main과 develop 둘 다 머지
- **예시**:
  - `hotfix/security-patch`
  - `hotfix/data-loss`

## 머지 순서

```
feature/* → develop → main
```

1. feature 브랜치에서 작업 완료
2. develop으로 머지 (--no-ff)
3. develop에서 main으로 머지 (--no-ff)
4. feature 브랜치는 삭제하지 않음

## Commit Message Convention

**형식**: `<type>(<scope>): <subject>`

### Types
- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 추가/수정
- `style`: 코드 포맷 (로직 변경 X)
- `refactor`: 코드 개선
- `perf`: 성능 개선
- `test`: 테스트 추가/수정
- `chore`: 빌드, 의존성, CI/CD 설정 변경

### 예시
```
feat(auth): add email verification
fix(db): handle connection timeout
docs: update setup guide
chore: add GOTOOLCHAIN to Dockerfile
```

## 금지 사항

- main에서 직접 작업
- feature 브랜치 삭제
- develop 건너뛰고 main 직접 머지
- --no-ff 없이 머지

## 참고자료

- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
- [Conventional Commits](https://www.conventionalcommits.org/)
