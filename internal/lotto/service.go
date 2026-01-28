package lotto

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/example/LottoSmash/internal/logger"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

type Service struct {
	repo     *Repository
	client   *Client
	analyzer *Analyzer
	log      *logger.Logger
	docsPath string
}

func NewService(repo *Repository, client *Client, analyzer *Analyzer, log *logger.Logger) *Service {
	return &Service{
		repo:     repo,
		client:   client,
		analyzer: analyzer,
		log:      log,
	}
}

// InitializeDraws 서버 시작 시 당첨번호 초기화
func (s *Service) InitializeDraws(ctx context.Context, docsPath string) error {
	s.docsPath = docsPath
	s.log.Infof("initializing lotto draws...")

	// 1. lotto_draw 테이블에 최신정보(현재날짜의 가장 최근 토요일 기준)까지 적재되어 있는지 확인
	expectedDrawNo := s.calculateExpectedDrawNo()

	latestInDB, err := s.repo.GetLatestDrawNo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest draw from DB: %w", err)
	}

	s.log.Infof("Draw status - DB: %d, Expected: %d", latestInDB, expectedDrawNo)

	if latestInDB >= expectedDrawNo {
		s.log.Infof("DB is up to date.")
		// 초기화 시에도 분석 실행
		return s.RunAnalysis(ctx)
	}

	// 2. 없으면 csv 파일 확인
	s.log.Infof("DB is outdated. Checking CSV files in %s...", docsPath)

	// 디버깅: docs 폴더 파일 목록 출력 (파일이 존재하는지 확인)
	entries, err := os.ReadDir(docsPath)
	if err == nil {
		s.log.Infof("Found %d files in %s:", len(entries), docsPath)
		for _, e := range entries {
			s.log.Infof(" - %s", e.Name())
		}
	} else {
		s.log.Errorf("Failed to read docs dir: %v", err)
	}

	csvLoaded := false
	// 파일 검색 로직 개선 (대소문자 구분 없이 .csv 확인)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
				continue
			}
			csvPath := filepath.Join(docsPath, entry.Name())
			s.log.Infof("Found candidate CSV: %s", entry.Name())

			if err := s.loadFromCSV(ctx, csvPath); err == nil {
				s.log.Infof("Successfully loaded draws from csv: %s", filepath.Base(csvPath))
				csvLoaded = true
				break
			} else {
				s.log.Infof("Failed to load CSV %s: %v", filepath.Base(csvPath), err)
				s.log.Errorf("Failed to load CSV %s: %v", filepath.Base(csvPath), err)
			}
		}
	} else {
		s.log.Infof("No CSV files found in %s", docsPath)
	}

	if csvLoaded {
		latestInDB, err = s.repo.GetLatestDrawNo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest draw from DB after CSV load: %w", err)
		}
		if latestInDB >= expectedDrawNo {
			s.log.Infof("DB is up to date after CSV load.")
			return nil
		}
	}

	// 3. csv 파일에도 없으면 사이트에서 정보 가져오고 csv 파일 업데이트 그리고 db update
	s.log.Infof("DB still outdated (DB: %d, Expected: %d). Fetching from external source...", latestInDB, expectedDrawNo)

	s.log.Infof("Attempting to fetch missing draws (from %d to %d) from external site...", latestInDB+1, expectedDrawNo)

	// 누락분 가져오기 (DB 저장은 fetchAndSaveDraws 내부에서 수행)
	if err := s.fetchAndSaveDraws(ctx, latestInDB+1, expectedDrawNo); err != nil {
		s.log.Infof("Failed to fetch draws from external site: %v", err)
		s.log.Errorf("failed to fetch draws from API: %v", err)
		// API 실패 시 스크래핑 시도 (마지막 1회차인 경우)
		if expectedDrawNo-latestInDB == 1 {
			s.log.Infof("trying to scrape latest draw %d...", expectedDrawNo)
			scraped, errScrape := s.scrapeLatestDraw(ctx)
			if errScrape == nil && scraped.DrawNo == expectedDrawNo {
				if errSave := s.repo.SaveDraw(ctx, scraped); errSave == nil {
					s.log.Infof("saved scraped draw %d", scraped.DrawNo)
				}
			}
		}
	} else {
		s.log.Infof("Successfully fetched and saved draws from external site.")
	}

	// 최종 확인: 여전히 최신 데이터가 없으면 에러 반환
	finalLatest, err := s.repo.GetLatestDrawNo(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify final DB state: %w", err)
	}

	if finalLatest < expectedDrawNo {
		return fmt.Errorf("failed to initialize draws: could not fetch up to draw %d (current: %d)", expectedDrawNo, finalLatest)
	}

	// CSV 파일 업데이트
	if err := s.updateCSVFile(ctx); err != nil {
		return err
	}

	// 초기 데이터 로드 후 분석 실행
	return s.RunAnalysis(ctx)
}

