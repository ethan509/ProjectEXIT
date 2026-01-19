// # TxID 생성/컨텍스트 저장/조회
package txid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type ctxKey struct{}

func FromContext(ctx context.Context) string {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func WithTxID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func NewID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
