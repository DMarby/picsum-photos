package hmac_test

import (
	"testing"

	"github.com/DMarby/picsum-photos/internal/hmac"
)

var key = []byte("foobar")
var message = "test"

func TestHMAC(t *testing.T) {
	h := &hmac.HMAC{
		Key: key,
	}

	mac, err := h.Create(message)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := h.Validate(message, mac)
	if err != nil {
		t.Fatal(err)
	}

	if !matches {
		t.Error("hmac does not match")
	}

	matches, err = h.Validate("doesnotmatch", mac)
	if err != nil {
		t.Fatal(err)
	}

	if matches {
		t.Error("hmac matches when it should not")
	}
}
