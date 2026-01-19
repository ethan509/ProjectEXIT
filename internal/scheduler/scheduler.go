package scheduler

import (
	"context"
	"time"

	"github.com/example/ProjectEXIT/internal/config"
	"github.com/example/ProjectEXIT/internal/logger"
)

type Scheduler struct {
	tz   *time.Location
	log  *logger.Logger
	quit chan struct{}
}

func New(cfg config.Config, log *logger.Logger) (*Scheduler, error) {
	loc, err := time.LoadLocation(cfg.Scheduler.Timezone)
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		tz:   loc,
		log:  log,
		quit: make(chan struct{}),
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
			// 매주 일요일(Sunday) 06시 00분 체크
			if now.Weekday() == time.Sunday && now.Hour() == 6 && now.Minute() == 0 &&
				now.YearDay() != lastYearDay {
				lastYearDay = now.YearDay()
				s.log.Infof("running weekly job at %s", now)
				go s.runWeekly(now)
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

func (s *Scheduler) runWeekly(t time.Time) {
	s.log.Infof("weekly job executed at %s", t)
}

func (s *Scheduler) runMonthly(t time.Time) {
	s.log.Infof("monthly job executed at %s", t)
}

func (s *Scheduler) runYearly(t time.Time) {
	s.log.Infof("yearly job executed at %s", t)
}