func (s *Service) calculateExpectedDrawNo() int {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		loc = time.FixedZone("KST", 9*60*60)
	}
	now := time.Now().In(loc)

	// 1회차: 2002-12-07 20:00 KST (토요일)
	// 추첨 기준: 매주 토요일 21:00 이후에는 해당 회차가 존재해야 함
	baseDate := time.Date(2002, 12, 7, 21, 0, 0, 0, loc)

	if now.Before(baseDate) {
		return 1
	}

	diff := now.Sub(baseDate)
	weeks := int(diff.Hours() / (24 * 7))
	return 1 + weeks
}

// fetchAndSaveDraws 범위 내 당첨번호 수집 및 저장
func (s *Service) fetchAndSaveDraws(ctx context.Context, from, to int) error {
	total := to - from + 1
	s.log.Infof("fetching %d draws...", total)

	draws, err := s.client.FetchDrawRange(ctx, from, to)
	if err != nil {
		return err
	}

	if err := s.repo.SaveDrawBatch(ctx, draws); err != nil {
		return err
	}

	s.log.Infof("saved %d draws successfully", len(draws))

	// 분석 실행
	return s.RunAnalysis(ctx)
}

// loadFromCSV CSV 파일 파싱 및 저장
func (s *Service) loadFromCSV(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read csv file: %w", err)
	}

	// BOM 제거 (UTF-8 with BOM)
	if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
		content = content[3:]
	}

	var reader io.Reader
	// UTF-8 유효성 검사
	if utf8.Valid(content) {
		reader = bytes.NewReader(content)
	} else {
		// 유효하지 않은 UTF-8이면 EUC-KR로 간주하고 변환
		s.log.Infof("Detected non-UTF-8 encoding for %s, converting from EUC-KR", filepath.Base(filePath))
		reader = transform.NewReader(bytes.NewReader(content), korean.EUCKR.NewDecoder())
	}

	draws, err := s.parseCSV(reader)
	if err != nil {
		return err
	}

	s.log.Infof("parsed %d draws from csv", len(draws))
	return s.repo.SaveDrawBatch(ctx, draws)
}

