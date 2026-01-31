.PHONY: build run test test-crawler docker-up docker-down docker-logs docker-clean

# 로컬 빌드 및 실행
build:
	go build -o lottosmash cmd/server/main.go

run: build
	./lottosmash

test:
	go test -v ./...

# 크롤러 테스트
test-crawler:
	go run cmd/test-crawler/main.go

# Docker 관련 명령어
docker-up:
	docker-compose -f docker/docker-compose.yml up -d --build

docker-down:
	docker-compose -f docker/docker-compose.yml down

docker-logs:
	docker-compose -f docker/docker-compose.yml logs -f

docker-clean:
	docker-compose -f docker/docker-compose.yml down -v

docker-ps:
	docker-compose -f docker/docker-compose.yml ps