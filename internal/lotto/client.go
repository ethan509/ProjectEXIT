package lotto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	dhlotteryBaseURL = "https://www.dhlottery.co.kr/common.do?method=getLottoNumber&drwNo=%d"
	dhlotteryHTMLURL = "https://www.dhlottery.co.kr/gameResult.do?method=byWin&drwNo=%d"
	clientTimeout    = 10 * time.Second
	requestDelay     = 100 * time.Millisecond
	maxRetries       = 3
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	// 리다이렉트를 자동으로 따라가도록 설정
	client := &http.Client{
		Timeout: clientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // 리다이렉트 따라가기
		},
	}
	return &Client{
		httpClient: client,
	}
}

// FetchDraw 특정 회차 당첨번호 조회
func (c *Client) FetchDraw(ctx context.Context, drawNo int) (*LottoDraw, error) {
	url := fmt.Sprintf(dhlotteryBaseURL, drawNo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 많은 헤더를 추가해서 실제 브라우저 요청처럼 보이기
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.dhlottery.co.kr/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Pragma", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draw %d: %w", drawNo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// HTML 응답(에러 페이지 등)인지 확인
	if len(bodyBytes) > 0 && bodyBytes[0] == '<' {
		return c.fetchDrawFromHTML(ctx, drawNo)
	}

	var apiResp DhlotteryResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return c.fetchDrawFromHTML(ctx, drawNo)
	}

	if apiResp.ReturnValue != "success" {
		return nil, fmt.Errorf("draw %d not found or not yet available", drawNo)
	}

	return apiResp.ToLottoDraw()
}

// fetchDrawFromHTML HTML 파싱을 통해 당첨번호 조회
func (c *Client) fetchDrawFromHTML(ctx context.Context, drawNo int) (*LottoDraw, error) {
	url := fmt.Sprintf(dhlotteryHTMLURL, drawNo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTML request: %w", err)
	}

	// 브라우저 헤더 추가
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.dhlottery.co.kr/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draw HTML %d: %w", drawNo, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTML body: %w", err)
	}
	body := string(bodyBytes)

	// 회차 검증
	// EUC-KR 인코딩 문제로 한글 '회' 매칭 실패 가능성 있음 -> 숫자와 태그만 매칭
	reDrawNo := regexp.MustCompile(`<strong>\s*([0-9]+)`)
	matches := reDrawNo.FindStringSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("draw number not found in HTML")
	}
	parsedDrawNo, _ := strconv.Atoi(matches[1])
	if parsedDrawNo != drawNo {
		return nil, fmt.Errorf("fetched draw number %d does not match requested %d", parsedDrawNo, drawNo)
	}

	// 날짜 파싱
	// (2002년 12월 07일 ...) 형식, 한글 깨짐 대비하여 숫자 위주로 매칭
	reDate := regexp.MustCompile(`\(([0-9]{4})[^0-9]+([0-9]{2})[^0-9]+([0-9]{2})`)
	dateMatches := reDate.FindStringSubmatch(body)
	var drawDate string
	if len(dateMatches) >= 4 {
		dateStr := fmt.Sprintf("%s-%s-%s", dateMatches[1], dateMatches[2], dateMatches[3])
		t, _ := time.Parse("2006-01-02", dateStr)
		drawDate = t.Format("2006.01.02")
	}

	// 번호 파싱
	reBalls := regexp.MustCompile(`<span class="ball_645 ball[0-9]+">([0-9]+)</span>`)
	ballMatches := reBalls.FindAllStringSubmatch(body, -1)
	if len(ballMatches) < 7 {
		return nil, fmt.Errorf("lotto numbers not found in HTML")
	}

	nums := make([]int, 7)
	for i := 0; i < 7; i++ {
		nums[i], _ = strconv.Atoi(ballMatches[i][1])
	}

	draw := &LottoDraw{
		DrawNo:   parsedDrawNo,
		DrawDate: drawDate,
		Num1:     nums[0], Num2: nums[1], Num3: nums[2],
		Num4: nums[3], Num5: nums[4], Num6: nums[5],
		BonusNum: nums[6],
	}

	// 당첨금 및 당첨자 정보 파싱
	c.parsePrizeInfo(body, draw)

	return draw, nil
}

