package main

import (
	"strings"
	"testing"
	"time"
)

func TestIsTruthyEnv(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "true", in: "true", want: true},
		{name: "uppercase true", in: "TRUE", want: true},
		{name: "trimmed true", in: "  true  ", want: true},
		{name: "one", in: "1", want: true},
		{name: "yes", in: "yes", want: true},
		{name: "on", in: "on", want: true},
		{name: "false", in: "false", want: false},
		{name: "empty", in: "", want: false},
		{name: "random", in: "maybe", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTruthyEnv(tt.in)
			if got != tt.want {
				t.Fatalf("isTruthyEnv(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoadAppConfigFromEnv(t *testing.T) {
	t.Setenv("LIBSQL_DATABASE_URL", "libsql://example.turso.io")
	t.Setenv("LIBSQL_AUTH_TOKEN", "secret-token")
	t.Setenv("LIBSQL_SYNC_INTERVAL", "15")
	t.Setenv("LIBSQL_REQUIRE_REMOTE", "true")

	cfg := loadAppConfigFromEnv()

	if cfg.libsqlURL != "libsql://example.turso.io" {
		t.Fatalf("libsqlURL = %q", cfg.libsqlURL)
	}
	if cfg.libsqlToken != "secret-token" {
		t.Fatalf("libsqlToken = %q", cfg.libsqlToken)
	}
	if cfg.syncInterval != 15*time.Second {
		t.Fatalf("syncInterval = %v", cfg.syncInterval)
	}
	if !cfg.requireRemote {
		t.Fatalf("requireRemote = false, want true")
	}
}

func TestLoadAppConfigFromEnvUsesDefaultSyncIntervalForInvalidValue(t *testing.T) {
	t.Setenv("LIBSQL_SYNC_INTERVAL", "not-a-number")

	cfg := loadAppConfigFromEnv()

	if cfg.syncInterval != 60*time.Second {
		t.Fatalf("syncInterval = %v, want 60s", cfg.syncInterval)
	}
}

func TestValidateAppConfigAllowsLocalWhenRemoteNotRequired(t *testing.T) {
	cfg := appConfig{requireRemote: false}

	if err := validateAppConfig(cfg); err != nil {
		t.Fatalf("validateAppConfig() error = %v, want nil", err)
	}
}

func TestValidateAppConfigRequiresURLWhenStrict(t *testing.T) {
	cfg := appConfig{requireRemote: true, libsqlToken: "secret"}

	err := validateAppConfig(cfg)
	if err == nil {
		t.Fatalf("validateAppConfig() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "LIBSQL_DATABASE_URL is required") {
		t.Fatalf("error = %q, want LIBSQL_DATABASE_URL message", err.Error())
	}
}

func TestValidateAppConfigRequiresTokenWhenStrict(t *testing.T) {
	cfg := appConfig{requireRemote: true, libsqlURL: "libsql://example.turso.io"}

	err := validateAppConfig(cfg)
	if err == nil {
		t.Fatalf("validateAppConfig() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "LIBSQL_AUTH_TOKEN is required") {
		t.Fatalf("error = %q, want LIBSQL_AUTH_TOKEN message", err.Error())
	}
}

func TestShouldSkipDBInit(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: []string{"app"}, want: false},
		{name: "help flag", args: []string{"app", "--help"}, want: true},
		{name: "short help", args: []string{"app", "-h"}, want: true},
		{name: "version flag", args: []string{"app", "--version"}, want: true},
		{name: "short version", args: []string{"app", "-v"}, want: true},
		{name: "help command", args: []string{"app", "help"}, want: true},
		{name: "update command", args: []string{"app", "update"}, want: true},
		{name: "serve command", args: []string{"app", "serve"}, want: false},
		{name: "migrate command", args: []string{"app", "migrate"}, want: false},
		{name: "serve help flag", args: []string{"app", "serve", "--help"}, want: true},
		{name: "migrate short help", args: []string{"app", "migrate", "-h"}, want: true},
		{name: "serve version flag", args: []string{"app", "serve", "--version"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipDBInitArgs(tt.args)
			if got != tt.want {
				t.Fatalf("shouldSkipDBInitArgs(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
