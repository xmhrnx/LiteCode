package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
)

// GeminiBaseURL is Google's Generative Language API.
const GeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// geminiClient speaks Gemini's native generateContent API.
type geminiClient struct {
	p   Provider
	key string
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

func (c *geminiClient) auth(h http.Header) {
	setKey(h, c.p, c.key, "x-goog-api-key", "")
}

func (c *geminiClient) body(r Request) map[string]any {
	contents := make([]geminiContent, 0, len(r.Messages))
	for _, m := range r.Messages {
		role := "user"
		if m.Role == Assistant {
			role = "model"
		}
		contents = append(contents, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}
	body := map[string]any{
		"contents":         contents,
		"generationConfig": map[string]any{"maxOutputTokens": maxTokens(r)},
	}
	if r.System != "" {
		body["systemInstruction"] = geminiContent{Parts: []geminiPart{{Text: r.System}}}
	}
	return body
}

func (c *geminiClient) Stream(ctx context.Context, r Request, onText func(string)) error {
	model := strings.TrimPrefix(r.Model, "models/")
	url := join(c.p.BaseURL, "models/"+model+":streamGenerateContent?alt=sse")
	req, err := newRequest(ctx, c.p, http.MethodPost, url, c.body(r), c.auth)
	if err != nil {
		return err
	}
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	wrote := false
	blocked := ""
	err = readSSE(resp.Body, func(data []byte) error {
		var chunk struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text    string `json:"text"`
						Thought bool   `json:"thought"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason"`
			} `json:"candidates"`
			PromptFeedback struct {
				BlockReason string `json:"blockReason"`
			} `json:"promptFeedback"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &chunk); err != nil {
			return nil
		}
		if chunk.Error != nil {
			return errors.New(chunk.Error.Message)
		}
		if chunk.PromptFeedback.BlockReason != "" {
			blocked = chunk.PromptFeedback.BlockReason
		}
		for _, cand := range chunk.Candidates {
			for _, part := range cand.Content.Parts {
				if part.Text != "" && !part.Thought {
					wrote = true
					onText(part.Text)
				}
			}
			if cand.FinishReason == "SAFETY" || cand.FinishReason == "PROHIBITED_CONTENT" {
				blocked = cand.FinishReason
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if blocked != "" && !wrote {
		return ErrRefused
	}
	return nil
}

func (c *geminiClient) ListModels(ctx context.Context) ([]string, error) {
	var ids []string
	token := ""
	for {
		url := join(c.p.BaseURL, "models?pageSize=1000")
		if token != "" {
			url += "&pageToken=" + token
		}
		req, err := newRequest(ctx, c.p, http.MethodGet, url, nil, c.auth)
		if err != nil {
			return nil, err
		}
		resp, err := do(req)
		if err != nil {
			return nil, err
		}
		var body struct {
			Models []struct {
				Name    string   `json:"name"`
				Methods []string `json:"supportedGenerationMethods"`
			} `json:"models"`
			NextPageToken string `json:"nextPageToken"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		for _, m := range body.Models {
			for _, method := range m.Methods {
				if method == "generateContent" {
					ids = append(ids, strings.TrimPrefix(m.Name, "models/"))
					break
				}
			}
		}
		if body.NextPageToken == "" {
			break
		}
		token = body.NextPageToken
	}
	sort.Strings(ids)
	return ids, nil
}
