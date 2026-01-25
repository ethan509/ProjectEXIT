package scheduler

import (
	"context"
	"time"

	"github.com/example/LottoSmash/internal/config"
	"github.com/example/LottoSmash/internal/logger"
	"github.com/example/LottoSmash/internal/lotto"
)

type Scheduler struct {
	tz       *time.Location
	log      *logger.Logger
	lottoSvc *lotto.Service
	quit     chan struct{}
}

func New(cfg config.Config, log *logger.Logger, lottoSvc *lotto.Service) (*Scheduler, error) {
	loc, err := time.LoadLocation(cfg.Scheduler.Timezone)
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		tz:       loc,
		log:      log,
		lottoSvc: lottoSvc,
		quit:     make(chan struct{}),
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	go s.loopDaily(ctx)
	go s.loopWeekly(ctx)
	go s.loopMonthly(ctx)
	go s.loopYearly(ctx)
}

func (s *Scheduler) loopDaily(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	var lastDay int
	for {
		select {
		case <-ctx.Done():
			s.log.Infof("daily scheduler stopping")
			return
		case <-ticker.C:
			now := time.Now().In(s.tz)
			if now.Hour() == 6 && now.Minute() == 0 && now.Day() != lastDay {
				lastDay = now.Day()
				s.log.Infof("running daily job at %s", now)
				go s.runDaily(now)
			}
		}
	}
}

func (s *Scheduler) loopWeekly(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	var lastYearDay int
	for {
		select {
		case <-ctx.Done():
			s.log.Infof("weekly scheduler stopping")
			return
		case <-ticker.C:
			now := time.Now().In(s.tz)
			// 매주 토요일(Saturday) 21시 27분 체크 (로또 추첨 결과 수집)
			if now.Weekday() == time.Saturday && now.Hour() == 21 && now.Minute() == 27 &&
				now.YearDay() != lastYearDay {
				lastYearDay = now.YearDay()
				s.log.Infof("running weekly lotto job at %s", now)
				go s.runWeekly(ctx)
			}
		}
	}
}

func (s *Scheduler) loopMonthly(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	var lastMonth int
	var lastYear int
	for {
		select {
		case <-ctx.Done():
			s.log.Infof("monthly scheduler stopping")
			return
		case <-ticker.C:
			now := time.Now().In(s.tz)
			if now.Day() == 1 && now.Hour() == 6 && now.Minute() == 0 &&
				(now.Month() != time.Month(lastMonth) || now.Year() != lastYear) {
				lastMonth = int(now.Month())
				lastYear = now.Year()
				s.log.Infof("running monthly job at %s", now)
				go s.runMonthly(now)
			}
		}
	}
}

func (s *Scheduler) loopYearly(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	var lastYear int
	for {
		select {
		case <-ctx.Done():
			s.log.Infof("yearly scheduler stopping")
			return
		case <-ticker.C:
			now := time.Now().In(s.tz)
			if now.Month() == time.January && now.Day() == 1 && now.Hour() == 6 && now.Minute() == 0 &&
				now.Year() != lastYear {
				lastYear = now.Year()
				s.log.Infof("running yearly job at %s", now)
				go s.runYearly(now)
			}
		}
	}
}

func (s *Scheduler) runDaily(t time.Time) {
	s.log.Infof("daily job executed at %s", t)
}

func (s *Scheduler) runWeekly(ctx context.Context) {
	s.log.Infof("weekly lotto job started")

	if s.lottoSvc == nil {
		s.log.Infof("lotto service not initialized, skipping weekly job")
		return
	}

	// 새 당첨번호 수집
	if err := s.lottoSvc.FetchNewDraw(ctx); err != nil {
		s.log.Errorf("failed to fetch new lotto draw: %v", err)
		return
	}

	// 분석 실행
	if err := s.lottoSvc.RunAnalysis(ctx); err != nil {
		s.log.Errorf("failed to run lotto analysis: %v", err)
		return
	}

	s.log.Infof("weekly lotto job completed")
}

func (s *Scheduler) runMonthly(t time.Time) {
	s.log.Infof("monthly job executed at %s", t)
}

func (s *Scheduler) runYearly(t time.Time) {
	s.log.Infof("yearly job executed at %s", t)
}
