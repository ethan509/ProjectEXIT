.PHONY: build run test docker-up docker-down docker-logs docker-clean

# 로컬 빌드 및 실행
build:
	go build -o lottosmash cmd/server/main.go

run: build
	./lottosmash

test:
	go test -v ./...

# Docker 관련 명령어
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v

docker-ps:
	docker-compose ps