// FetchDrawWithRetry 재시도 로직이 포함된 당첨번호 조회
func (c *Client) FetchDrawWithRetry(ctx context.Context, drawNo int) (*LottoDraw, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		draw, err := c.FetchDraw(ctx, drawNo)
		if err == nil {
			return draw, nil
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// FetchLatestDrawNo 최신 회차 번호 조회 (이진 검색)
func (c *Client) FetchLatestDrawNo(ctx context.Context) (int, error) {
	// 로또 1회차: 2002년 12월 7일
	// 대략적인 현재 회차 계산 (주당 1회)
	startDate := time.Date(2002, 12, 7, 0, 0, 0, 0, time.Local)
	weeks := int(time.Since(startDate).Hours() / 24 / 7)
	estimatedDraw := weeks + 1

	// 이진 검색으로 정확한 최신 회차 찾기
	low, high := estimatedDraw-10, estimatedDraw+10

	// high 값이 유효한지 확인하고 조정
	for {
		_, err := c.FetchDraw(ctx, high)
		if err != nil {
			high--
			if high <= low {
				break
			}
		} else {
			break
		}
	}

	// 최신 회차 찾기
	for low < high {
		mid := (low + high + 1) / 2
		_, err := c.FetchDraw(ctx, mid)
		if err != nil {
			high = mid - 1
		} else {
			low = mid
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(requestDelay):
		}
	}

	return low, nil
}

// FetchAllDraws 1회차부터 최신 회차까지 모든 당첨번호 조회
func (c *Client) FetchAllDraws(ctx context.Context) ([]LottoDraw, error) {
	// 최신 회차 번호 조회
	latestDrawNo, err := c.FetchLatestDrawNo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest draw number: %w", err)
	}

	// 1회차부터 최신 회차까지 조회
	return c.FetchDrawRange(ctx, 1, latestDrawNo)
}

// FetchDrawRange 범위 내 당첨번호 일괄 조회
func (c *Client) FetchDrawRange(ctx context.Context, from, to int) ([]LottoDraw, error) {
	var draws []LottoDraw

	for drawNo := from; drawNo <= to; drawNo++ {
		select {
		case <-ctx.Done():
			return draws, ctx.Err()
		case <-time.After(requestDelay):
		}

		draw, err := c.FetchDrawWithRetry(ctx, drawNo)
		if err != nil {
			return draws, fmt.Errorf("failed to fetch draw %d: %w", drawNo, err)
		}

		draws = append(draws, *draw)
	}

	return draws, nil
}

// parsePrizeInfo HTML에서 상금 및 당첨자 정보 파싱
func (c *Client) parsePrizeInfo(htmlBody string, draw *LottoDraw) {
	// 당첨금 정보 테이블 구조:
	// 1등 당첨금: <td class="tdn"> ... 원</td>
	// 당첨자: <td class="tdm"> ... 명</td>
	// 1게임당: <td class="tdn"> ... 원</td>
	// 이를 2~5등까지 반복

	// 상금 추출 (쉼표와 '원' 제거) - 총 당첨금과 1게임당 당첨금
	rePrice := regexp.MustCompile(`<td[^>]*>([0-9,]+)원</td>`)
	priceMatches := rePrice.FindAllStringSubmatch(htmlBody, -1)

	// 당첨게임수 (명) 추출
	reWinners := regexp.MustCompile(`<td[^>]*>([0-9,]+)명</td>`)
	winnersMatches := reWinners.FindAllStringSubmatch(htmlBody, -1)

	// 패턴: [총당첨금, 1게임당, 당첨게임수] × 5등급
	// priceMatches: 0=1등총, 1=1등개별, 2=2등총, 3=2등개별, ...
	// winnersMatches: 0=1등명, 1=2등명, 2=3등명, 3=4등명, 4=5등명

	// 1등
	if len(priceMatches) >= 2 {
		draw.FirstPrize = parsePrice(priceMatches[0][1])
		draw.FirstPerGame = parsePrice(priceMatches[1][1])
	}
	if len(winnersMatches) >= 1 {
		draw.FirstWinners = parseCount(winnersMatches[0][1])
	}

	// 2등
	if len(priceMatches) >= 4 {
		draw.SecondPrize = parsePrice(priceMatches[2][1])
		draw.SecondPerGame = parsePrice(priceMatches[3][1])
	}
	if len(winnersMatches) >= 2 {
		draw.SecondWinners = parseCount(winnersMatches[1][1])
	}

	// 3등
	if len(priceMatches) >= 6 {
		draw.ThirdPrize = parsePrice(priceMatches[4][1])
		draw.ThirdPerGame = parsePrice(priceMatches[5][1])
	}
	if len(winnersMatches) >= 3 {
		draw.ThirdWinners = parseCount(winnersMatches[2][1])
	}

	// 4등
	if len(priceMatches) >= 8 {
		draw.FourthPrize = parsePrice(priceMatches[6][1])
		draw.FourthPerGame = parsePrice(priceMatches[7][1])
	}
	if len(winnersMatches) >= 4 {
		draw.FourthWinners = parseCount(winnersMatches[3][1])
	}

	// 5등
	if len(priceMatches) >= 10 {
		draw.FifthPrize = parsePrice(priceMatches[8][1])
		draw.FifthPerGame = parsePrice(priceMatches[9][1])
	}
	if len(winnersMatches) >= 5 {
		draw.FifthWinners = parseCount(winnersMatches[4][1])
	}
}

// parsePrice 가격 문자열 파싱 (쉼표 제거)
func parsePrice(priceStr string) int64 {
	// 쉼표 제거
	cleaned := strings.ReplaceAll(priceStr, ",", "")
	price, _ := strconv.ParseInt(cleaned, 10, 64)
	return price
}

// parseCount 개수 문자열 파싱 (쉼표 제거)
func parseCount(countStr string) int {
	// 쉼표 제거
	cleaned := strings.ReplaceAll(countStr, ",", "")
	count, _ := strconv.Atoi(cleaned)
	return count
}
