package detect

import (
	"regexp"
	"unicode"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

var (
	urlRE  = regexp.MustCompile(`(?i)\b(?:https?://|www\.)\S+`)
	pathRE = regexp.MustCompile(`(?:^|\s)(?:~?/|\.{0,2}/)[^\s]+`)
	fileRE = regexp.MustCompile(`\b[A-Za-z0-9_-]+\.(?:md|txt|json|ya?ml|go|js|ts|py|rs|java|c|cpp|h|html|css)\b`)
)

type Stats struct {
	CJK, Latin, Kana, Hangul, Natural, LongestCJK int
}

// Target performs a local, deterministic direction decision. Chinese-dominant
// input goes to English; every other language goes to Simplified Chinese.
func Target(text string) model.Target {
	s := Analyze(text)
	if s.Hangul > 0 || s.Kana >= 2 || s.Kana > 0 && s.CJK <= 5 {
		return model.TargetZH
	}
	if s.CJK == 0 {
		return model.TargetZH
	}
	if s.Latin == 0 || s.CJK >= 4 && s.LongestCJK >= 3 {
		return model.TargetEN
	}
	ratio := float64(s.CJK) / float64(max(1, s.CJK+s.Latin))
	if s.CJK <= 2 && s.Latin >= 8 && ratio < .24 {
		return model.TargetZH
	}
	if s.LongestCJK >= 2 && ratio >= .25 || ratio >= .42 {
		return model.TargetEN
	}
	return model.TargetZH
}

func Analyze(text string) Stats {
	text = urlRE.ReplaceAllString(text, " ")
	text = pathRE.ReplaceAllString(text, " ")
	text = fileRE.ReplaceAllString(text, " ")
	var s Stats
	run := 0
	for _, r := range text {
		switch {
		case isCJK(r):
			s.CJK++
			s.Natural++
			run++
			if run > s.LongestCJK {
				s.LongestCJK = run
			}
		case unicode.Is(unicode.Latin, r) && unicode.IsLetter(r):
			s.Latin++
			s.Natural++
			run = 0
		case unicode.In(r, unicode.Hiragana, unicode.Katakana):
			s.Kana++
			s.Natural++
			run = 0
		case unicode.In(r, unicode.Hangul):
			s.Hangul++
			s.Natural++
			run = 0
		default:
			run = 0
		}
	}
	return s
}

func isCJK(r rune) bool {
	return unicode.In(r, unicode.Han) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x3134F)
}
