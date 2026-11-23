package llm

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"litecode/internal/config"
)

// ProviderSettings are what the user changed about a provider. API keys are
// never stored here: they live in the system's credential store.
type ProviderSettings struct {
	BaseURL string            `json:"baseURL,omitempty"`
	Model   string            `json:"model,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Settings are the assistant's saved configuration.
type Settings struct {
	// Active is the provider the assistant uses.
	Active    string                      `json:"active"`
	Providers map[string]ProviderSettings `json:"providers"`
	// Custom are the providers the user added.
	Custom []Provider `json:"custom"`
	// SystemPrompt is prepended to every conversation.
	SystemPrompt string `json:"systemPrompt"`
	MaxTokens    int    `json:"maxTokens"`
}

// DefaultSystemPrompt frames the assistant as a coding helper in the editor.
const DefaultSystemPrompt = "You are a coding assistant inside the LiteCode editor. Answer concisely. " +
	"Use Markdown with fenced code blocks that name their language. When the user shares a file, refer to it by name."

// LoadSettings reads ai.json from the config directory.
func LoadSettings() (Settings, error) {
	s := Settings{Providers: map[string]ProviderSettings{}, SystemPrompt: DefaultSystemPrompt, MaxTokens: DefaultMaxTokens}
	path, err := config.Path("ai.json")
	if err != nil {
		return s, err
	}
	if err := config.LoadJSON(path, &s); err != nil {
		return s, err
	}
	if s.Providers == nil {
		s.Providers = map[string]ProviderSettings{}
	}
	if s.MaxTokens <= 0 {
		s.MaxTokens = DefaultMaxTokens
	}
	return s, nil
}

// Save writes ai.json.
func (s Settings) Save() error {
	path, err := config.Path("ai.json")
	if err != nil {
		return err
	}
	return config.SaveJSON(path, s)
}

// All lists the built-in providers, then the custom ones.
func (s Settings) All() []Provider {
	return append(Catalog(), s.Custom...)
}

// Provider returns a provider with the user's changes applied.
func (s Settings) Provider(id string) (Provider, bool) {
	for _, p := range s.All() {
		if p.ID != id {
			continue
		}
		o := s.Providers[id]
		if o.BaseURL != "" {
			p.BaseURL = o.BaseURL
		}
		if o.Model != "" {
			p.DefaultModel = o.Model
		}
		if len(o.Headers) > 0 {
			h := maps.Clone(p.Headers)
			if h == nil {
				h = map[string]string{}
			}
			maps.Copy(h, o.Headers)
			p.Headers = h
		}
		return p, true
	}
	return Provider{}, false
}

// AddCustom adds (or replaces, by ID) a custom provider, giving it an ID
// from its name when it has none.
func (s *Settings) AddCustom(p Provider) Provider {
	p.Kind = Custom
	if p.ID == "" {
		p.ID = s.uniqueID("custom-" + slug(p.Name))
	}
	s.Custom = slices.DeleteFunc(s.Custom, func(c Provider) bool { return c.ID == p.ID })
	s.Custom = append(s.Custom, p)
	return p
}

// RemoveCustom drops a custom provider and its settings.
func (s *Settings) RemoveCustom(id string) {
	s.Custom = slices.DeleteFunc(s.Custom, func(c Provider) bool { return c.ID == id })
	delete(s.Providers, id)
	if s.Active == id {
		s.Active = ""
	}
}

func (s Settings) uniqueID(base string) string {
	id := base
	for n := 2; ; n++ {
		if _, taken := s.Provider(id); !taken {
			return id
		}
		id = fmt.Sprintf("%s-%d", base, n)
	}
}

func slug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	if s := strings.Trim(b.String(), "-"); s != "" {
		return s
	}
	return "provider"
}
