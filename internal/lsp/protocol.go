// Package lsp talks to language servers: it starts one per language for
// a project, keeps it told about open documents, and asks it for
// completions (with the edits that come with them, such as imports).
package lsp

import "encoding/json"

// Position is a place in a document: a 0-based line and a character
// offset in UTF-16 code units, as the protocol counts.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is the text between two positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// TextEdit replaces a range with new text.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// CompletionKind is what a completion is: a function, a variable, …
type CompletionKind int

// The completion kinds the protocol names.
const (
	KindText CompletionKind = iota + 1
	KindMethod
	KindFunction
	KindConstructor
	KindField
	KindVariable
	KindClass
	KindInterface
	KindModule
	KindProperty
	KindUnit
	KindValue
	KindEnum
	KindKeyword
	KindSnippet
	KindColor
	KindFile
	KindReference
	KindFolder
	KindEnumMember
	KindConstant
	KindStruct
	KindEvent
	KindOperator
	KindTypeParameter
)

// insertSnippet is the insert text format of snippets ("$1", "${2:x}").
const insertSnippet = 2

// Completion is one suggestion. Edit, when set, says what to replace;
// otherwise the word before the caret gets Insert. Extra edits (an
// import) go with it.
type Completion struct {
	Label    string
	Detail   string
	Kind     CompletionKind
	SortText string
	Filter   string
	Insert   string
	Snippet  bool
	Edit     *TextEdit
	Extra    []TextEdit
	raw      json.RawMessage // as the server sent it, for resolving
}

// wireCompletion is a completion as it comes over the wire.
type wireCompletion struct {
	Label            string          `json:"label"`
	Kind             CompletionKind  `json:"kind"`
	Detail           string          `json:"detail"`
	SortText         string          `json:"sortText"`
	FilterText       string          `json:"filterText"`
	InsertText       string          `json:"insertText"`
	InsertTextFormat int             `json:"insertTextFormat"`
	TextEdit         json.RawMessage `json:"textEdit"`
	AdditionalEdits  []TextEdit      `json:"additionalTextEdits"`
	LabelDetails     *struct {
		Detail      string `json:"detail"`
		Description string `json:"description"`
	} `json:"labelDetails"`
}

// decode reads a completion item, keeping its JSON for resolving.
func decode(raw json.RawMessage) (Completion, error) {
	var w wireCompletion
	if err := json.Unmarshal(raw, &w); err != nil {
		return Completion{}, err
	}
	c := Completion{
		Label: w.Label, Detail: w.Detail, Kind: w.Kind, SortText: w.SortText, Filter: w.FilterText,
		Insert: w.InsertText, Snippet: w.InsertTextFormat == insertSnippet, Extra: w.AdditionalEdits, raw: raw,
	}
	if c.Detail == "" && w.LabelDetails != nil {
		c.Detail = w.LabelDetails.Description
		if c.Detail == "" {
			c.Detail = w.LabelDetails.Detail
		}
	}
	if len(w.TextEdit) > 0 && string(w.TextEdit) != "null" {
		// A TextEdit, or an InsertReplaceEdit: use its insert range.
		var e struct {
			Range   *Range `json:"range"`
			Insert  *Range `json:"insert"`
			NewText string `json:"newText"`
		}
		if json.Unmarshal(w.TextEdit, &e) == nil {
			r := e.Range
			if r == nil {
				r = e.Insert
			}
			if r != nil {
				c.Edit = &TextEdit{Range: *r, NewText: e.NewText}
			}
		}
	}
	if c.Insert == "" {
		c.Insert = c.Label
	}
	if c.Filter == "" {
		c.Filter = c.Label
	}
	return c, nil
}
