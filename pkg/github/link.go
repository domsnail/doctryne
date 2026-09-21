package github

import (
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var nextRx = regexp.MustCompile(`after=(.+)>; rel="next"`)
var prevRx = regexp.MustCompile(`after=(.+)>; rel="prev"`)

type Link struct {
	Next string
	Prev string
}

func newLinkFromHeader(header http.Header) Link {
	if header == nil {
		return Link{}
	}

	link := Link{
		Next: "",
		Prev: "",
	}

	h := header.Get("Link")
	if h == "" {
		return Link{}
	}

	hrefs := strings.SplitSeq(h, ",")
	for hr := range hrefs {
		r := strings.Split(strings.TrimSpace(hr), ";")
		if len(r) != 2 {
			continue
		}

		switch strings.TrimSpace(r[1]) {
		case "rel=\"next\"":
			u, err := url.Parse(strings.Trim(r[0], "<>"))
			if err != nil {
				slog.Warn("failed to parse github link header",
					slog.String("next_link", r[0]),
					slog.String("error", err.Error()),
				)

				continue
			}

			link.Next = u.Query().Get("after")
		default:
			continue
		}
	}

	return link
}
