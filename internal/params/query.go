package params

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Utilities for building a URL with query params

// BuildQuery builds a query parameter string for the given values
// It differs from the stdlib url.Values.Encode in that it encodes query parameters with an empty value as "?key" instead of "?key="
func BuildQuery(v url.Values) string {
	var buf strings.Builder

	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := v.Get(key)

		if value != "" {
			addQueryParam(&buf, fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(value)))
		} else {
			addQueryParam(&buf, fmt.Sprintf("%s", url.QueryEscape(key)))
		}
	}

	return buf.String()
}

// addQueryParam adds a query parameter to a byte buffer
func addQueryParam(buf *strings.Builder, param string) {
	if buf.Len() > 0 {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}

	buf.WriteString(param)
}
