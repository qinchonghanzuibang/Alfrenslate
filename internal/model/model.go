package model

import "time"

type Target string

const (
	TargetZH Target = "zh"
	TargetEN Target = "en"
)

func (t Target) Label() string {
	if t == TargetEN {
		return "EN"
	}
	return "ZH-HANS"
}

type Request struct {
	Text   string
	Target Target
}

type Result struct {
	Provider string        `json:"provider"`
	Text     string        `json:"text,omitempty"`
	Source   string        `json:"source,omitempty"`
	Target   Target        `json:"target"`
	Elapsed  time.Duration `json:"-"`
	Cached   bool          `json:"cached,omitempty"`
	Err      error         `json:"-"`
}
