package http

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/domsnail/doctryne/web"
	"golang.org/x/text/language"
)

func (h *Handler) static() http.HandlerFunc {
	distFS, err := fs.Sub(web.StaticEmbed, "static")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(distFS))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			fileServer.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/styles/") {
			fileServer.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/assets/") {
			fileServer.ServeHTTP(w, r)
			return
		}

		lang := preferredLanguage(r.Header.Get("Accept-Language"))
		switch lang {
		case "ru":
			r.URL.Path = "/ru/" + r.URL.Path // todo: or use redirect
		default:
			r.URL.Path = "/en/" + r.URL.Path
		}

		fileServer.ServeHTTP(w, r)
	}
}

var supportedLanguages = []language.Tag{
	language.English,
	language.Russian,
}

var matcher = language.NewMatcher(supportedLanguages)

func preferredLanguage(acceptLanguageHeader string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLanguageHeader)
	if err != nil {
		return "en"
	}

	tag, _, _ := matcher.Match(tags...)
	base, _ := tag.Base()
	return base.String()
}
