// Package llm talks to language model providers: hosted APIs (Anthropic,
// OpenAI, Gemini and the OpenAI-compatible services), local runtimes (Ollama,
// LM Studio, llama.cpp, vLLM, …) and custom endpoints the user adds. Each
// provider speaks one wire protocol; Client hides which.
package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Protocol is a wire format.
type Protocol string

const (
	// OpenAI is the Chat Completions API, which most providers and local
	// runtimes also serve.
	OpenAI Protocol = "openai"
	// Anthropic is the Claude Messages API, called through Anthropic's SDK.
	Anthropic Protocol = "anthropic"
	// Gemini is Google's native generateContent API.
	Gemini Protocol = "gemini"
	// Ollama is Ollama's native /api/chat API.
	Ollama Protocol = "ollama"
)

// Kind groups providers in the settings.
type Kind string

const (
	Cloud  Kind = "cloud"
	Local  Kind = "local"
	Custom Kind = "custom"
)

// Provider describes an endpoint and how to call it.
type Provider struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Kind     Kind     `json:"kind"`
	Protocol Protocol `json:"protocol"`
	BaseURL  string   `json:"baseURL"`
	// NeedsKey says whether the provider requires an API key.
	NeedsKey bool `json:"needsKey"`
	// AuthHeader and AuthPrefix say how the key is sent; empty means the
	// protocol's standard header.
	AuthHeader string            `json:"authHeader,omitempty"`
	AuthPrefix string            `json:"authPrefix,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	// Query is added to every request URL (Azure's api-version).
	Query map[string]string `json:"query,omitempty"`
	// DefaultModel and Models are suggestions; ListModels asks the endpoint.
	DefaultModel string   `json:"defaultModel"`
	Models       []string `json:"models,omitempty"`
	// TokenField is the request field for the output limit when it isn't
	// the protocol's default ("max_completion_tokens" for OpenAI).
	TokenField string `json:"tokenField,omitempty"`
	// KeyEnv is an environment variable the key is read from when none is
	// saved (OPENAI_API_KEY).
	KeyEnv  string `json:"keyEnv,omitempty"`
	KeyURL  string `json:"keyURL,omitempty"`
	DocsURL string `json:"docsURL,omitempty"`
	Note    string `json:"note,omitempty"`
}

// Role is who wrote a message.
type Role string

const (
	User      Role = "user"
	Assistant Role = "assistant"
)

// Message is one turn of a conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Request is one completion.
type Request struct {
	Model     string
	System    string
	Messages  []Message
	MaxTokens int
}

// DefaultMaxTokens bounds replies when a request sets no limit.
const DefaultMaxTokens = 16000

// Client calls one provider.
type Client interface {
	// Stream sends the request and calls onText with each piece of the reply
	// as it arrives, on the calling goroutine.
	Stream(ctx context.Context, req Request, onText func(string)) error
	// ListModels returns the model IDs the endpoint offers.
	ListModels(ctx context.Context) ([]string, error)
}

// ErrNoKey means the provider needs an API key and none is set.
var ErrNoKey = errors.New("no API key set")

// ErrRefused means the model declined to answer.
var ErrRefused = errors.New("the model declined this request")

// httpClient has no overall timeout: replies stream for as long as they
// take, and callers cancel through the context.
var httpClient = &http.Client{Transport: &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	ResponseHeaderTimeout: 2 * time.Minute,
	IdleConnTimeout:       90 * time.Second,
}}

// NewClient returns a client for a provider, with its API key ("" for none).
func NewClient(p Provider, key string) (Client, error) {
	if p.NeedsKey && strings.TrimSpace(key) == "" {
		return nil, ErrNoKey
	}
	switch p.Protocol {
	case OpenAI:
		return &openAIClient{p: p, key: key}, nil
	case Anthropic:
		return newAnthropicClient(p, key), nil
	case Gemini:
		return &geminiClient{p: p, key: key}, nil
	case Ollama:
		return &ollamaClient{p: p, key: key}, nil
	}
	return nil, fmt.Errorf("unknown protocol %q", p.Protocol)
}

// join appends path to a base URL without doubling slashes.
func join(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func maxTokens(r Request) int {
	if r.MaxTokens > 0 {
		return r.MaxTokens
	}
	return DefaultMaxTokens
}
