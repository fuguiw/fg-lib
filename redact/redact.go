// Package redact provides a secret-scrubbing pipeline shared across
// fg-lib-consuming services. It covers three concerns:
//
//   - Textual rules (key=value, HTTP headers, URL userinfo, common token
//     prefixes, AWS access keys, PEM blocks) via [Redact].
//   - Structured JSON field-name redaction (snake_case + camelCase secret
//     names) via [RedactJSON].
//   - A [RedactedString] type marking values that have already passed
//     through [Redact], so type-safe "must redact" sinks can refuse plain
//     strings.
//
// The package has no other dependencies beyond the standard library and is
// safe for use from any layer.
package redact

import (
	"bytes"
	"encoding/json"
	"regexp"
)

// Redact returns s with secrets replaced by ***.
//
// Rules (kept in sync with redact_test.go):
//   - key=value pairs (api_key/token/secret/...);
//   - HTTP header lines (Authorization / Cookie / ...);
//   - URL userinfo;
//   - Anthropic/OpenAI `sk-...` tokens;
//   - AWS AKIA access key id;
//   - PEM private key blocks.
//
// Structured JSON field redaction (api_key/token/... by name) is the job of
// [RedactJSON]; Redact only mutates the input string textually.
func Redact(s string) string {
	if s == "" {
		return s
	}
	out := s
	for _, r := range redactRules {
		out = r.re.ReplaceAllString(out, r.repl)
	}
	return out
}

// RedactJSON walks raw and replaces any value whose JSON key matches a
// secret-name pattern (case-insensitive). Non-string leaf values are left
// alone; string leaves are additionally passed through [Redact].
//
// If raw is not valid JSON, the input is run through textual [Redact]
// instead of being dropped.
func RedactJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return json.RawMessage(Redact(string(raw)))
	}
	v = redactJSONNode(v)
	out, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(Redact(string(raw)))
	}
	return out
}

func redactJSONNode(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if IsSecretName(k) {
				t[k] = "***"
				continue
			}
			t[k] = redactJSONNode(child)
		}
		return t
	case []any:
		for i, item := range t {
			t[i] = redactJSONNode(item)
		}
		return t
	case string:
		return Redact(t)
	default:
		return t
	}
}

// RedactedString carries a string that is known to have already passed
// through [Redact]. A value of this type may be written to sinks declared
// "must redact". Production code must construct RedactedString only via
// [NewRedacted]; downstream projects should mirror the CI lint rule that
// blocks direct conversions elsewhere.
type RedactedString struct {
	s string
}

// NewRedacted is the sole legitimate constructor for [RedactedString].
func NewRedacted(s string) RedactedString { return RedactedString{s: Redact(s)} }

// String returns the redacted contents.
func (r RedactedString) String() string { return r.s }

// MarshalJSON renders r as a JSON string.
func (r RedactedString) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(r.s)+2))
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r.s); err != nil {
		return nil, err
	}
	// json.Encoder appends a newline; trim it.
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	return b, nil
}

type redactRule struct {
	re   *regexp.Regexp
	repl string
}

var redactRules = []redactRule{
	{
		// Generic key=value / key: value for sensitive identifiers.
		re:   regexp.MustCompile(`(?i)\b(api[-_]?key|token|secret|password|passwd|access[-_]?key|client[-_]?secret|bearer)\b\s*[:=]\s*["']?([A-Za-z0-9._\-+/=]{8,})["']?`),
		repl: "$1=***",
	},
	{
		// HTTP header lines — redact the entire value, not just the first token.
		re:   regexp.MustCompile(`(?im)^(Authorization|Cookie|Proxy-Authorization|X-Api-Key|X-Auth-Token):\s*.+$`),
		repl: "$1: ***",
	},
	{
		// URL userinfo: scheme://user:password@host → scheme://user:***@host
		re:   regexp.MustCompile(`([a-z][a-z0-9+\-.]*://)([^:/\s]+):([^@\s]+)@`),
		repl: "$1$2:***@",
	},
	{
		// Anthropic / OpenAI sk-... tokens.
		re:   regexp.MustCompile(`\bsk-[A-Za-z0-9_\-]{20,}\b`),
		repl: "sk-***",
	},
	{
		// AWS access key id.
		re:   regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
		repl: "AKIA****************",
	},
	{
		// PEM private key block.
		re:   regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`),
		repl: "-----BEGIN PRIVATE KEY----- ***REDACTED*** -----END PRIVATE KEY-----",
	},
}

// secretNamePattern matches JSON field names that should be wholesale replaced
// with "***" (snake_case + camelCase variants).
var secretNamePattern = regexp.MustCompile(`(?i)^(api[-_]?key|access[-_]?key|client[-_]?secret|secret|password|passwd|token|bearer|encrypt[-_]?key|verify[-_]?token|app[-_]?secret|web[-_]?token|cookie|authorization)$`)

// IsSecretName reports whether name matches a recognised secret field name
// (case-insensitive, snake or camel). Exposed so callers that build their
// own slog ReplaceAttr / response filter can reuse the same rule set.
func IsSecretName(name string) bool { return secretNamePattern.MatchString(name) }
