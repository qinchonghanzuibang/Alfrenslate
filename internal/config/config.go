package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultOrder = "deepseek,openai-compatible,deepl,youdao,baidu,google,microsoft"

type Config struct {
	Order        []string
	Timeout      time.Duration
	Retries      int
	CacheEnabled bool
	CacheTTL     time.Duration
	CacheDir     string
	Debug        bool
}

func Load() Config {
	seconds := envInt("REQUEST_TIMEOUT_SECONDS", 8, 1, 60)
	hours := envInt("CACHE_TTL_HOURS", 24, 1, 24*365)
	dir := strings.TrimSpace(os.Getenv("alfred_workflow_cache"))
	if dir == "" {
		dir = strings.TrimSpace(os.Getenv("ALFRENSLATE_CACHE_DIR"))
	}
	order := strings.Split(env("PROVIDER_ORDER", DefaultOrder), ",")
	for i := range order {
		order[i] = strings.TrimSpace(order[i])
	}
	return Config{
		Order: order, Timeout: time.Duration(seconds) * time.Second,
		Retries: envInt("RETRY_COUNT", 1, 0, 5), CacheEnabled: envBool("CACHE_ENABLED", true),
		CacheTTL: time.Duration(hours) * time.Hour, CacheDir: dir, Debug: envBool("DEBUG_LOGGING", false),
	}
}

func env(k, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return fallback
}

func envBool(k string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func envInt(k string, fallback, min, max int) int {
	n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(k)))
	if err != nil || n < min || n > max {
		return fallback
	}
	return n
}

func Enabled(id string) bool {
	return envBool(strings.ToUpper(strings.ReplaceAll(id, "-", "_"))+"_ENABLED", false)
}

func Value(key string) string { return strings.TrimSpace(os.Getenv(key)) }
