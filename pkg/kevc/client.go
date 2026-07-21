package kevc

import (
	"net/http"
	"net/url"
)

const defaultRegistryFileURL = "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json"
const defaultApiURL = "https://api.npmjs.org"

// Client is Known Exploit Vulnerabilities Registry API client
// git ref: https://github.com/cisagov/kev-data
type Client struct {
	h *http.Client

	registry *url.URL
}
