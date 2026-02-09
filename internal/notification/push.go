package notification

import (
	"context"

	"github.com/example/LottoSmash/internal/logger"
)

// PushSender push 알림 발송 인터페이스
type PushSender interface {
	Send(ctx context.Context, token string, title string, body string, data map[string]string) error
}

// NoopPushSender FCM 미설정 시 사용하는 더미 구현체 (로그만 출력)
type NoopPushSender struct {
	log *logger.Logger
}

func NewNoopPushSender(log *logger.Logger) *NoopPushSender {
	return &NoopPushSender{log: log}
}

func (n *NoopPushSender) Send(ctx context.Context, token string, title string, body string, data map[string]string) error {
	n.log.Infof("[NoopPush] token=%s title=%q body=%q data=%v", token, title, body, data)
	return nil
}
