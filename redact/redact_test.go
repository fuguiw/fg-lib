package redact

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRedact_TextualRules covers the textual redactor rules.
func TestRedact_TextualRules(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		mustMatch []string
		mustMiss  []string
	}{
		{
			name: "api_key kv",
			// kv rule consumes both halves; AKIA shape is also gone via the kv rule.
			in:        `api_key=AKIA0123456789ABCDEF token: "bearer-abcdefg12345"`,
			mustMatch: []string{"api_key=***"},
			mustMiss:  []string{"AKIA0123456789ABCDEF", "bearer-abcdefg12345"},
		},
		{
			name:      "standalone AKIA token",
			in:        "saw key AKIAIOSFODNN7EXAMPLE in audit log",
			mustMatch: []string{"AKIA****************"},
			mustMiss:  []string{"AKIAIOSFODNN7EXAMPLE"},
		},
		{
			name:      "http header authorization",
			in:        "Authorization: Bearer xyz-very-secret-token",
			mustMatch: []string{"Authorization: ***"},
			mustMiss:  []string{"xyz-very-secret-token"},
		},
		{
			name:      "url userinfo",
			in:        "redis://admin:hunter2@cache.local:6379/0",
			mustMatch: []string{"admin:***@cache.local"},
			mustMiss:  []string{"hunter2"},
		},
		{
			name:      "anthropic sk-",
			in:        "key=sk-ant-api03-AAAAAAAAAAAAAAAAAAAAAAAA",
			mustMatch: []string{"sk-***"},
			mustMiss:  []string{"sk-ant-api03-AAAAAAAAAAAAAAAAAAAAAAAA"},
		},
		{
			name: "pem private key",
			in: "-----BEGIN RSA PRIVATE KEY-----\n" +
				"MIIEpAIBAAKCAQEA...\n" +
				"-----END RSA PRIVATE KEY-----",
			mustMatch: []string{"***REDACTED***"},
			mustMiss:  []string{"MIIEpAIBAAKCAQEA"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Redact(tc.in)
			for _, m := range tc.mustMatch {
				if !strings.Contains(got, m) {
					t.Errorf("expected %q in output\ngot: %s", m, got)
				}
			}
			for _, m := range tc.mustMiss {
				if strings.Contains(got, m) {
					t.Errorf("expected %q to be redacted out\ngot: %s", m, got)
				}
			}
		})
	}
}

func TestRedactJSON_KeyNameRedaction(t *testing.T) {
	in := []byte(`{
		"app_secret": "shh",
		"app_id": "cli_visible",
		"nested": {"token": "t-XYZ", "ok": 1},
		"arr": [{"password": "p"}, {"name": "ok"}]
	}`)
	out := RedactJSON(json.RawMessage(in))
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v\nout=%s", err, out)
	}
	if m["app_secret"] != "***" {
		t.Errorf("app_secret should be ***, got %v", m["app_secret"])
	}
	if m["app_id"] != "cli_visible" {
		t.Errorf("app_id should remain plain, got %v", m["app_id"])
	}
	nested := m["nested"].(map[string]any)
	if nested["token"] != "***" {
		t.Errorf("nested token should be ***, got %v", nested["token"])
	}
	arr := m["arr"].([]any)
	first := arr[0].(map[string]any)
	if first["password"] != "***" {
		t.Errorf("password in arr should be ***, got %v", first["password"])
	}
}

// TestRedactJSON_CamelCaseKeyRedaction locks in "snake & camel two-form"
// support — camelCase variants of secret names must match without depending
// on explicit separators.
func TestRedactJSON_CamelCaseKeyRedaction(t *testing.T) {
	in := []byte(`{
		"apiKey": "ak-1",
		"clientSecret": "cs-1",
		"verifyToken": "vt-1",
		"encryptKey": "ek-1",
		"appSecret": "as-1",
		"webToken": "wt-1",
		"accessKey": "ack-1",
		"safeField": "keep"
	}`)
	out := RedactJSON(json.RawMessage(in))
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v\nout=%s", err, out)
	}
	for _, k := range []string{"apiKey", "clientSecret", "verifyToken", "encryptKey", "appSecret", "webToken", "accessKey"} {
		if m[k] != "***" {
			t.Errorf("camelCase secret %s must be ***, got %v", k, m[k])
		}
	}
	if m["safeField"] != "keep" {
		t.Errorf("non-secret camelCase field altered: %v", m["safeField"])
	}
}

// TestIsSecretName_CamelCase guards the regex used by both RedactJSON and
// any slog ReplaceAttr key-based wholesale redact path.
func TestIsSecretName_CamelCase(t *testing.T) {
	for _, name := range []string{
		"api_key", "apiKey", "APIKey",
		"app_secret", "appSecret",
		"verify_token", "verifyToken",
		"encrypt_key", "encryptKey",
		"web_token", "webToken",
		"Authorization", "Cookie",
		"access_key", "accessKey",
	} {
		if !IsSecretName(name) {
			t.Errorf("IsSecretName(%q) should be true", name)
		}
	}
	for _, name := range []string{"app_id", "user_id", "bot_open_id", "safeField", "ok"} {
		if IsSecretName(name) {
			t.Errorf("IsSecretName(%q) should be false", name)
		}
	}
}

func TestRedactedString_OnlyViaNewRedacted(t *testing.T) {
	r := NewRedacted("authorization: Bearer abcd-secret-token-xxxxx")
	if strings.Contains(r.String(), "abcd-secret-token-xxxxx") {
		t.Fatalf("RedactedString leaked secret: %q", r.String())
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "abcd-secret-token-xxxxx") {
		t.Fatalf("RedactedString JSON leaked: %s", b)
	}
}
