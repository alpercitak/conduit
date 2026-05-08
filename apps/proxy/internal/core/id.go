package core

import (
	"crypto/rand"
	"encoding/hex"
)

func NewRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
