//go:build live

package main

import (
	"context"
	"os"
	"testing"
)

func TestLiveEnabledProviders(t *testing.T) {
	if os.Getenv("ALFRENSLATE_LIVE_TEST") != "1" {
		t.Skip("set ALFRENSLATE_LIVE_TEST=1 to run paid live provider tests")
	}
	if err := testCmd(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
}
