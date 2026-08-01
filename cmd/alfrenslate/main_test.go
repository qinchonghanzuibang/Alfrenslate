package main

import (
	"testing"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

func TestParseDirection(t *testing.T) {
	tests := []struct {
		text  string
		flag  stringFlag
		want  model.Target
		clean string
		err   bool
	}{{"hello", stringFlag{}, model.TargetZH, "hello", false}, {"你好", stringFlag{}, model.TargetEN, "你好", false}, {"--to zh 你好", stringFlag{}, model.TargetZH, "你好", false}, {"--to en Hello", stringFlag{}, model.TargetEN, "Hello", false}, {"-t zh hello", stringFlag{}, model.TargetZH, "hello", false}, {"-t en 你好", stringFlag{}, model.TargetEN, "你好", false}, {"--to en First paragraph\n\nSecond paragraph", stringFlag{}, model.TargetEN, "First paragraph\n\nSecond paragraph", false}, {"hello", stringFlag{value: "en", set: true}, model.TargetEN, "hello", false}, {"--to fr hello", stringFlag{}, "", "", true}, {"--bad hello", stringFlag{}, "", "", true}}
	for _, x := range tests {
		got, clean, e := parseDirection(x.text, x.flag)
		if (e != nil) != x.err || !x.err && (got != x.want || clean != x.clean) {
			t.Errorf("parse(%q)=%s,%q,%v want %s,%q err=%v", x.text, got, clean, e, x.want, x.clean, x.err)
		}
	}
}
