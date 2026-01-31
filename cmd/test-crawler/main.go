package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/example/LottoSmash/internal/lotto"
)

func main() {
	// 커맨드 라인 플래그 설정
	drawNoPtr := flag.Int("no", 0, "조회할 회차 번호 (0일 경우 최신 회차 조회)")
	saveHTMLPtr := flag.Bool("save", true, "HTML 원본 파일 저장 여부")
	flag.Parse()

	drawNo := *drawNoPtr
	// 위치 인자로 회차 번호가 전달된 경우 처리 (예: go run main.go 1208)
	if drawNo == 0 && flag.NArg() > 0 {
		if v, err := strconv.Atoi(flag.Arg(0)); err == nil {
			drawNo = v
		}
	}

	// 클라이언트 초기화
	client := lotto.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 로거 설정 (표준 출력)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if drawNo > 0 {
		testSpecificDraw(ctx, client, drawNo, *saveHTMLPtr)
	} else {
		testLatestDraw(ctx, client, *saveHTMLPtr)
	}
}

func testSpecificDraw(ctx context.Context, client *lotto.Client, drawNo int, saveHTML bool) {
	log.Printf("=== %d회차 당첨번호 조회 테스트 시작 ===", drawNo)

	// AJAX 디버그
	log.Printf("\n📡 AJAX 파라미터 디버그:")
	client.FetchAjaxDebug(ctx, drawNo)

	if saveHTML {
		saveHTMLFile(ctx, client, drawNo)
	}

	start := time.Now()
	draw, err := client.FetchDraw(ctx, drawNo)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ 조회 실패 (소요시간: %v): %v", duration, err)
		return
	}

	log.Printf("✅ 조회 성공 (소요시간: %v)", duration)
	printDraw(draw)
}

func testLatestDraw(ctx context.Context, client *lotto.Client, saveHTML bool) {
	log.Println("=== 최신 회차 번호 및 데이터 조회 테스트 시작 ===")

	// 1. 최신 회차 번호 찾기
	log.Println("1. 최신 회차 번호 검색 중...")
	start := time.Now()
	latestNo, err := client.FetchLatestDrawNo(ctx)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ 최신 회차 번호 조회 실패 (소요시간: %v): %v", duration, err)
		return
	}
	log.Printf("✅ 최신 회차 번호 발견: %d회 (소요시간: %v)", latestNo, duration)

	if saveHTML {
		saveHTMLFile(ctx, client, latestNo)
	}

	// 2. 해당 회차 데이터 조회
	log.Printf("2. %d회차 상세 데이터 조회 중...", latestNo)
	start = time.Now()
	draw, err := client.FetchDraw(ctx, latestNo)
	duration = time.Since(start)

	if err != nil {
		log.Printf("❌ 상세 데이터 조회 실패 (소요시간: %v): %v", duration, err)
		return
	}

	log.Printf("✅ 상세 데이터 조회 성공 (소요시간: %v)", duration)
	printDraw(draw)
}

func saveHTMLFile(ctx context.Context, client *lotto.Client, drawNo int) {
	// TODO: HTML 저장 기능은 나중에 구현
	// log.Printf("💾 %d회차 HTML 원본 다운로드 중...", drawNo)
}

func printDraw(d *lotto.LottoDraw) {
	fmt.Println("---------------------------------------------------")
	fmt.Printf("회차: %d\n", d.DrawNo)
	fmt.Printf("날짜: %s\n", d.DrawDate)
	fmt.Printf("번호: %d, %d, %d, %d, %d, %d + %d\n", d.Num1, d.Num2, d.Num3, d.Num4, d.Num5, d.Num6, d.BonusNum)
	fmt.Printf("1등 당첨금: 총 %d원, (1인당) %d원 (%d명)\n", d.FirstPrize, d.FirstPerGame, d.FirstWinners)
	fmt.Println("---------------------------------------------------")
}
