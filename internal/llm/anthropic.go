package llm

import (
	"context"
	"sort"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

// AnthropicBaseURL is the first-party Claude API.
const AnthropicBaseURL = "https://api.anthropic.com"

// anthropicClient calls the Messages API through Anthropic's Go SDK.
type anthropicClient struct {
	p      Provider
	client anthropic.Client
}

func newAnthropicClient(p Provider, key string) *anthropicClient {
	opts := []option.RequestOption{option.WithMaxRetries(2)}
	if p.BaseURL != "" && p.BaseURL != AnthropicBaseURL {
		opts = append(opts, option.WithBaseURL(p.BaseURL))
	}
	switch {
	case p.AuthHeader != "" && key != "":
		// A gateway with its own auth scheme in front of the Messages API.
		opts = append(opts, option.WithHeader(p.AuthHeader, p.AuthPrefix+key))
	case key != "":
		opts = append(opts, option.WithAPIKey(key))
	}
	for k, v := range p.Headers {
		opts = append(opts, option.WithHeader(k, v))
	}
	return &anthropicClient{p: p, client: anthropic.NewClient(opts...)}
}

// firstParty reports whether the client talks to Anthropic itself, where
// server-side refusal fallbacks are available.
func (c *anthropicClient) firstParty() bool {
	return c.p.BaseURL == "" || c.p.BaseURL == AnthropicBaseURL
}

func (c *anthropicClient) params(r Request) anthropic.BetaMessageNewParams {
	msgs := make([]anthropic.BetaMessageParam, 0, len(r.Messages))
	for _, m := range r.Messages {
		block := anthropic.NewBetaTextBlock(m.Content)
		if m.Role == Assistant {
			msgs = append(msgs, anthropic.BetaMessageParam{OfAssistant: &anthropic.BetaAssistantMessageParam{
				Content: []anthropic.BetaContentBlockParamUnion{block},
			}})
			continue
		}
		msgs = append(msgs, anthropic.NewBetaUserMessage(block))
	}
	p := anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(r.Model),
		MaxTokens: int64(maxTokens(r)),
		Messages:  msgs,
	}
	if r.System != "" {
		p.System = []anthropic.BetaTextBlockParam{{Text: r.System}}
	}
	if c.firstParty() {
		// When a safety classifier declines, the API re-serves the request
		// on a fallback model it picks by category, within the same call.
		p.Betas = []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01}
		p.Fallbacks = anthropic.BetaFallbacksParamOfDefault()
	}
	return p
}

func (c *anthropicClient) Stream(ctx context.Context, r Request, onText func(string)) error {
	stream := c.client.Beta.Messages.NewStreaming(ctx, c.params(r))
	defer stream.Close()
	wrote := false
	var stop anthropic.BetaStopReason
	for stream.Next() {
		switch ev := stream.Current().AsAny().(type) {
		case anthropic.BetaRawContentBlockDeltaEvent:
			if d, ok := ev.Delta.AsAny().(anthropic.BetaTextDelta); ok && d.Text != "" {
				wrote = true
				onText(d.Text)
			}
		case anthropic.BetaRawMessageDeltaEvent:
			stop = ev.Delta.StopReason
		}
	}
	if err := stream.Err(); err != nil {
		return err
	}
	if stop == anthropic.BetaStopReasonRefusal && !wrote {
		return ErrRefused
	}
	return nil
}

func (c *anthropicClient) ListModels(ctx context.Context) ([]string, error) {
	pager := c.client.Models.ListAutoPaging(ctx, anthropic.ModelListParams{})
	var ids []string
	for pager.Next() {
		ids = append(ids, pager.Current().ID)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}
	sort.Strings(ids)
	return ids, nil
}
