package params

import (
	"bytes"
	"fmt"
)

// Utilities for building a URL with query params

// BuildQuery builds query parameters for the given arguments
func BuildQuery(grayscale bool, blur bool, blurAmount int) string {
	if !grayscale && !blur {
		return ""
	}

	var buf bytes.Buffer

	if blur {
		addParam(&buf, fmt.Sprintf("blur=%d", blurAmount))
	}

	if grayscale {
		addParam(&buf, "grayscale")
	}

	return buf.String()
}

// addParam adds a query parameter to a byte buffer
func addParam(buf *bytes.Buffer, param string) {
	if buf.Len() > 0 {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}

	buf.WriteString(param)
}
