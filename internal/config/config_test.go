package config

import (
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestParseDefaults(t *testing.T) {
	cfg, err := parse(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 || cfg.LogLevel != "info" || cfg.SessionTTL != 30*24*time.Hour {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestParseRequiresDatabaseURL(t *testing.T) {
	if _, err := parse(env(nil)); err == nil {
		t.Fatal("expected an error without DATABASE_URL")
	}
}

func TestParseRejectsBadValues(t *testing.T) {
	cases := map[string]map[string]string{
		"port text":    {"DATABASE_URL": "x", "PORT": "abc"},
		"port range":   {"DATABASE_URL": "x", "PORT": "70000"},
		"ttl negative": {"DATABASE_URL": "x", "SESSION_TTL": "-1h"},
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse(env(m)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
