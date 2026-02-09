package notification

import "time"

// DeviceToken FCM 디바이스 토큰
type DeviceToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WinningCheck 추천번호 당첨 확인 결과
type WinningCheck struct {
	ID               int64     `json:"id"`
	RecommendationID int64     `json:"recommendation_id"`
	UserID           *int64    `json:"user_id,omitempty"`
	DrawNo           int       `json:"draw_no"`
	MatchedNumbers   []int     `json:"matched_numbers"`
	MatchedCount     int       `json:"matched_count"`
	BonusMatched     bool      `json:"bonus_matched"`
	PrizeRank        *int      `json:"prize_rank,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// PushNotification 푸시 알림 기록
type PushNotification struct {
	ID           int64      `json:"id"`
	UserID       *int64     `json:"user_id,omitempty"`
	Type         string     `json:"type"`
	Title        string     `json:"title"`
	Body         string     `json:"body"`
	Data         *string    `json:"data,omitempty"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
}

// RegisterTokenRequest 디바이스 토큰 등록 요청
type RegisterTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

// DeleteTokenRequest 디바이스 토큰 삭제 요청
type DeleteTokenRequest struct {
	Token string `json:"token"`
}

// NotificationListResponse 알림 목록 응답
type NotificationListResponse struct {
	Notifications []PushNotification `json:"notifications"`
	TotalCount    int                `json:"total_count"`
}

// WinningListResponse 당첨 결과 목록 응답
type WinningListResponse struct {
	Winnings   []WinningCheck `json:"winnings"`
	TotalCount int            `json:"total_count"`
}
