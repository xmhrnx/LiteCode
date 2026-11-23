package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// server answers one path with a handler and records the request.
type server struct {
	*httptest.Server
	got  *http.Request
	body map[string]any
}

func newServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *server {
	t.Helper()
	s := &server{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.got = r
		data, _ := io.ReadAll(r.Body)
		s.body = nil
		_ = json.Unmarshal(data, &s.body)
		handler(w, r)
	}))
	t.Cleanup(s.Close)
	return s
}

func stream(t *testing.T, c Client, req Request) string {
	t.Helper()
	var out strings.Builder
	if err := c.Stream(context.Background(), req, func(s string) { out.WriteString(s) }); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

var chat = Request{Model: "m1", System: "be brief", Messages: []Message{{Role: User, Content: "hi"}, {Role: Assistant, Content: "yo"}, {Role: User, Content: "again"}}}

func TestOpenAIStreamAndModels(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			fmt.Fprint(w, `{"object":"list","data":[{"id":"b"},{"id":"a"}]}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": keep-alive\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	})
	p := Provider{Protocol: OpenAI, BaseURL: s.URL + "/v1", NeedsKey: true, TokenField: "max_completion_tokens", Headers: map[string]string{"X-Title": "LiteCode"}}
	c, err := NewClient(p, "sk-1")
	if err != nil {
		t.Fatal(err)
	}
	if got := stream(t, c, chat); got != "Hello" {
		t.Fatalf("text %q", got)
	}
	if s.got.URL.Path != "/v1/chat/completions" || s.got.Header.Get("Authorization") != "Bearer sk-1" || s.got.Header.Get("X-Title") != "LiteCode" {
		t.Fatalf("request %s %v", s.got.URL.Path, s.got.Header)
	}
	msgs := s.body["messages"].([]any)
	if len(msgs) != 4 || msgs[0].(map[string]any)["role"] != "system" || s.body["stream"] != true || s.body["max_completion_tokens"] == nil || s.body["max_tokens"] != nil {
		t.Fatalf("body %v", s.body)
	}
	ids, err := c.ListModels(context.Background())
	if err != nil || strings.Join(ids, ",") != "a,b" {
		t.Fatalf("models %v %v", ids, err)
	}
}

func TestOpenAIErrorMessage(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		fmt.Fprint(w, `{"error":{"message":"Incorrect API key provided","type":"invalid_request_error"}}`)
	})
	c, _ := NewClient(Provider{Protocol: OpenAI, BaseURL: s.URL}, "")
	err := c.Stream(context.Background(), chat, func(string) {})
	if err == nil || err.Error() != "HTTP 401: Incorrect API key provided" {
		t.Fatal(err)
	}
}

func TestAzureStyleAuthAndQuery(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")
	})
	p := Provider{Protocol: OpenAI, BaseURL: s.URL + "/openai/v1", AuthHeader: "api-key", Query: map[string]string{"api-version": "preview"}}
	c, _ := NewClient(p, "az")
	stream(t, c, chat)
	if s.got.Header.Get("api-key") != "az" || s.got.Header.Get("Authorization") != "" || s.got.URL.Query().Get("api-version") != "preview" {
		t.Fatalf("%v %v", s.got.Header, s.got.URL)
	}
}

func TestNeedsKey(t *testing.T) {
	if _, err := NewClient(Provider{Protocol: OpenAI, NeedsKey: true}, " "); err != ErrNoKey {
		t.Fatal(err)
	}
}

func TestGemini(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			fmt.Fprint(w, `{"models":[{"name":"models/gemini-x","supportedGenerationMethods":["generateContent","countTokens"]},{"name":"models/embed","supportedGenerationMethods":["embedContent"]}]}`)
			return
		}
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"thinking\",\"thought\":true}],\"role\":\"model\"}}]}\r\n\r\n")
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Hi \"}],\"role\":\"model\"}}]}\r\n\r\n")
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"there\"}],\"role\":\"model\"},\"finishReason\":\"STOP\"}]}\r\n\r\n")
	})
	c, _ := NewClient(Provider{Protocol: Gemini, BaseURL: s.URL + "/v1beta", NeedsKey: true}, "g-key")
	if got := stream(t, c, Request{Model: "models/gemini-x", System: "sys", Messages: chat.Messages}); got != "Hi there" {
		t.Fatalf("%q", got)
	}
	if s.got.URL.Path != "/v1beta/models/gemini-x:streamGenerateContent" || s.got.URL.Query().Get("alt") != "sse" || s.got.Header.Get("x-goog-api-key") != "g-key" {
		t.Fatalf("%v %v", s.got.URL, s.got.Header)
	}
	contents := s.body["contents"].([]any)
	if contents[1].(map[string]any)["role"] != "model" || s.body["systemInstruction"] == nil {
		t.Fatalf("%v", s.body)
	}
	ids, err := c.ListModels(context.Background())
	if err != nil || len(ids) != 1 || ids[0] != "gemini-x" {
		t.Fatalf("%v %v", ids, err)
	}
}

func TestGeminiBlocked(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data: {\"promptFeedback\":{\"blockReason\":\"SAFETY\"}}\n\n")
	})
	c, _ := NewClient(Provider{Protocol: Gemini, BaseURL: s.URL}, "k")
	if err := c.Stream(context.Background(), chat, func(string) {}); err != ErrRefused {
		t.Fatal(err)
	}
}

func TestOllama(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			fmt.Fprint(w, `{"models":[{"name":"qwen3:8b","model":"qwen3:8b"},{"name":"llama3.2:latest"}]}`)
			return
		}
		fmt.Fprintln(w, `{"model":"m1","message":{"role":"assistant","content":"Lo"},"done":false}`)
		fmt.Fprintln(w, `{"model":"m1","message":{"role":"assistant","content":"cal"},"done":false}`)
		fmt.Fprintln(w, `{"model":"m1","message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}`)
	})
	c, _ := NewClient(Provider{Protocol: Ollama, BaseURL: s.URL}, "")
	if got := stream(t, c, chat); got != "Local" {
		t.Fatalf("%q", got)
	}
	if s.got.URL.Path != "/api/chat" || s.body["options"].(map[string]any)["num_predict"] == nil {
		t.Fatalf("%v %v", s.got.URL, s.body)
	}
	ids, _ := c.ListModels(context.Background())
	if strings.Join(ids, ",") != "llama3.2:latest,qwen3:8b" {
		t.Fatal(ids)
	}
}

func TestAnthropicThroughSDK(t *testing.T) {
	events := []string{
		`event: message_start` + "\n" + `data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"m1","content":[],"stop_reason":null,"usage":{"input_tokens":5,"output_tokens":0}}}`,
		`event: content_block_start` + "\n" + `data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`event: content_block_delta` + "\n" + `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello "}}`,
		`event: content_block_delta` + "\n" + `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Claude"}}`,
		`event: content_block_stop` + "\n" + `data: {"type":"content_block_stop","index":0}`,
		`event: message_delta` + "\n" + `data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}`,
		`event: message_stop` + "\n" + `data: {"type":"message_stop"}`,
	}
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, e := range events {
			fmt.Fprint(w, e+"\n\n")
		}
	})
	c, _ := NewClient(Provider{Protocol: Anthropic, BaseURL: s.URL, NeedsKey: true}, "sk-ant")
	if got := stream(t, c, chat); got != "Hello Claude" {
		t.Fatalf("%q", got)
	}
	if s.got.URL.Path != "/v1/messages" || s.got.Header.Get("Authorization") != "Bearer sk-ant" || s.got.Header.Get("anthropic-version") == "" {
		t.Fatalf("%v %v", s.got.URL, s.got.Header)
	}
	if s.body["stream"] != true || s.body["system"] == nil || len(s.body["messages"].([]any)) != 3 {
		t.Fatalf("%v", s.body)
	}
	// A gateway isn't Anthropic: no fallback beta there.
	if s.body["fallbacks"] != nil || s.got.Header.Get("anthropic-beta") != "" {
		t.Fatalf("fallbacks sent to a gateway: %v %v", s.body["fallbacks"], s.got.Header.Get("anthropic-beta"))
	}
}

