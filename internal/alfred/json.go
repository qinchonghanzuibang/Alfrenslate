package alfred

import (
	"fmt"
	"strings"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Response struct {
	Items []Item `json:"items"`
}
type Item struct {
	UID       string            `json:"uid,omitempty"`
	Title     string            `json:"title"`
	Subtitle  string            `json:"subtitle,omitempty"`
	Arg       string            `json:"arg,omitempty"`
	Valid     bool              `json:"valid"`
	Icon      *Icon             `json:"icon,omitempty"`
	Mods      map[string]Mod    `json:"mods,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}
type Icon struct {
	Path string `json:"path"`
}
type Mod struct {
	Valid    bool   `json:"valid"`
	Arg      string `json:"arg,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
}

func TranslationItems(results []model.Result, all []providers.Provider) Response {
	names := map[string]string{}
	for _, p := range all {
		names[p.ID()] = p.Name()
	}
	items := make([]Item, 0, len(results))
	for _, r := range results {
		name := names[r.Provider]
		if name == "" {
			name = r.Provider
		}
		if r.Err != nil {
			items = append(items, Item{UID: "error-" + r.Provider, Title: name + " failed", Subtitle: safeError(r.Err.Error()), Valid: false})
			continue
		}
		elapsed := fmt.Sprintf("%d ms", r.Elapsed.Milliseconds())
		if r.Cached {
			elapsed = "cached"
		}
		source := r.Source
		if source == "" {
			source = "AUTO"
		}
		subtitle := fmt.Sprintf("%s · %s → %s · %s", name, strings.ToUpper(source), r.Target.Label(), elapsed)
		items = append(items, Item{UID: "translation-" + r.Provider, Title: r.Text, Subtitle: subtitle, Arg: r.Text, Valid: true, Mods: map[string]Mod{"cmd": {Valid: true, Arg: r.Text, Subtitle: "Show with Alfred Large Type"}, "alt": {Valid: true, Arg: r.Text, Subtitle: "Copy and paste into the frontmost app"}}})
	}
	return Response{Items: items}
}
func safeError(s string) string {
	if len(s) > 180 {
		s = s[:180] + "…"
	}
	return s + " · Open Workflow Configuration or run alfrenslate diagnostics"
}
func Hint(title, subtitle string) Response {
	return Response{Items: []Item{{Title: title, Subtitle: subtitle, Valid: false}}}
}
