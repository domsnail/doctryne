package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

const (
	mimeJSON = "application/json"
	mimeHTML = "text/html"
)

// offeredMediaTypes lists supported response types in server preference order.
// The first one is used when the client does not send an Accept header.
var offeredMediaTypes = []string{mimeJSON, mimeHTML}

type AcceptMux struct {
	json *http.ServeMux
	html *http.ServeMux
}

func NewAcceptMux(opts *HandlerOptions) *AcceptMux {
	if opts.InspectionService == nil || opts.DeveloperService == nil || opts.VulnerabilityService == nil {
		panic("service is nil")
	}

	jsonMux := http.NewServeMux()
	jsonHandler := newJSONHandler(opts)
	jsonHandler.HandleMux(jsonMux)

	htmlMux := http.NewServeMux()
	htmlHandler := newHTMLHandler(opts)
	htmlHandler.HandleMux(htmlMux)

	return &AcceptMux{
		json: jsonMux,
		html: htmlMux,
	}
}

func (mux AcceptMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// A request may carry several Accept header lines, which are equivalent to one comma-separated list.
	accept := strings.Join(r.Header.Values("Accept"), ",")

	w.Header().Add("Vary", "Accept")

	switch negotiateMediaType(accept, offeredMediaTypes) {
	case mimeJSON:
		w.Header().Set("Content-Type", mimeJSON)
		mux.json.ServeHTTP(w, r)
	case mimeHTML:
		w.Header().Set("Content-Type", mimeHTML)
		mux.html.ServeHTTP(w, r)
	default:
		slog.WarnContext(r.Context(), "accept header is not supported",
			slog.String("accept", accept),
		)

		http.Error(w, "406 Not Acceptable", http.StatusNotAcceptable)
	}
}

type mediaRange struct {
	typ     string
	subtype string
	quality float64
}

// negotiateMediaType picks the offer the client prefers most according to the Accept header (RFC 9110, section 12.5.1).
// Each offer takes the quality of the most specific media range matching it. The highest quality wins,
// ties are broken by match specificity and then by offer order. It returns "" when no offer is acceptable.
func negotiateMediaType(accept string, offers []string) string {
	if len(offers) == 0 {
		return ""
	}

	if strings.TrimSpace(accept) == "" {
		return offers[0]
	}

	ranges := parseAccept(accept)

	best := ""
	bestQuality, bestSpecificity := 0.0, -1

	for _, offer := range offers {
		typ, subtype, _ := strings.Cut(offer, "/")

		quality, specificity := 0.0, -1
		for _, mr := range ranges {
			s := mr.specificity(typ, subtype)
			if s > specificity {
				quality, specificity = mr.quality, s
			}
		}

		if quality <= 0 {
			continue
		}

		if quality > bestQuality || (quality == bestQuality && specificity > bestSpecificity) {
			best, bestQuality, bestSpecificity = offer, quality, specificity
		}
	}

	return best
}

// specificity reports how precisely the range matches the media type:
// 2 for type/subtype, 1 for type/*, 0 for */* and -1 for no match.
func (mr mediaRange) specificity(typ, subtype string) int {
	switch {
	case mr.typ == "*" && mr.subtype == "*":
		return 0
	case mr.typ == typ && mr.subtype == "*":
		return 1
	case mr.typ == typ && mr.subtype == subtype:
		return 2
	default:
		return -1
	}
}

// parseAccept parses an Accept header value into media ranges, skipping malformed entries.
func parseAccept(accept string) []mediaRange {
	var ranges []mediaRange

	for part := range strings.SplitSeq(accept, ",") {
		params := strings.Split(part, ";")

		mediaType := strings.ToLower(strings.TrimSpace(params[0]))
		if mediaType == "" {
			continue
		}

		// Some clients send a bare "*" instead of "*/*".
		if mediaType == "*" {
			mediaType = "*/*"
		}

		typ, subtype, ok := strings.Cut(mediaType, "/")
		if !ok || typ == "" || subtype == "" || (typ == "*" && subtype != "*") {
			continue
		}

		mr := mediaRange{typ: typ, subtype: subtype, quality: 1}
		valid := true

		for _, param := range params[1:] {
			key, value, _ := strings.Cut(param, "=")
			if !strings.EqualFold(strings.TrimSpace(key), "q") {
				continue
			}

			q, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil || q < 0 || q > 1 {
				valid = false
			}

			mr.quality = q

			// Parameters after q are accept-extensions and do not affect matching.
			break
		}

		if valid {
			ranges = append(ranges, mr)
		}
	}

	return ranges
}
