package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// newRequest builds a request to a provider: its auth, extra headers and
// query parameters applied. body is JSON-encoded when not nil.
func newRequest(ctx context.Context, p Provider, method, rawURL string, body any, auth func(h http.Header)) (*http.Request, error) {
	var rd io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rd = bytes.NewReader(data)
	}
	if len(p.Query) > 0 {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for k, v := range p.Query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		rawURL = u.String()
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth != nil {
		auth(req.Header)
	}
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// setKey sends key in the provider's auth header, or the protocol's default.
func setKey(h http.Header, p Provider, key, defHeader, defPrefix string) {
	if key == "" {
		return
	}
	name, prefix := defHeader, defPrefix
	if p.AuthHeader != "" {
		name, prefix = p.AuthHeader, p.AuthPrefix
	}
	h.Set(name, prefix+key)
}

// do sends a request and returns the response, or the provider's error.
func do(req *http.Request) (*http.Response, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		return nil, apiError(resp)
	}
	return resp, nil
}

// apiError turns an error response into an error carrying the provider's
// own message: {"error":{"message":…}}, {"error":"…"} or {"message":…}.
func apiError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	var body struct {
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
		Detail  string          `json:"detail"`
	}
	msg := ""
	if json.Unmarshal(data, &body) == nil {
		var nested struct {
			Message string `json:"message"`
		}
		var plain string
		switch {
		case json.Unmarshal(body.Error, &nested) == nil && nested.Message != "":
			msg = nested.Message
		case json.Unmarshal(body.Error, &plain) == nil && plain != "":
			msg = plain
		case body.Message != "":
			msg = body.Message
		case body.Detail != "":
			msg = body.Detail
		}
	}
	if msg == "" {
		msg = strings.TrimSpace(string(data))
	}
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
}

// readSSE calls fn with the data of each server-sent event until the stream
// ends, "[DONE]" arrives, or fn returns an error.
func readSSE(r io.Reader, fn func(data []byte) error) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64<<10), 8<<20)
	var data []byte
	flush := func() error {
		if len(data) == 0 {
			return nil
		}
		d := data
		data = nil
		if string(d) == "[DONE]" {
			return io.EOF
		}
		return fn(d)
	}
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			if err := flush(); err != nil {
				return ignoreEOF(err)
			}
		case strings.HasPrefix(line, "data:"):
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " ")...)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return ignoreEOF(flush())
}

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}
