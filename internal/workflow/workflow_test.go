package workflow_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type workflow struct {
	Version                string                  `json:"version"`
	UserConfiguration      []configurationItem     `json:"userconfigurationconfig"`
	VariablesDoNotExport   []string                `json:"variablesdontexport"`
	Objects                []workflowObject        `json:"objects"`
	Connections            map[string][]connection `json:"connections"`
	UserInterfacePositions map[string]any          `json:"uidata"`
}

type configurationItem struct {
	Config      map[string]any `json:"config"`
	Description string         `json:"description"`
	Label       string         `json:"label"`
	Type        string         `json:"type"`
	Variable    string         `json:"variable"`
}

type workflowObject struct {
	Config  map[string]any `json:"config"`
	Type    string         `json:"type"`
	UID     string         `json:"uid"`
	Version int            `json:"version"`
}

type connection struct {
	DestinationUID string `json:"destinationuid"`
	Modifiers      int    `json:"modifiers"`
	VetoClose      bool   `json:"vitoclose"`
}

func loadWorkflow(t *testing.T) workflow {
	t.Helper()
	path := filepath.Join("..", "..", "workflow", "info.plist")
	output, err := exec.Command("/usr/bin/plutil", "-convert", "json", "-o", "-", path).Output()
	if err != nil {
		t.Fatalf("parse %s with plutil: %v", path, err)
	}
	var result workflow
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode parsed workflow plist: %v", err)
	}
	return result
}

func TestWorkflowConfigurationSchema(t *testing.T) {
	wf := loadWorkflow(t)
	if wf.Version != "1.0.1" {
		t.Fatalf("workflow version = %q, want 1.0.1", wf.Version)
	}
	if len(wf.UserConfiguration) == 0 {
		t.Fatal("userconfigurationconfig must not be empty")
	}

	expectedVariables := []string{
		"BAIDU_APP_ID", "BAIDU_APP_SECRET", "BAIDU_ENABLED",
		"CACHE_ENABLED", "CACHE_TTL_HOURS", "DEBUG_LOGGING",
		"DEEPL_API_KEY", "DEEPL_API_TIER", "DEEPL_BASE_URL", "DEEPL_ENABLED",
		"DEEPSEEK_API_KEY", "DEEPSEEK_BASE_URL", "DEEPSEEK_ENABLED", "DEEPSEEK_MODEL",
		"GOOGLE_API_KEY", "GOOGLE_ENABLED", "MICROSOFT_API_KEY", "MICROSOFT_ENABLED",
		"MICROSOFT_ENDPOINT", "MICROSOFT_REGION", "OPENAI_COMPATIBLE_API_KEY",
		"OPENAI_COMPATIBLE_BASE_URL", "OPENAI_COMPATIBLE_ENABLED", "OPENAI_COMPATIBLE_MODEL",
		"PROVIDER_ORDER", "REQUEST_TIMEOUT_SECONDS", "RETRY_COUNT", "YOUDAO_APP_KEY",
		"YOUDAO_APP_SECRET", "YOUDAO_DOMAIN", "YOUDAO_ENABLED",
	}
	allowedTypes := map[string]bool{"checkbox": true, "popupbutton": true, "textfield": true}
	allowedSections := map[string]bool{
		"ADVANCED": true, "BAIDU": true, "DEEPL": true, "DEEPSEEK": true,
		"GENERAL": true, "GOOGLE": true, "MICROSOFT": true,
		"OPENAI-COMPATIBLE": true, "YOUDAO": true,
	}
	seen := make(map[string]configurationItem, len(wf.UserConfiguration))
	for i, item := range wf.UserConfiguration {
		if item.Label == "" || item.Description == "" || item.Variable == "" || item.Config == nil {
			t.Fatalf("configuration item %d has metadata at the wrong level: %#v", i, item)
		}
		if !allowedTypes[item.Type] {
			t.Fatalf("configuration item %q uses unsupported Alfred type %q", item.Label, item.Type)
		}
		if _, duplicate := seen[item.Variable]; duplicate {
			t.Fatalf("duplicate configuration variable %q", item.Variable)
		}
		seen[item.Variable] = item
		section := strings.TrimSpace(strings.SplitN(item.Description, "·", 2)[0])
		if !allowedSections[section] {
			t.Errorf("configuration item %q has no recognized section prefix", item.Label)
		}
		for _, misplaced := range []string{"description", "label", "variable"} {
			if _, exists := item.Config[misplaced]; exists {
				t.Errorf("configuration item %q has %q inside config", item.Label, misplaced)
			}
		}
	}

	actualVariables := make([]string, 0, len(seen))
	for variable := range seen {
		actualVariables = append(actualVariables, variable)
	}
	sort.Strings(actualVariables)
	if !reflect.DeepEqual(actualVariables, expectedVariables) {
		t.Fatalf("configuration variables mismatch\nactual:   %v\nexpected: %v", actualVariables, expectedVariables)
	}

	checkboxes := []string{
		"BAIDU_ENABLED", "CACHE_ENABLED", "DEBUG_LOGGING", "DEEPL_ENABLED",
		"DEEPSEEK_ENABLED", "GOOGLE_ENABLED", "MICROSOFT_ENABLED",
		"OPENAI_COMPATIBLE_ENABLED", "YOUDAO_ENABLED",
	}
	for _, variable := range checkboxes {
		if seen[variable].Type != "checkbox" {
			t.Errorf("%s must be a checkbox, got %s", variable, seen[variable].Type)
		}
	}
	for _, variable := range []string{"DEEPL_API_TIER", "YOUDAO_DOMAIN"} {
		item := seen[variable]
		if item.Type != "popupbutton" {
			t.Errorf("%s must be a popupbutton, got %s", variable, item.Type)
		}
		validatePopup(t, item)
	}
}

