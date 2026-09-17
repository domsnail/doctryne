package nvd

import "net/http"

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept-Encoding", "gzip") // 9.1722302s ==> 926.3489ms
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("apiKey", c.token)
	}
}