// parseCSV CSV 데이터 파싱
func (s *Service) parseCSV(r io.Reader) ([]LottoDraw, error) {
	csvReader := csv.NewReader(r)
	csvReader.LazyQuotes = true       // 따옴표 처리 완화
	csvReader.TrimLeadingSpace = true // 앞 공백 제거
	csvReader.FieldsPerRecord = -1    // 필드 개수 가변 허용

	// 헤더 스킵
	if _, err := csvReader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var draws []LottoDraw
	line := 1 // 헤더가 1번째 줄

	// 숫자 정제 헬퍼 함수: 숫자만 추출 (EUC-KR 등 인코딩 이슈 및 특수문자 제거)
	cleanNum := func(s string) string {
		var sb strings.Builder
		for _, r := range s {
			if r >= '0' && r <= '9' {
				sb.WriteRune(r)
			}
		}
		return sb.String()
	}

	for {
		line++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			s.log.Errorf("CSV parse error at line %d: %v", line, err)
			continue
		}

		if len(record) < 12 {
			s.log.Errorf("Skipping line %d: insufficient columns (got %d, want 12+)", line, len(record))
			continue
		}

		draw := LottoDraw{}

		// 회차 정보
		if draw.DrawNo, err = strconv.Atoi(cleanNum(record[1])); err != nil {
			s.log.Errorf("Skipping line %d: invalid draw number '%s' ->(draw number:%d)", line, record[1], draw.DrawNo)
			continue
		}

		// 날짜 계산 (1회차: 2002-12-07 토요일)
		// CSV에 날짜가 없으므로 회차 번호를 기반으로 정확히 계산
		// 매주 토요일에 추첨되므로 7일 단위로 계산
		baseDate := time.Date(2002, 12, 7, 0, 0, 0, 0, time.UTC)
		draw.DrawDate = baseDate.AddDate(0, 0, (draw.DrawNo-1)*7).Format("2006.01.02")
		s.log.Debugf("line %d: draw %d calculated date: %s", line, draw.DrawNo, draw.DrawDate)

		// 로또 당첨번호 6개 + 보너스번호
		var parseErr error
		parseInt := func(s string) int {
			v, err := strconv.Atoi(cleanNum(s))
			if err != nil {
				parseErr = err
			}
			return v
		}
		draw.Num1 = parseInt(record[2])
		draw.Num2 = parseInt(record[3])
		draw.Num3 = parseInt(record[4])
		draw.Num4 = parseInt(record[5])
		draw.Num5 = parseInt(record[6])
		draw.Num6 = parseInt(record[7])
		draw.BonusNum = parseInt(record[8])
		if parseErr != nil {
			s.log.Errorf("Skipping line %d: invalid lotto numbers", line)
			continue
		}

		// 순위(Index 9)는 "1등"이므로 저장하지 않음
		// 당첨자수 (Index 10): "6 명" -> "명" 제거
		if draw.FirstWinners, err = strconv.Atoi(cleanNum(record[10])); err != nil {
			s.log.Errorf("Skipping line %d: invalid winners '%s'", line, record[10])
			continue
		}

		// 당첨금액 (Index 11): "5,001,713,625 원" -> 콤마(,)와 "원" 제거
		if draw.FirstPrize, err = strconv.ParseInt(cleanNum(record[11]), 10, 64); err != nil {
			s.log.Errorf("Skipping line %d: invalid prize '%s'", line, record[11])
			continue
		}

		draws = append(draws, draw)
	}

	return draws, nil
}

// scrapeLatestDraw 웹사이트에서 최신 당첨번호 스크래핑
func (s *Service) scrapeLatestDraw(ctx context.Context) (*LottoDraw, error) {
	url := "https://www.dhlottery.co.kr/gameResult.do?method=byWin"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 브라우저 헤더 추가 (차단 방지)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.dhlottery.co.kr/")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	// 회차 파싱: <strong>1154회</strong>
	reDrawNo := regexp.MustCompile(`<strong>\s*([0-9]+)`)
	matches := reDrawNo.FindStringSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("draw number not found")
	}
	drawNo, _ := strconv.Atoi(matches[1])

	// 날짜 파싱: (2025년 01월 11일 추첨)
	reDate := regexp.MustCompile(`\(([0-9]{4}).*?([0-9]{2}).*?([0-9]{2})`)
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
		return nil, fmt.Errorf("lotto numbers not found")
	}

	nums := make([]int, 7)
	for i := 0; i < 7; i++ {
		nums[i], _ = strconv.Atoi(ballMatches[i][1])
	}

	return &LottoDraw{
		DrawNo:   drawNo,
		DrawDate: drawDate,
		Num1:     nums[0], Num2: nums[1], Num3: nums[2],
		Num4: nums[3], Num5: nums[4], Num6: nums[5],
		BonusNum: nums[6],
	}, nil
}

