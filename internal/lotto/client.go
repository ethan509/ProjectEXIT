package lotto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	dhlotteryBaseURL   = "https://www.dhlottery.co.kr/common.do?method=getLottoNumber&drwNo=%d"
	dhlotteryHTMLURL   = "https://www.dhlottery.co.kr/gameResult.do?method=byWin&drwNo=%d"
	dhlotteryAJAXURL   = "https://www.dhlottery.co.kr/lt645/selectPstLt645Info.do"
	dhlotteryResultURL = "https://www.dhlottery.co.kr/lt645/result"
	clientTimeout      = 10 * time.Second
	requestDelay       = 100 * time.Millisecond
	maxRetries         = 3
)

type Client struct {
	httpClient    *http.Client
	sessionInited bool
}

func NewClient() *Client {
	// 쿠키 자동 관리를 위한 cookiejar 생성
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Timeout: clientTimeout,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // 리다이렉트 따라가기
		},
	}
	return &Client{
		httpClient:    client,
		sessionInited: false,
	}
}

// setCommonHeaders 공통 브라우저 헤더 설정
func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
}

// initSession 세션 초기화 (쿠키 획득)
func (c *Client) initSession(ctx context.Context) error {
	if c.sessionInited {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dhlotteryResultURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create session request: %w", err)
	}

	c.setCommonHeaders(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to init session: %w", err)
	}
	defer resp.Body.Close()

	// 응답 본문 읽어서 쿠키 설정 완료
	io.ReadAll(resp.Body)
	c.sessionInited = true
	return nil
}

