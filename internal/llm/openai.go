package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
)

// openAIClient speaks the Chat Completions API.
type openAIClient struct {
	p   Provider
	key string
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *openAIClient) auth(h http.Header) {
	setKey(h, c.p, c.key, "Authorization", "Bearer ")
}

func (c *openAIClient) body(r Request) map[string]any {
	msgs := make([]openAIMessage, 0, len(r.Messages)+1)
	if r.System != "" {
		msgs = append(msgs, openAIMessage{Role: "system", Content: r.System})
	}
	for _, m := range r.Messages {
		msgs = append(msgs, openAIMessage{Role: string(m.Role), Content: m.Content})
	}
	field := c.p.TokenField
	if field == "" {
		field = "max_tokens"
	}
	return map[string]any{
		"model":    r.Model,
		"messages": msgs,
		"stream":   true,
		field:      maxTokens(r),
	}
}

func (c *openAIClient) Stream(ctx context.Context, r Request, onText func(string)) error {
	req, err := newRequest(ctx, c.p, http.MethodPost, join(c.p.BaseURL, "chat/completions"), c.body(r), c.auth)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return readSSE(resp.Body, func(data []byte) error {
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
					Refusal string `json:"refusal"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &chunk); err != nil {
			return nil // keep-alive comments and vendor extras
		}
		if chunk.Error != nil {
			return errors.New(chunk.Error.Message)
		}
		for _, ch := range chunk.Choices {
			if ch.Delta.Content != "" {
				onText(ch.Delta.Content)
			}
			if ch.Delta.Refusal != "" {
				onText(ch.Delta.Refusal)
			}
		}
		return nil
	})
}

func (c *openAIClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := newRequest(ctx, c.p, http.MethodGet, join(c.p.BaseURL, "models"), nil, c.auth)
	if err != nil {
		return nil, err
	}
	resp, err := do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Most servers answer {"data":[…]}; some (Mistral's docs, Together)
	// show a bare array.
	type model struct {
		ID string `json:"id"`
	}
	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	var list []model
	var wrapped struct {
		Data []model `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Data != nil {
		list = wrapped.Data
	} else if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(list))
	for _, m := range list {
		ids = append(ids, m.ID)
	}
	sort.Strings(ids)
	return ids, nil
}
