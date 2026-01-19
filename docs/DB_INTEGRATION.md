# 데이터베이스 통합 가이드

ProjectEXIT 템플릿에 데이터베이스를 통합하는 방법을 안내합니다.

> **참고:** 이 템플릿은 DB 연결 코드를 포함하지 않습니다. 필요에 따라 PostgreSQL, MySQL, MongoDB 등을 자유롭게 선택할 수 있습니다.

## 목차

- [PostgreSQL 통합](#postgresql-통합)
- [MySQL 통합](#mysql-통합)
- [MongoDB 통합](#mongodb-통합)
- [DB 마이그레이션](#db-마이그레이션)
- [예제: CRUD 구현](#예제-crud-구현)

---

## PostgreSQL 통합

### 1. 의존성 추가

```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

### 2. DB 패키지 생성

`internal/database/postgres.go`:

```go
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewPostgresPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 커넥션 풀 설정
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// 연결 테스트
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
```

### 3. main.go에서 초기화

```go
import (
	"github.com/example/ProjectEXIT/internal/database"
	"os"
)

func main() {
	// ... 기존 코드 ...

	// DB 초기화
	dbConfig := database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}

	dbPool, err := database.NewPostgresPool(ctx, dbConfig)
	if err != nil {
		lg.Criticalf("failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	lg.Infof("database connected: %s", dbConfig.Host)

	// Dependencies에 DB 추가
	deps := server.Dependencies{
		ConfigMgr: cfgMgr,
		Logger:    lg,
		Pools:     pools,
		DB:        dbPool,  // 추가
	}

	// ... 나머지 코드 ...
}
```

### 4. Dependencies 구조체 수정

`internal/server/router.go`:

```go
import "github.com/jackc/pgx/v5/pgxpool"

type Dependencies struct {
	ConfigMgr config.Configger
	Logger    *logger.Logger
	Pools     *worker.Pools
	DB        *pgxpool.Pool  // 추가
}
```

### 5. /readyz에 DB 체크 추가

`internal/server/router.go`:

```go
r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
	// DB 연결 확인
	if deps.DB != nil {
		if err := deps.DB.Ping(r.Context()); err != nil {
			response.JSON(w, r, http.StatusServiceUnavailable,
				"NOT_READY", "database not ready", nil)
			return
		}
	}
	response.JSON(w, r, http.StatusOK, "READY", "ready", nil)
})
```

---

## MySQL 통합

### 1. 의존성 추가

```bash
go get github.com/go-sql-driver/mysql
go get github.com/jmoiron/sqlx
```

### 2. DB 패키지 생성

`internal/database/mysql.go`:

```go
package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewMySQLPool(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	// 커넥션 풀 설정
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 연결 테스트
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return db, nil
}
```

### 3. docker-compose.yml 수정

```yaml
services:
  db:
    image: mysql:8.0
    container_name: xxxdongxxx-db
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: xxxdongxxx_db
      MYSQL_USER: devuser
      MYSQL_PASSWORD: devpassword
    volumes:
      - db_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
```

---

## MongoDB 통합

### 1. 의존성 추가

```bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
```

### 2. DB 패키지 생성

`internal/database/mongodb.go`:

```go
package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(ctx context.Context, cfg Config) (*mongo.Client, error) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port,
	)

	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(25).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	// 연결 테스트
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return client, nil
}
```

### 3. docker-compose.yml 수정

```yaml
services:
  db:
    image: mongo:7
    container_name: xxxdongxxx-db
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: devuser
      MONGO_INITDB_ROOT_PASSWORD: devpassword
      MONGO_INITDB_DATABASE: xxxdongxxx_db
    volumes:
      - db_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5
```

---

## DB 마이그레이션

### golang-migrate 사용

#### 1. 설치

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

#### 2. 마이그레이션 파일 생성

```bash
mkdir -p migrations

migrate create -ext sql -dir migrations -seq create_users_table
```

`migrations/000001_create_users_table.up.sql`:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

`migrations/000001_create_users_table.down.sql`:

```sql
DROP TABLE IF EXISTS users;
```

#### 3. 마이그레이션 실행

```bash
# 적용
migrate -path migrations -database "postgres://devuser:devpassword@localhost:5432/xxxdongxxx_db?sslmode=disable" up

# 롤백
migrate -path migrations -database "postgres://devuser:devpassword@localhost:5432/xxxdongxxx_db?sslmode=disable" down
```

---

## 예제: CRUD 구현

### 1. Repository 패턴

`internal/repository/user.go`:

```go
package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, username, email string) (*User, error) {
	query := `
		INSERT INTO users (username, email)
		VALUES ($1, $2)
		RETURNING id, username, email, created_at, updated_at
	`

	var user User
	err := r.db.QueryRow(ctx, query, username, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, id int64, username, email string) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, username, email, id)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
```

### 2. 핸들러 추가

`internal/server/handlers.go`:

```go
import "github.com/example/ProjectEXIT/internal/repository"

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func CreateUserHandler(deps Dependencies) http.HandlerFunc {
	userRepo := repository.NewUserRepository(deps.DB)

	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.ErrorJSON(w, r, &response.AppError{
				Code:       "BAD_REQUEST",
				Message:    "invalid json",
				HTTPStatus: http.StatusBadRequest,
				Err:        err,
			})
			return
		}

		user, err := userRepo.Create(r.Context(), req.Username, req.Email)
		if err != nil {
			response.ErrorJSON(w, r, &response.AppError{
				Code:       "INTERNAL_ERROR",
				Message:    "failed to create user",
				HTTPStatus: http.StatusInternalServerError,
				Err:        err,
			})
			return
		}

		response.JSON(w, r, http.StatusCreated, "CREATED", "user created", user)
	}
}
```

### 3. 라우터 추가

`internal/server/router.go`:

```go
// RESTful API
r.Route("/api/v1/users", func(r chi.Router) {
	r.Post("/", CreateUserHandler(deps))
	r.Get("/{id}", GetUserHandler(deps))
	r.Put("/{id}", UpdateUserHandler(deps))
	r.Delete("/{id}", DeleteUserHandler(deps))
	r.Get("/", ListUsersHandler(deps))
})
```

---

## 트랜잭션 처리

### PostgreSQL 트랜잭션 예제

```go
func (r *UserRepository) CreateWithProfile(ctx context.Context, username, email, bio string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. 사용자 생성
	var userID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id",
		username, email,
	).Scan(&userID)
	if err != nil {
		return err
	}

	// 2. 프로필 생성
	_, err = tx.Exec(ctx,
		"INSERT INTO profiles (user_id, bio) VALUES ($1, $2)",
		userID, bio,
	)
	if err != nil {
		return err
	}

	// 커밋
	return tx.Commit(ctx)
}
```

---

## 연결 풀 튜닝

### PostgreSQL (pgxpool)

```go
poolConfig.MaxConns = 25                // 최대 연결 수
poolConfig.MinConns = 5                 // 최소 연결 수
poolConfig.MaxConnLifetime = time.Hour  // 연결 최대 수명
poolConfig.MaxConnIdleTime = 30 * time.Minute  // 유휴 연결 타임아웃
poolConfig.HealthCheckPeriod = time.Minute     // 헬스체크 주기
```

### 권장 설정

- **API 서버**: MaxConns = CPU 코어 * 2 ~ 4
- **백그라운드 워커**: MaxConns = 워커 수 + 2
- **마이크로서비스**: MinConns = 2~5, MaxConns = 10~25

---

## 환경 변수 설정

`.env` 파일 (개발용):

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=devuser
DB_PASSWORD=devpassword
DB_NAME=xxxdongxxx_db
```

Docker Compose는 이미 설정되어 있습니다 (`docker/docker-compose.yml`).

---

## 문제 해결

### 연결 실패

```bash
# PostgreSQL 컨테이너 로그 확인
docker logs xxxdongxxx-db

# DB 직접 접속 테스트
docker exec -it xxxdongxxx-db psql -U devuser -d xxxdongxxx_db
```

### 성능 최적화

1. **인덱스 추가**: 자주 조회하는 컬럼에 인덱스 생성
2. **쿼리 최적화**: EXPLAIN ANALYZE로 쿼리 분석
3. **커넥션 풀 튜닝**: 애플리케이션 부하에 맞게 조정
4. **준비된 구문**: 반복 쿼리는 Prepared Statement 사용

---

## 참고 자료

- [pgx 문서](https://github.com/jackc/pgx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [sqlx 문서](https://github.com/jmoiron/sqlx)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