func TestWorkflowSecretsHaveNoDefaultsAndDoNotExport(t *testing.T) {
	wf := loadWorkflow(t)
	items := make(map[string]configurationItem, len(wf.UserConfiguration))
	for _, item := range wf.UserConfiguration {
		items[item.Variable] = item
	}
	doNotExport := make(map[string]bool, len(wf.VariablesDoNotExport))
	for _, variable := range wf.VariablesDoNotExport {
		doNotExport[variable] = true
	}
	secretVariables := []string{
		"BAIDU_APP_ID", "BAIDU_APP_SECRET", "DEEPL_API_KEY", "DEEPSEEK_API_KEY",
		"GOOGLE_API_KEY", "MICROSOFT_API_KEY", "OPENAI_COMPATIBLE_API_KEY",
		"YOUDAO_APP_KEY", "YOUDAO_APP_SECRET",
	}
	for _, variable := range secretVariables {
		if value, ok := items[variable].Config["default"].(string); !ok || value != "" {
			t.Errorf("%s must have an empty default", variable)
		}
		if !doNotExport[variable] {
			t.Errorf("%s must be listed in variablesdontexport", variable)
		}
	}
}

func TestUniversalActionUsesTranslationScriptFilter(t *testing.T) {
	wf := loadWorkflow(t)
	objects := make(map[string]workflowObject, len(wf.Objects))
	var universalActions []workflowObject
	for _, object := range wf.Objects {
		if object.UID == "" {
			t.Fatal("workflow object has an empty UID")
		}
		if _, duplicate := objects[object.UID]; duplicate {
			t.Fatalf("duplicate workflow object UID %q", object.UID)
		}
		objects[object.UID] = object
		if object.Type == "alfred.workflow.trigger.universalaction" {
			universalActions = append(universalActions, object)
		}
	}
	if len(universalActions) != 1 {
		t.Fatalf("expected one Universal Action, got %d", len(universalActions))
	}
	action := universalActions[0]
	if action.Config["name"] != "Translate with Alfrenslate" {
		t.Errorf("unexpected Universal Action name: %v", action.Config["name"])
	}
	for key, expected := range map[string]bool{
		"acceptsfiles": false, "acceptsmulti": false, "acceptstext": true, "acceptsurls": false,
	} {
		if actual, ok := action.Config[key].(bool); !ok || actual != expected {
			t.Errorf("Universal Action %s = %v, want %v", key, action.Config[key], expected)
		}
	}
	connections := wf.Connections[action.UID]
	if len(connections) != 1 {
		t.Fatalf("Universal Action must have one default connection, got %d", len(connections))
	}
	destination := objects[connections[0].DestinationUID]
	if destination.Type != "alfred.workflow.input.scriptfilter" {
		t.Fatalf("Universal Action destination is %q, want Script Filter", destination.Type)
	}
	if destination.Config["keyword"] != "ts" {
		t.Errorf("Universal Action does not reuse the ts Script Filter")
	}
	if destination.Config["skipuniversalaction"] != true {
		t.Errorf("ts Script Filter must opt out of Alfred's implicit Universal Actions to avoid a duplicate")
	}
	if destination.Config["script"] != `exec "./run.sh" translate --alfred --text "$1"` {
		t.Errorf("Universal Action destination does not reuse the translation CLI: %v", destination.Config["script"])
	}
	for _, name := range []string{"README.md", "README.zh-CN.md"} {
		contents, err := os.ReadFile(filepath.Join("..", "..", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(contents), "Translate with Alfrenslate") {
			t.Errorf("%s does not document the installed Universal Action name", name)
		}
	}
}

func TestWorkflowConnectionGraph(t *testing.T) {
	wf := loadWorkflow(t)
	objects := make(map[string]bool, len(wf.Objects))
	for _, object := range wf.Objects {
		objects[object.UID] = true
	}
	for source, connections := range wf.Connections {
		if !objects[source] {
			t.Errorf("connection source %q does not exist", source)
		}
		for _, connection := range connections {
			if !objects[connection.DestinationUID] {
				t.Errorf("connection %q -> %q has a missing destination", source, connection.DestinationUID)
			}
		}
	}
	for uid := range wf.UserInterfacePositions {
		if !objects[uid] {
			t.Errorf("uidata references missing object %q", uid)
		}
	}
}

func validatePopup(t *testing.T, item configurationItem) {
	t.Helper()
	pairs, ok := item.Config["pairs"].([]any)
	if !ok || len(pairs) == 0 {
		t.Fatalf("%s must define popup pairs", item.Variable)
	}
	values := make(map[string]bool, len(pairs))
	for _, rawPair := range pairs {
		pair, ok := rawPair.([]any)
		if !ok || len(pair) != 2 {
			t.Fatalf("%s has an invalid popup pair: %#v", item.Variable, rawPair)
		}
		label, labelOK := pair[0].(string)
		value, valueOK := pair[1].(string)
		if !labelOK || !valueOK || label == "" || value == "" {
			t.Fatalf("%s has an empty popup pair: %#v", item.Variable, rawPair)
		}
		values[value] = true
	}
	defaultValue, ok := item.Config["default"].(string)
	if !ok || !values[defaultValue] {
		t.Errorf("%s default %q is not in popup options", item.Variable, defaultValue)
	}
}