// updateCSVFile DB의 모든 당첨번호를 CSV 파일로 저장하고 파일명을 최신화함
func (s *Service) updateCSVFile(ctx context.Context) error {
	total, err := s.repo.GetTotalDrawCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	if total == 0 {
		return nil
	}

	// 모든 데이터 조회
	draws, err := s.repo.GetDraws(ctx, int(total), 0)
	if err != nil {
		return fmt.Errorf("failed to get all draws: %w", err)
	}

	// 최신 회차 찾기
	maxNo := 0
	for _, d := range draws {
		if d.DrawNo > maxNo {
			maxNo = d.DrawNo
		}
	}

	// 새 파일명 생성
	newFileName := fmt.Sprintf("로또 회차별 당첨번호_1_%d.csv", maxNo)
	newFilePath := filepath.Join(s.docsPath, newFileName)

	// 파일 생성 및 쓰기
	f, err := os.Create(newFilePath)
	if err != nil {
		return fmt.Errorf("failed to create csv file: %w", err)
	}
	defer f.Close()

	// 엑셀 호환을 위한 BOM 추가
	if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}

	w := csv.NewWriter(f)

	// 헤더 작성
	header := []string{"No", "회차", "번호1", "번호2", "번호3", "번호4", "번호5", "번호6", "보너스", "순위", "1등당첨자수", "1등당첨금"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, d := range draws {
		record := []string{
			strconv.Itoa(d.DrawNo), // No
			strconv.Itoa(d.DrawNo), // 회차
			strconv.Itoa(d.Num1),
			strconv.Itoa(d.Num2),
			strconv.Itoa(d.Num3),
			strconv.Itoa(d.Num4),
			strconv.Itoa(d.Num5),
			strconv.Itoa(d.Num6),
			strconv.Itoa(d.BonusNum),
			"1등", // 순위
			strconv.Itoa(d.FirstWinners),
			strconv.FormatInt(d.FirstPrize, 10),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	// 기존 파일 정리 (다른 이름의 CSV 삭제)
	matches, _ := filepath.Glob(filepath.Join(s.docsPath, "로또 회차별 당첨번호*.csv"))
	for _, m := range matches {
		if m != newFilePath {
			if err := os.Remove(m); err != nil {
				s.log.Errorf("failed to remove old csv %s: %v", m, err)
			} else {
				s.log.Infof("removed old csv: %s", filepath.Base(m))
			}
		}
	}

	s.log.Infof("updated csv file: %s", newFileName)
	return nil
}

// UpdateExcelFile 모든 당첨번호를 엑셀 파일로 저장
// TODO: 향후 엑셀 파일 저장 기능 추가
// func (s *Service) UpdateExcelFile(ctx context.Context, docsPath string) error {
// 	// DB에서 모든 당첨번호 조회
// 	draws, err := s.repo.GetAllDraws(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to get all draws: %w", err)
// 	}

// 	if len(draws) == 0 {
// 		return fmt.Errorf("no draws to save")
// 	}

// 	// 최소/최대 회차 찾기
// 	minNo := draws[0].DrawNo
// 	maxNo := draws[0].DrawNo
// 	for _, draw := range draws {
// 		if draw.DrawNo < minNo {
// 			minNo = draw.DrawNo
// 		}
// 		if draw.DrawNo > maxNo {
// 			maxNo = draw.DrawNo
// 		}
// 	}

// 	// 엑셀 파일로 저장
// 	excelLoader := NewExcelLoader(docsPath)
// 	return excelLoader.SaveDrawsToExcel(draws, minNo, maxNo)
// }

func (s *Service) FetchNewDraw(ctx context.Context) error {
	s.log.Infof("fetching new lotto draw...")

	// DB에서 최신 회차 확인
	latestInDB, err := s.repo.GetLatestDrawNo(ctx)
	if err != nil {
		return err
	}

	// 다음 회차 조회 시도
	nextDrawNo := latestInDB + 1
	draw, err := s.client.FetchDrawWithRetry(ctx, nextDrawNo)
	if err != nil {
		s.log.Infof("draw %d not available yet: %v", nextDrawNo, err)
		return nil // 아직 발표 안됨 - 에러가 아님
	}

	// 저장
	if err := s.repo.SaveDraw(ctx, draw); err != nil {
		return err
	}

	s.log.Infof("saved new draw %d successfully", nextDrawNo)

	if err := s.updateCSVFile(ctx); err != nil {
		s.log.Errorf("failed to update csv file: %v", err)
	}

	return nil
}

// RunAnalysis 분석 실행
func (s *Service) RunAnalysis(ctx context.Context) error {
	s.log.Infof("RunAnalysis called - starting lotto analysis...")

	if s.analyzer == nil {
		s.log.Errorf("analyzer is nil - cannot run analysis")
		return fmt.Errorf("analyzer is nil")
	}

	s.log.Infof("about to call analyzer.RunFullAnalysis()")
	if err := s.analyzer.RunFullAnalysis(ctx); err != nil {
		s.log.Errorf("failed to run analysis: %v", err)
		return err
	}

	s.log.Infof("lotto analysis completed successfully")
	return nil
}

// GetDraws 당첨번호 목록 조회
func (s *Service) GetDraws(ctx context.Context, limit, offset int) (*DrawListResponse, error) {
	draws, err := s.repo.GetDraws(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.GetTotalDrawCount(ctx)
	if err != nil {
		return nil, err
	}

	latest, err := s.repo.GetLatestDrawNo(ctx)
	if err != nil {
		return nil, err
	}

	return &DrawListResponse{
		Draws:      draws,
		TotalCount: total,
		LatestDraw: latest,
	}, nil
}

// GetDrawByNo 특정 회차 당첨번호 조회
func (s *Service) GetDrawByNo(ctx context.Context, drawNo int) (*LottoDraw, error) {
	return s.repo.GetDrawByNo(ctx, drawNo)
}

// GetStats 통계 조회
func (s *Service) GetStats(ctx context.Context) (*StatsResponse, error) {
	numberStats, err := s.repo.GetAllNumberStats(ctx)
	if err != nil {
		return nil, err
	}

	reappearStats, err := s.repo.GetAllReappearStats(ctx)
	if err != nil {
		return nil, err
	}

	latestDraw, err := s.repo.GetLatestDrawNo(ctx)
	if err != nil {
		return nil, err
	}

	return &StatsResponse{
		NumberStats:   numberStats,
		ReappearStats: reappearStats,
		LatestDrawNo:  latestDraw,
		CalculatedAt:  time.Now(),
	}, nil
}

// GetNumberStats 번호별 통계 조회
func (s *Service) GetNumberStats(ctx context.Context) ([]NumberStat, error) {
	return s.repo.GetAllNumberStats(ctx)
}

// GetReappearStats 재등장 통계 조회
func (s *Service) GetReappearStats(ctx context.Context) ([]ReappearStat, error) {
	return s.repo.GetAllReappearStats(ctx)
}

// GetFirstLastStats 첫번째/마지막 번호 확률 조회
func (s *Service) GetFirstLastStats(ctx context.Context) (*FirstLastStatsResponse, error) {
	return s.analyzer.CalculateFirstLastStats(ctx)
}

// GetPairStats 번호 쌍 동반 출현 통계 조회
func (s *Service) GetPairStats(ctx context.Context, topN int) (*PairStatsResponse, error) {
	return s.analyzer.CalculatePairStats(ctx, topN)
}

// GetConsecutiveStats 연번 패턴 통계 조회
func (s *Service) GetConsecutiveStats(ctx context.Context) (*ConsecutiveStatsResponse, error) {
	return s.analyzer.CalculateConsecutiveStats(ctx)
}

// GetRatioStats 홀짝/고저 비율 통계 조회
func (s *Service) GetRatioStats(ctx context.Context) (*RatioStatsResponse, error) {
	return s.analyzer.CalculateRatioStats(ctx)
}

// TriggerSync 수동 동기화 (관리자용)
func (s *Service) TriggerSync(ctx context.Context) error {
	if err := s.FetchNewDraw(ctx); err != nil {
		return err
	}
	return s.RunAnalysis(ctx)
}
