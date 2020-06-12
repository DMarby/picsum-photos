package hmac

import (
	cryptoHMAC "crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// HMAC is a utility for creating and verifying HMACs
type HMAC struct {
	Key []byte
}

// Create creates a HMAC based on the parameter values, encoded as urlsafe base64
func (h *HMAC) Create(message string) (string, error) {
	mac := cryptoHMAC.New(sha256.New, h.Key)

	_, err := mac.Write([]byte(message))
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// Validate validates that the parameter values matches a given HMAC
func (h *HMAC) Validate(message, mac string) (bool, error) {
	expectedMAC, err := h.Create(message)
	if err != nil {
		return false, err
	}

	return cryptoHMAC.Equal([]byte(mac), []byte(expectedMAC)), nil
}
