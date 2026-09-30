package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNegotiateMediaType(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   string
	}{
		{name: "empty header", accept: "", want: mimeJSON},
		{name: "any", accept: "*/*", want: mimeJSON},
		{name: "bare asterisk", accept: "*", want: mimeJSON},
		{name: "json", accept: "application/json", want: mimeJSON},
		{name: "html", accept: "text/html", want: mimeHTML},
		{name: "html with charset", accept: "text/html; charset=utf-8", want: mimeHTML},
		{name: "case insensitive", accept: "Text/HTML", want: mimeHTML},
		{name: "firefox", accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", want: mimeHTML},
		{name: "chrome", accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7", want: mimeHTML},
		{name: "specific beats wildcard on tie", accept: "*/*, text/html", want: mimeHTML},
		{name: "higher quality wins", accept: "text/html;q=0.5, application/json", want: mimeJSON},
		{name: "type wildcard", accept: "text/*", want: mimeHTML},
		{name: "specific range overrides wildcard", accept: "*/*, application/json;q=0", want: mimeHTML},
		{name: "server order on full tie", accept: "text/html, application/json", want: mimeJSON},
		{name: "unsupported", accept: "image/png", want: ""},
		{name: "all rejected", accept: "*/*;q=0", want: ""},
		{name: "malformed entries skipped", accept: "garbage, */html, text/html;q=abc, application/json", want: mimeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, negotiateMediaType(tt.accept, offeredMediaTypes))
		})
	}
}
