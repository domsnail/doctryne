package http

import (
	"net/http"
)

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
	accept := r.Header.Get("Accept")

	switch accept {
	case "*/*", "application/json":
		w.Header().Set("Content-Type", "application/json")
		mux.json.ServeHTTP(w, r)
		return
	case "text/html":
		w.Header().Set("Content-Type", "text/html")
		mux.html.ServeHTTP(w, r)
		return
	default:
		http.Error(w, "415 Unsupported Media Type", http.StatusUnsupportedMediaType)
	}
}
