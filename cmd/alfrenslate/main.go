package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"unicode"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/alfred"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/cache"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/detect"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/baidu"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/deepl"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/deepseek"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/google"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/microsoft"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/openaicompatible"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers/youdao"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/security"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/translate"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/version"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, security.Redact(err.Error(), credentialValues()...))
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "translate":
		return translateCmd(ctx, args[1:])
	case "detect":
		return detectCmd(args[1:])
	case "doctor":
		return doctorCmd()
	case "test-providers":
		return testCmd(ctx, "")
	case "test-provider":
		if len(args) < 2 {
			return errors.New("provider ID is required")
		}
		return testCmd(ctx, args[1])
	case "clear-cache":
		return clearCache()
	case "manage":
		return manageCmd()
	case "action":
		if len(args) < 2 {
			return errors.New("action is required")
		}
		return actionCmd(ctx, args[1])
	case "version", "--version", "-v":
		fmt.Println(version.Version)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
func usageError() error {
	return errors.New("usage: alfrenslate <translate|detect|doctor|test-providers|test-provider|clear-cache|version>")
}

type stringFlag struct {
	value string
	set   bool
}

func (f *stringFlag) String() string     { return f.value }
func (f *stringFlag) Set(v string) error { f.value = v; f.set = true; return nil }
func translateCmd(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("translate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var text string
	var to stringFlag
	var asAlfred bool
	fs.StringVar(&text, "text", "", "text to translate")
	fs.Var(&to, "to", "target language: zh or en")
	fs.Var(&to, "t", "target language: zh or en")
	fs.BoolVar(&asAlfred, "alfred", false, "emit Alfred Script Filter JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		if text != "" {
			return outputError(asAlfred, "Unexpected arguments", strings.Join(rest, " "))
		}
		text = strings.Join(rest, " ")
	}
	target, clean, err := parseDirection(text, to)
	if err != nil {
		return outputError(asAlfred, "Invalid translation options", err.Error())
	}
	text = clean
	if strings.TrimSpace(text) == "" {
		if asAlfred {
			return emit(alfred.Hint("Type text to translate", "Examples: ts hello · ts 你好"))
		}
		return errors.New("translation text is required")
	}
	cfg := config.Load()
	all := allProviders(cfg)
	enabled := enabledProviders(cfg, all)
	if len(enabled) == 0 {
		if asAlfred {
			return emit(alfred.Hint("No provider is enabled", "Open Workflow Configuration and enable at least one configured provider"))
		}
		return errors.New("no provider is enabled")
	}
	store := cache.New(cfg.CacheDir, cfg.CacheTTL, cfg.CacheEnabled)
	engine := translate.Engine{Providers: enabled, Cache: store, Timeout: cfg.Timeout}
	results := engine.Translate(ctx, model.Request{Text: text, Target: target})
	if asAlfred {
		return emit(alfred.TranslationItems(results, all))
	}
	failed := false
	succeeded := 0
	for _, r := range results {
		if r.Err != nil {
			failed = true
			fmt.Fprintf(os.Stderr, "%s: %s\n", r.Provider, security.Redact(r.Err.Error(), credentialValues()...))
		} else {
			succeeded++
			fmt.Printf("%s\t%s\n", r.Provider, r.Text)
		}
	}
	if failed && succeeded == 0 {
		return errors.New("translation failed")
	}
	return nil
}
func parseDirection(text string, to stringFlag) (model.Target, string, error) {
	target := model.Target("")
	if to.set {
		switch strings.ToLower(to.value) {
		case "zh", "zh-hans", "zh-cn":
			target = model.TargetZH
		case "en":
			target = model.TargetEN
		default:
			return "", text, errors.New("target must be zh or en")
		}
	}
	trim := strings.TrimSpace(text)
	if strings.HasPrefix(trim, "--to ") || strings.HasPrefix(trim, "-t ") {
		optionEnd := strings.IndexFunc(trim, unicode.IsSpace)
		remainder := strings.TrimLeftFunc(trim[optionEnd:], unicode.IsSpace)
		targetEnd := strings.IndexFunc(remainder, unicode.IsSpace)
		if targetEnd < 0 {
			return "", text, errors.New("use --to zh <text> or --to en <text>")
		}
		targetName := remainder[:targetEnd]
		switch strings.ToLower(targetName) {
		case "zh":
			target = model.TargetZH
		case "en":
			target = model.TargetEN
		default:
			return "", text, fmt.Errorf("unsupported target %q; use zh or en", targetName)
		}
		trim = strings.TrimLeftFunc(remainder[targetEnd:], unicode.IsSpace)
		if trim == "" {
			return "", text, errors.New("translation text is required after the target language")
		}
	} else if strings.HasPrefix(trim, "-") {
		return "", text, errors.New("unknown option; supported options are --to zh, --to en, -t zh, and -t en")
	}
	if target == "" {
		target = detect.Target(trim)
	}
	return target, trim, nil
}
func outputError(jsonMode bool, title, message string) error {
	if jsonMode {
		return emit(alfred.Hint(title, message))
	}
	return errors.New(message)
}
func detectCmd(args []string) error {
	fs := flag.NewFlagSet("detect", flag.ContinueOnError)
	var text string
	fs.StringVar(&text, "text", "", "text to inspect")
	if e := fs.Parse(args); e != nil {
		return e
	}
	if text == "" {
		text = strings.Join(fs.Args(), " ")
	}
	if strings.TrimSpace(text) == "" {
		return errors.New("text is required")
	}
	fmt.Println(detect.Target(text))
	return nil
}
func allProviders(cfg config.Config) []providers.Provider {
	client := providers.NewClient(cfg.Timeout, cfg.Retries)
	list := make([]providers.Provider, 0, 7)
	if p, e := deepseek.New(client); e == nil {
		list = append(list, p)
	} else {
		list = append(list, &invalid{id: "deepseek", name: "DeepSeek", err: e})
	}
	if p, e := openaicompatible.New(client); e == nil {
		list = append(list, p)
	} else {
		list = append(list, &invalid{id: "openai-compatible", name: "OpenAI-Compatible", err: e})
	}
	list = append(list, deepl.New(client), youdao.New(client), baidu.New(client), google.New(client))
	if p, e := microsoft.New(client); e == nil {
		list = append(list, p)
	} else {
		list = append(list, &invalid{id: "microsoft", name: "Microsoft Translator", err: e})
	}
	return providers.Ordered(cfg.Order, list)
}

