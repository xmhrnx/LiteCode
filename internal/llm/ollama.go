package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
)

// OllamaBaseURL is a local Ollama server.
const OllamaBaseURL = "http://localhost:11434"

// ollamaClient speaks Ollama's native API, which streams one JSON object a
// line and lists every pulled model.
type ollamaClient struct {
	p   Provider
	key string
}

func (c *ollamaClient) auth(h http.Header) {
	setKey(h, c.p, c.key, "Authorization", "Bearer ")
}

func (c *ollamaClient) Stream(ctx context.Context, r Request, onText func(string)) error {
	msgs := make([]openAIMessage, 0, len(r.Messages)+1)
	if r.System != "" {
		msgs = append(msgs, openAIMessage{Role: "system", Content: r.System})
	}
	for _, m := range r.Messages {
		msgs = append(msgs, openAIMessage{Role: string(m.Role), Content: m.Content})
	}
	body := map[string]any{
		"model":    r.Model,
		"messages": msgs,
		"stream":   true,
		"options":  map[string]any{"num_predict": maxTokens(r)},
	}
	req, err := newRequest(ctx, c.p, http.MethodPost, join(c.p.BaseURL, "api/chat"), body, c.auth)
	if err != nil {
		return err
	}
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64<<10), 8<<20)
	for sc.Scan() {
		var line struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done  bool   `json:"done"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal(sc.Bytes(), &line); err != nil {
			continue
		}
		if line.Error != "" {
			return errors.New(line.Error)
		}
		if line.Message.Content != "" {
			onText(line.Message.Content)
		}
		if line.Done {
			return nil
		}
	}
	return sc.Err()
}

func (c *ollamaClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := newRequest(ctx, c.p, http.MethodGet, join(c.p.BaseURL, "api/tags"), nil, c.auth)
	if err != nil {
		return nil, err
	}
	resp, err := do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(body.Models))
	for _, m := range body.Models {
		ids = append(ids, m.Name)
	}
	sort.Strings(ids)
	return ids, nil
}
