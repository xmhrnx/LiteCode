package core

import (
	"errors"
	"strings"

	"litecode/internal/llm"
)

// DescribeAIError turns a provider error into something to act on.
func DescribeAIError(p llm.Provider, err error) string {
	msg := err.Error()
	switch {
	case errors.Is(err, llm.ErrNoKey):
		return "Add an API key for " + p.Name + " in Settings › AI Assistant."
	case errors.Is(err, llm.ErrRefused):
		return "The model declined this request. Try rephrasing it."
	case strings.Contains(msg, "connection refused") && p.Kind != llm.Cloud:
		return p.Name + " isn't running at " + p.BaseURL + ". Start it, or change the URL."
	case strings.Contains(msg, "no such host"):
		return "Can't reach " + p.BaseURL + ": the host name doesn't resolve."
	}
	return msg
}