func TestAnthropicFirstPartyUsesFallbacks(t *testing.T) {
	c := newAnthropicClient(Provider{Protocol: Anthropic, BaseURL: AnthropicBaseURL}, "k")
	data, err := json.Marshal(c.params(chat))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"fallbacks":"default"`) {
		t.Fatalf("%s", data)
	}
}

func TestSSEMultiline(t *testing.T) {
	var got []string
	_ = readSSE(strings.NewReader("data: a\ndata: b\n\ndata: [DONE]\n\ndata: c\n\n"), func(d []byte) error {
		got = append(got, string(d))
		return nil
	})
	if len(got) != 1 || got[0] != "a\nb" {
		t.Fatal(got)
	}
}

func TestAnthropicGatewayHeader(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	})
	c, _ := NewClient(Provider{Protocol: Anthropic, BaseURL: s.URL, AuthHeader: "x-api-key"}, "gw")
	stream(t, c, chat)
	if s.got.Header.Get("x-api-key") != "gw" || s.got.Header.Get("Authorization") != "" {
		t.Fatal(s.got.Header)
	}
}

func TestModelsBareArray(t *testing.T) {
	s := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"id":"z"},{"id":"y"}]`)
	})
	c, _ := NewClient(Provider{Protocol: OpenAI, BaseURL: s.URL}, "")
	ids, err := c.ListModels(context.Background())
	if err != nil || strings.Join(ids, ",") != "y,z" {
		t.Fatal(ids, err)
	}
}

func TestCatalogIsConsistent(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range Catalog() {
		if p.ID == "" || p.Name == "" || p.BaseURL == "" || seen[p.ID] {
			t.Errorf("bad provider %+v", p)
		}
		seen[p.ID] = true
		if p.Kind == Cloud && !p.NeedsKey {
			t.Errorf("%s: cloud providers need a key", p.ID)
		}
		if p.DefaultModel != "" && len(p.Models) > 0 && p.Models[0] != p.DefaultModel {
			t.Errorf("%s: default model should lead the list", p.ID)
		}
	}
}

func TestSettingsOverridesAndCustom(t *testing.T) {
	s := Settings{Providers: map[string]ProviderSettings{"ollama": {BaseURL: "http://gpu-box:11434", Model: "qwen3:32b"}}}
	p, ok := s.Provider("ollama")
	if !ok || p.BaseURL != "http://gpu-box:11434" || p.DefaultModel != "qwen3:32b" {
		t.Fatalf("%+v", p)
	}
	c := s.AddCustom(Provider{Name: "My Gateway!", Protocol: OpenAI, BaseURL: "https://gw.example/v1"})
	if c.ID != "custom-my-gateway" || c.Kind != Custom {
		t.Fatalf("%+v", c)
	}
	if c2 := s.AddCustom(Provider{Name: "My Gateway", Protocol: OpenAI}); c2.ID != "custom-my-gateway-2" {
		t.Fatalf("%+v", c2)
	}
	s.Active = c.ID
	s.RemoveCustom(c.ID)
	if _, ok := s.Provider(c.ID); ok || s.Active != "" {
		t.Fatal("not removed")
	}
}
