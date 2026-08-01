package detect

import (
	"strings"
	"testing"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

func TestTarget(t *testing.T) {
	tests := []struct {
		name, text string
		want       model.Target
	}{
		{"english", "hello world", model.TargetZH}, {"simplified", "你好世界", model.TargetEN},
		{"traditional", "這是一個測試", model.TargetEN}, {"word", "hello", model.TargetZH},
		{"short chinese", "你好", model.TargetEN}, {"acronym", "LLM", model.TargetZH},
		{"chinese acronym", "具身 AI", model.TargetEN}, {"file chinese", "README.md 怎么写", model.TargetEN},
		{"mixed chinese", "Use Qwen3 来测试这个方法", model.TargetEN}, {"mixed english", "The result 在 layer 18 is stable", model.TargetZH},
		{"english mixed", "How to train 一个 VLA model", model.TargetZH}, {"url chinese", "https://example.com 请检查", model.TargetEN},
		{"url english", "https://example.com hello", model.TargetZH}, {"path chinese", "/tmp/a.txt 怎么打开", model.TargetEN},
		{"markdown", "**hello** [docs](https://example.com)", model.TargetZH}, {"latex", `$E=mc^2$`, model.TargetZH},
		{"code", `fmt.Println("hello")`, model.TargetZH}, {"chinese comment", `fmt.Println(x) // 输出结果`, model.TargetEN},
		{"english comment", `x := 1 // return result`, model.TargetZH}, {"numbers", "12345", model.TargetZH},
		{"punctuation", "...?!", model.TargetZH}, {"emoji", "🧪🚀", model.TargetZH}, {"empty", "", model.TargetZH},
		{"japanese", "日本語を翻訳", model.TargetZH}, {"korean", "번역 테스트", model.TargetZH},
		{"paragraphs", "第一段内容。\n\nSecond section.", model.TargetEN}, {"long", strings.Repeat("这是中文。", 5000), model.TargetEN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Target(tt.text); got != tt.want {
				t.Fatalf("Target(%q)=%s want %s; stats=%+v", tt.text, got, tt.want, Analyze(tt.text))
			}
		})
	}
}

func TestAnalyzeIgnoresTechnicalNoise(t *testing.T) {
	s := Analyze("https://例子.测试 /路径/文件 README.md 🧪")
	if s.Natural != 0 {
		t.Fatalf("technical noise counted as natural language: %+v", s)
	}
}
