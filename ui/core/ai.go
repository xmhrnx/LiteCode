package core

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/xmhrnx/giode"

	"litecode/internal/llm"
)

// AI is the assistant's configuration. Keys are read from and written to
// the system's credential store off the UI goroutine; their presence is
// cached here so the settings can show it.
type AI struct {
	Settings llm.Settings
	// HasKey says which providers have a saved key (or one in their env var).
	HasKey map[string]bool
	// Rev goes up when the settings change, for views that cache.
	Rev int
}

// keyName is the credential store entry of a provider's key.
func keyName(providerID string) string { return "llm." + providerID }

// LoadAI reads the assistant settings and checks which keys are saved.
func (ide *IDE) LoadAI() error {
	s, err := llm.LoadSettings()
	ide.AI = &AI{Settings: s, HasKey: map[string]bool{}}
	providers := s.All()
	go func() {
		have := map[string]bool{}
		for _, p := range providers {
			if !p.NeedsKey {
				continue
			}
			if v, err := giode.SecureStore.Get(keyName(p.ID)); err == nil && len(v) > 0 {
				have[p.ID] = true
			} else if p.KeyEnv != "" && os.Getenv(p.KeyEnv) != "" {
				have[p.ID] = true
			}
		}
		ide.App.Post(func() {
			ide.AI.HasKey = have
			ide.AI.Rev++
		})
	}()
	return err
}

// SaveAI writes the assistant settings, reporting a failure.
func (ide *IDE) SaveAI() {
	ide.AI.Rev++
	ide.Error(ide.AI.Settings.Save())
}

// SetKey saves (or, with "", deletes) a provider's key, then runs done.
func (ide *IDE) SetKey(providerID, key string, done func()) {
	go func() {
		var err error
		if key == "" {
			err = giode.SecureStore.Delete(keyName(providerID))
			if errors.Is(err, giode.ErrNotFound) {
				err = nil
			}
		} else {
			err = giode.SecureStore.Set(keyName(providerID), []byte(key))
		}
		ide.App.Post(func() {
			if err != nil {
				ide.Error(errors.New("couldn't save the key in the system keychain: " + err.Error()))
				return
			}
			ide.AI.HasKey[providerID] = key != ""
			ide.AI.Rev++
			if done != nil {
				done()
			}
		})
	}()
}

// providerKey reads a provider's key: the saved one, else its env var. It
// blocks, so call it off the UI goroutine.
func providerKey(p llm.Provider) (string, error) {
	if !p.NeedsKey {
		v, _ := giode.SecureStore.Get(keyName(p.ID)) // optional key, e.g. a secured local server
		return string(v), nil
	}
	v, err := giode.SecureStore.Get(keyName(p.ID))
	if err == nil && len(v) > 0 {
		return string(v), nil
	}
	if p.KeyEnv != "" {
		if env := os.Getenv(p.KeyEnv); env != "" {
			return env, nil
		}
	}
	if err != nil && !errors.Is(err, giode.ErrNotFound) {
		return "", err
	}
	return "", llm.ErrNoKey
}

// ClientFor makes a client for a provider. It reads the key from the
// credential store, so call it off the UI goroutine.
func ClientFor(p llm.Provider) (llm.Client, error) {
	key, err := providerKey(p)
	if err != nil {
		return nil, err
	}
	return llm.NewClient(p, key)
}

// TestResult is the outcome of a connection test.
type TestResult struct {
	OK      bool
	Message string
	Took    time.Duration
}

// TestProvider asks a provider for a one-word reply, then runs done with
// the result on the UI goroutine.
func (ide *IDE) TestProvider(p llm.Provider, model string, done func(TestResult)) {
	go func() {
		start := time.Now()
		res := TestResult{}
		c, err := ClientFor(p)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			var reply []byte
			err = c.Stream(ctx, llm.Request{
				Model: model, MaxTokens: 64,
				Messages: []llm.Message{{Role: llm.User, Content: "Reply with the single word: ready"}},
			}, func(s string) { reply = append(reply, s...) })
			cancel()
			res.Message = string(reply)
		}
		res.Took = time.Since(start)
		if err != nil {
			res.Message = DescribeAIError(p, err)
		} else {
			res.OK = true
		}
		ide.App.Post(func() { done(res) })
	}()
}
