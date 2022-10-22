package cmd

import (
	"time"
)

// Http timeouts
const (
	ReadTimeout    = 5 * time.Second
	WriteTimeout   = time.Minute
	HandlerTimeout = 45 * time.Second
)