// FetchDraw 특정 회차 당첨번호 조회
func (c *Client) FetchDraw(ctx context.Context, drawNo int) (*LottoDraw, error) {
	// 세션 초기화 (쿠키 획득)
	if err := c.initSession(ctx); err != nil {
		// 세션 초기화 실패해도 계속 진행
	}

	url := fmt.Sprintf(dhlotteryBaseURL, drawNo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// AJAX 요청 헤더 설정
	c.setCommonHeaders(req)
	req.Header.Set("Referer", "https://www.dhlottery.co.kr/gameResult.do")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draw %d: %w", drawNo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fetchDrawFromAJAX(ctx, drawNo)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// HTML 응답(에러 페이지 등)인지 확인
	if len(bodyBytes) > 0 && bodyBytes[0] == '<' {
		return c.fetchDrawFromAJAX(ctx, drawNo)
	}

	var apiResp DhlotteryResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return c.fetchDrawFromAJAX(ctx, drawNo)
	}

	if apiResp.ReturnValue != "success" {
		return c.fetchDrawFromAJAX(ctx, drawNo)
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
	c.setCommonHeaders(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.dhlottery.co.kr/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

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

// FetchLatestDrawNo 최신 회차 번호 조회 (더 낮은 상한으로 조정)
func (c *Client) FetchLatestDrawNo(ctx context.Context) (int, error) {
	// 로또 1회차: 2002년 12월 7일
	// 대략적인 현재 회차 계산 (주당 1회)
	startDate := time.Date(2002, 12, 7, 0, 0, 0, 0, time.Local)
	weeks := int(time.Since(startDate).Hours() / 24 / 7)
	estimatedDraw := weeks + 1

	// 이진 검색으로 정확한 최신 회차 찾기
	// 더 보수적인 범위 설정 (추정값 -50 ~ +5)
	low, high := estimatedDraw-50, estimatedDraw+5

	// high 값이 유효한지 확인하고 조정 (아래로 감소)
	for high > low {
		_, err := c.FetchDraw(ctx, high)
		if err == nil {
			// high가 유효하면 break
			break
		}
		high--
	}

	// high가 low와 같으면 실패
	if high == low {
		return 0, fmt.Errorf("could not find valid draw number")
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

// FetchAjaxDebug AJAX 응답 디버깅용 함수 (각 파라미터별 응답 확인)
func (c *Client) FetchAjaxDebug(ctx context.Context, drawNo int) error {
	// 세션 초기화
	if err := c.initSession(ctx); err != nil {
		fmt.Printf("세션 초기화 실패: %v\n", err)
	}

	// 여러 파라미터 조합 시도
	paramCombinations := []string{
		fmt.Sprintf("srchLtEpsd=%d", drawNo),
	}

	for i, params := range paramCombinations {
		url := fmt.Sprintf("%s?%s", dhlotteryAJAXURL, params)

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		c.setCommonHeaders(req)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Referer", dhlotteryResultURL)
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			fmt.Printf("[%d] 요청 실패: %v\n", i, err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("[%d] URL: %s\n", i, url)
		fmt.Printf("    Status: %d\n", resp.StatusCode)

		var ajaxResp map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &ajaxResp); err != nil {
			fmt.Printf("    JSON 파싱 실패: %v\n", err)
			continue
		}

		// 응답에서 회차 정보 확인
		if data, ok := ajaxResp["data"].(map[string]interface{}); ok {
			if list, ok := data["list"].([]interface{}); ok && len(list) > 0 {
				if item, ok := list[0].(map[string]interface{}); ok {
					ltEpsd, _ := item["ltEpsd"]
					fmt.Printf("    응답 회차: %v\n", ltEpsd)
				}
			} else {
				fmt.Printf("    list 없음 또는 비어있음\n")
			}
		} else {
			fmt.Printf("    data 없음\n")
		}
	}

	return nil
}

// fetchDrawFromAJAX AJAX 엔드포인트에서 당첨번호 조회 (드롭다운 선택 시 사용되는 API)
func (c *Client) fetchDrawFromAJAX(ctx context.Context, drawNo int) (*LottoDraw, error) {
	// 세션 초기화 (쿠키 획득)
	if err := c.initSession(ctx); err != nil {
		// 세션 초기화 실패해도 계속 진행
	}

	// 파라미터 조합 시도 (srchLtEpsd가 회차 검색 파라미터)
	paramCombinations := []string{
		fmt.Sprintf("srchLtEpsd=%d", drawNo),
	}

	var lastErr error
	for _, params := range paramCombinations {
		url := fmt.Sprintf("%s?%s", dhlotteryAJAXURL, params)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// AJAX 요청 헤더 (새로운 브라우저 패턴)
		c.setCommonHeaders(req)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Referer", dhlotteryResultURL)
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// AJAX 응답 파싱
		var ajaxResp map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &ajaxResp); err != nil {
			lastErr = err
			continue
		}

		// 응답에서 데이터 추출
		draw := c.parseAJAXResponse(ajaxResp, drawNo)
		if draw != nil {
			return draw, nil
		}

		lastErr = fmt.Errorf("no valid data in AJAX response")
	}

	// AJAX 실패 시 HTML 파싱으로 폴백
	draw, err := c.fetchDrawFromHTML(ctx, drawNo)
	if err != nil {
		return nil, fmt.Errorf("AJAX failed (%v), HTML fallback also failed: %w", lastErr, err)
	}
	return draw, nil
}

// parseAJAXResponse AJAX 응답을 파싱하여 LottoDraw 객체 생성
func (c *Client) parseAJAXResponse(resp map[string]interface{}, drawNo int) *LottoDraw {
	// 응답 구조 확인
	dataInterface, ok := resp["data"]
	if !ok {
		return nil
	}

	dataMap, ok := dataInterface.(map[string]interface{})
	if !ok {
		return nil
	}

	listInterface, ok := dataMap["list"].([]interface{})
	if !ok || len(listInterface) == 0 {
		return nil
	}

	item, ok := listInterface[0].(map[string]interface{})
	if !ok {
		return nil
	}

	draw := &LottoDraw{DrawNo: drawNo}

	// 날짜 파싱 (ltRflYmd: "20260124" → "2026.01.24")
	if ltRflYmd, ok := item["ltRflYmd"].(string); ok && len(ltRflYmd) == 8 {
		draw.DrawDate = fmt.Sprintf("%s.%s.%s", ltRflYmd[:4], ltRflYmd[4:6], ltRflYmd[6:8])
	}

	// 당첨번호 (tm1WnNo ~ tm6WnNo + bnsWnNo)
	nums := []int{}
	for i := 1; i <= 6; i++ {
		key := fmt.Sprintf("tm%dWnNo", i)
		if num, ok := item[key].(float64); ok {
			nums = append(nums, int(num))
		}
	}
	if len(nums) == 6 {
		draw.Num1 = nums[0]
		draw.Num2 = nums[1]
		draw.Num3 = nums[2]
		draw.Num4 = nums[3]
		draw.Num5 = nums[4]
		draw.Num6 = nums[5]
	}

	if bnsWnNo, ok := item["bnsWnNo"].(float64); ok {
		draw.BonusNum = int(bnsWnNo)
	}

	// 상금 정보 파싱
	rankMap := map[string]struct {
		sumField    *int64
		perField    *int64
		winnerField *int
	}{
		"rnk1": {&draw.FirstPrize, &draw.FirstPerGame, &draw.FirstWinners},
		"rnk2": {&draw.SecondPrize, &draw.SecondPerGame, &draw.SecondWinners},
		"rnk3": {&draw.ThirdPrize, &draw.ThirdPerGame, &draw.ThirdWinners},
		"rnk4": {&draw.FourthPrize, &draw.FourthPerGame, &draw.FourthWinners},
		"rnk5": {&draw.FifthPrize, &draw.FifthPerGame, &draw.FifthWinners},
	}

	for rank, fields := range rankMap {
		if v, ok := item[rank+"SumWnAmt"].(float64); ok {
			*fields.sumField = int64(v)
		}
		if v, ok := item[rank+"WnAmt"].(float64); ok {
			*fields.perField = int64(v)
		}
		if v, ok := item[rank+"WnNope"].(float64); ok {
			*fields.winnerField = int(v)
		}
	}

	// 최소한 당첨번호와 날짜는 있어야 함
	if draw.DrawNo > 0 && draw.Num1 > 0 {
		return draw
	}

	return nil
}
