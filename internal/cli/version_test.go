package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

func TestResolveVersionInfo(t *testing.T) {
	t.Run("ldflags take priority", func(t *testing.T) {
		mockBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{Version: "v9.9.9"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "11111112222222"},
				},
			}, true
		}

		info := resolveVersionInfo("v1.2.3", "abcdef1", "2026-11-01T12:00:00Z", mockBuildInfo)
		if info.Version != "v1.2.3" {
			t.Errorf("got Version %q, want 'v1.2.3'", info.Version)
		}
		if info.GitCommit != "abcdef1" {
			t.Errorf("got GitCommit %q, want 'abcdef1'", info.GitCommit)
		}
		if info.BuildDate != "2026-11-01T12:00:00Z" {
			t.Errorf("got BuildDate %q, want '2026-11-01T12:00:00Z'", info.BuildDate)
		}
		if info.GoVersion != runtime.Version() {
			t.Errorf("got GoVersion %q, want %q", info.GoVersion, runtime.Version())
		}
	})

	t.Run("module version from debug info used when version is dev", func(t *testing.T) {
		mockBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{Version: "v0.8.13"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "9cc66fcedc10ecb33db365de55d96699d85f3847"},
					{Key: "vcs.time", Value: "2026-11-01T18:26:11Z"},
					{Key: "vcs.modified", Value: "false"},
				},
			}, true
		}

		info := resolveVersionInfo("dev", "", "", mockBuildInfo)
		if info.Version != "v0.8.13" {
			t.Errorf("got Version %q, want 'v0.8.13'", info.Version)
		}
		if info.GitCommit != "9cc66fc" {
			t.Errorf("got GitCommit %q, want '9cc66fc'", info.GitCommit)
		}
		if info.BuildDate != "2026-11-01T18:26:11Z" {
			t.Errorf("got BuildDate %q, want '2026-11-01T18:26:11Z'", info.BuildDate)
		}
	})

	t.Run("vcs revision with dirty suffix", func(t *testing.T) {
		mockBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abcdef123456"},
					{Key: "vcs.time", Value: "2026-11-02T10:00:00Z"},
					{Key: "vcs.modified", Value: "true"},
				},
			}, true
		}

		info := resolveVersionInfo("dev", "none", "unknown", mockBuildInfo)
		if info.Version != "dev" {
			t.Errorf("got Version %q, want 'dev'", info.Version)
		}
		if info.GitCommit != "abcdef1-dirty" {
			t.Errorf("got GitCommit %q, want 'abcdef1-dirty'", info.GitCommit)
		}
		if info.BuildDate != "2026-11-02T10:00:00Z" {
			t.Errorf("got BuildDate %q, want '2026-11-02T10:00:00Z'", info.BuildDate)
		}
	})

	t.Run("fallback to defaults when build info is unavailable", func(t *testing.T) {
		mockBuildInfo := func() (*debug.BuildInfo, bool) {
			return nil, false
		}

		info := resolveVersionInfo("", "", "", mockBuildInfo)
		if info.Version != "dev" {
			t.Errorf("got Version %q, want 'dev'", info.Version)
		}
		if info.GitCommit != "unknown" {
			t.Errorf("got GitCommit %q, want 'unknown'", info.GitCommit)
		}
		if info.BuildDate != "unknown" {
			t.Errorf("got BuildDate %q, want 'unknown'", info.BuildDate)
		}
	})
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestVersionCommand_Output(t *testing.T) {
	SetVersionInfo("v1.0.0-test", "1234567", "2026-11-02T09:00:00Z")

	t.Run("short output flag", func(t *testing.T) {
		versionShort = true
		versionJSON = false
		defer func() { versionShort = false }()

		out := captureOutput(func() {
			if err := versionCmd.RunE(versionCmd, nil); err != nil {
				t.Fatalf("versionCmd failed: %v", err)
			}
		})

		trimmed := strings.TrimSpace(out)
		if trimmed != "v1.0.0-test" {
			t.Errorf("got %q, want 'v1.0.0-test'", trimmed)
		}
	})

	t.Run("json output flag", func(t *testing.T) {
		versionShort = false
		versionJSON = true
		defer func() { versionJSON = false }()

		out := captureOutput(func() {
			if err := versionCmd.RunE(versionCmd, nil); err != nil {
				t.Fatalf("versionCmd failed: %v", err)
			}
		})

		var parsed VersionInfo
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("failed to parse json output: %v\nraw: %s", err, out)
		}
		if parsed.Version != "v1.0.0-test" {
			t.Errorf("got Version %q, want 'v1.0.0-test'", parsed.Version)
		}
		if parsed.GitCommit != "1234567" {
			t.Errorf("got GitCommit %q, want '1234567'", parsed.GitCommit)
		}
		if parsed.BuildDate != "2026-11-02T09:00:00Z" {
			t.Errorf("got BuildDate %q, want '2026-11-02T09:00:00Z'", parsed.BuildDate)
		}
	})

	t.Run("default LCARS output", func(t *testing.T) {
		versionShort = false
		versionJSON = false

		out := captureOutput(func() {
			if err := versionCmd.RunE(versionCmd, nil); err != nil {
				t.Fatalf("versionCmd failed: %v", err)
			}
		})

		if !strings.Contains(out, "v1.0.0-test") {
			t.Errorf("expected version in output, got: %s", out)
		}
		if !strings.Contains(out, "1234567") {
			t.Errorf("expected commit in output, got: %s", out)
		}
		if !strings.Contains(out, "2026-11-02T09:00:00Z") {
			t.Errorf("expected date in output, got: %s", out)
		}
	})
}
