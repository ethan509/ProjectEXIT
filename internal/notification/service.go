package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/example/LottoSmash/internal/logger"
	"github.com/example/LottoSmash/internal/lotto"
)

type Service struct {
	repo     *Repository
	lottoSvc *lotto.Service
	push     PushSender
	log      *logger.Logger
}

func NewService(repo *Repository, lottoSvc *lotto.Service, push PushSender, log *logger.Logger) *Service {
	return &Service{
		repo:     repo,
		lottoSvc: lottoSvc,
		push:     push,
		log:      log,
	}
}

// ProcessNewDraw 새 당첨번호 발표 시 추천번호 대조 및 알림 발송
func (s *Service) ProcessNewDraw(ctx context.Context, drawNo int) error {
	// 이미 처리된 회차인지 확인
	processed, err := s.repo.CheckAlreadyProcessed(ctx, drawNo)
	if err != nil {
		return fmt.Errorf("check already processed: %w", err)
	}
	if processed {
		s.log.Infof("draw %d already processed for winning checks, skipping", drawNo)
		return nil
	}

	// 당첨번호 조회
	draw, err := s.lottoSvc.GetDrawByNo(ctx, drawNo)
	if err != nil {
		return fmt.Errorf("get draw %d: %w", drawNo, err)
	}

	// 최근 7일간 추천 기록 조회
	since := time.Now().AddDate(0, 0, -7)
	recs, err := s.repo.GetRecentRecommendations(ctx, since)
	if err != nil {
		return fmt.Errorf("get recent recommendations: %w", err)
	}

	s.log.Infof("processing draw %d: checking %d recommendations from last 7 days", drawNo, len(recs))

	var winnersCount int
	for _, rec := range recs {
		result := CheckWinning(draw, rec)

		// 당첨 결과 저장
		if err := s.repo.SaveWinningCheck(ctx, &result); err != nil {
			s.log.Errorf("failed to save winning check for recommendation %d: %v", rec.ID, err)
			continue
		}

		// 당첨자에게 push 알림 발송
		if result.PrizeRank != nil {
			winnersCount++
			if err := s.sendWinningNotification(ctx, &result, draw); err != nil {
				s.log.Errorf("failed to send notification to user %v: %v", result.UserID, err)
			}
		}
	}

	s.log.Infof("draw %d winning check completed: %d recommendations checked, %d winners found", drawNo, len(recs), winnersCount)
	return nil
}

// sendWinningNotification 당첨자에게 push 알림 발송
func (s *Service) sendWinningNotification(ctx context.Context, wc *WinningCheck, draw *lotto.LottoDraw) error {
	if wc.UserID == nil {
		return nil
	}

	rankName := PrizeRankName(*wc.PrizeRank)
	title := fmt.Sprintf("축하합니다! 로또 %s 당첨!", rankName)
	body := fmt.Sprintf("%d회차 당첨번호와 %d개 일치! (%s)", draw.DrawNo, wc.MatchedCount, rankName)

	data := map[string]string{
		"type":        "winning_result",
		"draw_no":     strconv.Itoa(draw.DrawNo),
		"prize_rank":  strconv.Itoa(*wc.PrizeRank),
		"match_count": strconv.Itoa(wc.MatchedCount),
	}

	dataJSON, _ := json.Marshal(data)
	dataStr := string(dataJSON)

	// 알림 기록 저장
	pn := &PushNotification{
		UserID: wc.UserID,
		Type:   "winning_result",
		Title:  title,
		Body:   body,
		Data:   &dataStr,
		Status: "pending",
	}
	if err := s.repo.SavePushNotification(ctx, pn); err != nil {
		return fmt.Errorf("save push notification: %w", err)
	}

	// 디바이스 토큰 조회 후 발송
	tokens, err := s.repo.GetActiveTokensByUserID(ctx, *wc.UserID)
	if err != nil {
		s.repo.UpdatePushStatus(ctx, pn.ID, "failed", strPtr("failed to get device tokens"))
		return fmt.Errorf("get device tokens: %w", err)
	}

	if len(tokens) == 0 {
		s.log.Infof("no active device tokens for user %d, skipping push", *wc.UserID)
		s.repo.UpdatePushStatus(ctx, pn.ID, "sent", nil)
		return nil
	}

	var lastErr error
	for _, token := range tokens {
		if err := s.push.Send(ctx, token.Token, title, body, data); err != nil {
			s.log.Errorf("failed to send push to token %s: %v", token.Token, err)
			lastErr = err
			continue
		}
	}

	if lastErr != nil {
		errMsg := lastErr.Error()
		s.repo.UpdatePushStatus(ctx, pn.ID, "failed", &errMsg)
		return lastErr
	}

	s.repo.UpdatePushStatus(ctx, pn.ID, "sent", nil)
	return nil
}

// RegisterDeviceToken 디바이스 토큰 등록
func (s *Service) RegisterDeviceToken(ctx context.Context, userID int64, token, platform string) error {
	return s.repo.UpsertDeviceToken(ctx, userID, token, platform)
}

// DeactivateDeviceToken 디바이스 토큰 비활성화
func (s *Service) DeactivateDeviceToken(ctx context.Context, userID int64, token string) error {
	return s.repo.DeactivateDeviceToken(ctx, userID, token)
}

// GetNotifications 알림 목록 조회
func (s *Service) GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]PushNotification, int, error) {
	return s.repo.GetNotificationsByUserID(ctx, userID, limit, offset)
}

// GetWinnings 당첨 결과 조회
func (s *Service) GetWinnings(ctx context.Context, userID int64, limit, offset int) ([]WinningCheck, int, error) {
	return s.repo.GetWinningsByUserID(ctx, userID, limit, offset)
}

func strPtr(s string) *string {
	return &s
}