type invalid struct {
	id, name string
	err      error
}

func (p *invalid) ID() string            { return p.id }
func (p *invalid) Name() string          { return p.name }
func (p *invalid) Configured() error     { return p.err }
func (p *invalid) CacheIdentity() string { return "invalid" }
func (p *invalid) Translate(context.Context, model.Request) (model.Result, error) {
	return model.Result{}, p.err
}
func enabledProviders(cfg config.Config, all []providers.Provider) []providers.Provider {
	out := make([]providers.Provider, 0, len(all))
	for _, p := range all {
		if config.Enabled(p.ID()) {
			out = append(out, p)
		}
	}
	return out
}
func doctorCmd() error {
	cfg := config.Load()
	fmt.Printf("Alfrenslate %s\n", version.Version)
	fmt.Printf("Cache: %s (TTL %s)\n", map[bool]string{true: "Enabled", false: "Disabled"}[cfg.CacheEnabled], cfg.CacheTTL)
	fmt.Printf("Timeout: %s · Retries: %d\n", cfg.Timeout, cfg.Retries)
	for _, p := range allProviders(cfg) {
		state := "Disabled"
		if config.Enabled(p.ID()) {
			state = "Enabled"
		}
		configured := "Configured"
		if p.Configured() != nil {
			configured = "Missing credentials"
		}
		fmt.Printf("%-24s %s · %s\n", p.Name(), state, configured)
	}
	fmt.Println("Reachability is checked only by test-provider(s) to avoid paid requests during doctor.")
	return nil
}
func testCmd(ctx context.Context, only string) error {
	cfg := config.Load()
	all := allProviders(cfg)
	selected := []providers.Provider{}
	for _, p := range all {
		if only != "" && p.ID() != only {
			continue
		}
		if only == "" && !config.Enabled(p.ID()) {
			continue
		}
		selected = append(selected, p)
	}
	if len(selected) == 0 {
		return errors.New("no matching enabled provider")
	}
	engine := translate.Engine{Providers: selected, Cache: cache.New("", 0, false), Timeout: cfg.Timeout}
	results := engine.Translate(ctx, model.Request{Text: "Hello, this is an Alfrenslate provider test.", Target: model.TargetZH})
	failed := false
	for _, r := range results {
		if r.Err != nil {
			failed = true
			fmt.Printf("%s: Failed · %s\n", r.Provider, security.Redact(r.Err.Error(), credentialValues()...))
		} else {
			fmt.Printf("%s: Reachable · %d ms\n", r.Provider, r.Elapsed.Milliseconds())
		}
	}
	if failed {
		return errors.New("one or more provider tests failed")
	}
	return nil
}
func clearCache() error {
	cfg := config.Load()
	s := cache.New(cfg.CacheDir, cfg.CacheTTL, true)
	if e := s.Clear(); e != nil {
		return e
	}
	fmt.Println("Alfrenslate cache cleared.")
	return nil
}
func manageCmd() error {
	cfg := config.Load()
	items := []alfred.Item{
		{UID: "version", Title: "Alfrenslate " + version.Version, Subtitle: "Translation at Alfred speed.", Valid: false},
		{UID: "config", Title: "Open Workflow Configuration", Subtitle: "Configure providers, timeout, ordering, and local cache", Arg: "open-config", Valid: true},
		{UID: "test-all", Title: "Test all enabled providers", Subtitle: "Sends a small live translation request to each enabled provider", Arg: "test:all", Valid: true},
		{UID: "clear-cache", Title: "Clear local translation cache", Subtitle: "Deletes only Alfrenslate translation cache entries", Arg: "clear-cache", Valid: true},
	}
	for _, p := range allProviders(cfg) {
		enabled := "Disabled"
		if config.Enabled(p.ID()) {
			enabled = "Enabled"
		}
		configured := "Configured"
		if p.Configured() != nil {
			configured = "Missing credentials"
		}
		items = append(items, alfred.Item{UID: "provider-" + p.ID(), Title: p.Name() + " · " + enabled + " · " + configured, Subtitle: "Press Enter to test this provider", Arg: "test:" + p.ID(), Valid: config.Enabled(p.ID())})
	}
	items = append(items,
		alfred.Item{UID: "doctor", Title: "View local diagnostics", Subtitle: "Shows version, configuration state, cache, timeout, and architecture", Arg: "doctor", Valid: true},
		alfred.Item{UID: "github", Title: "Open GitHub repository", Arg: "github", Valid: true},
		alfred.Item{UID: "releases", Title: "Open GitHub Releases", Arg: "releases", Valid: true},
		alfred.Item{UID: "docs", Title: "Open README documentation", Arg: "docs", Valid: true},
	)
	return emit(alfred.Response{Items: items})
}
func actionCmd(ctx context.Context, a string) error {
	switch a {
	case "open-config":
		return openURL("alfredpreferences://navigateto/workflows>workflow>com.chonghanqin.alfrenslate")
	case "github":
		return openURL("https://github.com/qinchonghanzuibang/Alfrenslate")
	case "releases":
		return openURL("https://github.com/qinchonghanzuibang/Alfrenslate/releases")
	case "docs":
		return openURL("https://github.com/qinchonghanzuibang/Alfrenslate#readme")
	case "doctor":
		return doctorCmd()
	case "clear-cache":
		return clearCache()
	case "test:all":
		return testCmd(ctx, "")
	}
	if strings.HasPrefix(a, "test:") {
		return testCmd(ctx, strings.TrimPrefix(a, "test:"))
	}
	return errors.New("unknown management action")
}
func openURL(u string) error { return exec.Command("/usr/bin/open", u).Run() }
func emit(v any) error       { return json.NewEncoder(os.Stdout).Encode(v) }
func credentialValues() []string {
	keys := []string{"DEEPSEEK_API_KEY", "OPENAI_COMPATIBLE_API_KEY", "DEEPL_API_KEY", "YOUDAO_APP_KEY", "YOUDAO_APP_SECRET", "BAIDU_APP_ID", "BAIDU_APP_SECRET", "GOOGLE_API_KEY", "MICROSOFT_API_KEY"}
	sort.Strings(keys)
	out := []string{}
	for _, k := range keys {
		if v := config.Value(k); v != "" {
			out = append(out, v)
		}
	}
	return out
}
