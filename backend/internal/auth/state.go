package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	StateCookieName = "oauth_state"
	StateSeparator  = "."
	StateExpiry     = 10 * time.Minute
)

var ErrInvalidStateFormat = errors.New("invalid state format")
var ErrInvalidStateMAC = errors.New("invalid state MAC (tampered?)")
var ErrStateExpired = errors.New("state expired")

// signState generates a timestamped and HMAC-signed state string.
// Format: <original_state>.<timestamp>.<signature>
func SignState(stateValue string, secretKey []byte) string {
	if len(secretKey) == 0 {
		// Should not happen in production if configured correctly
		panic("OAuth state signing secret cannot be empty")
	}
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s%s%d", stateValue, StateSeparator, timestamp)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s%s%s", message, StateSeparator, signature)
}

// verifyAndExtractState checks the signature and expiry, returning the original state value.
func VerifyAndExtractState(signedState string, secretKey []byte) (string, error) {
	if len(secretKey) == 0 {
		panic("OAuth state signing secret cannot be empty")
	}
	parts := strings.Split(signedState, StateSeparator)
	if len(parts) != 3 {
		return "", ErrInvalidStateFormat
	}

	originalState := parts[0]
	timestampStr := parts[1]
	receivedSignature := parts[2]

	message := fmt.Sprintf("%s%s%s", originalState, StateSeparator, timestampStr)
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(receivedSignature), []byte(expectedSignature)) {
		return "", ErrInvalidStateMAC
	}

	var timestamp int64
	if _, err := fmt.Sscan(timestampStr, &timestamp); err != nil {
		return "", fmt.Errorf("invalid timestamp in state: %w", ErrInvalidStateFormat)
	}
	if time.Since(time.Unix(timestamp, 0)) > StateExpiry {
		return "", ErrStateExpired
	}

	return originalState, nil
}
