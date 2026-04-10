package logger

import (
	"encoding/json"
	"testing"
)

func TestSanitizeLogPayloadMasksSensitiveKeys(t *testing.T) {
	input := []byte(`{"level":"info","password":"abc123","token":"signed","meta":{"refresh_token":"r1","safe":"ok"},"data":[{"client_secret":"c1"},{"note":"visible"}]}` + "\n")
	output := sanitizeLogPayload(input)

	var payload map[string]any
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("failed to unmarshal sanitized payload: %v", err)
	}

	if got := payload["password"]; got != redactedLogValue {
		t.Fatalf("expected password to be redacted, got %#v", got)
	}

	if got := payload["token"]; got != redactedLogValue {
		t.Fatalf("expected token to be redacted, got %#v", got)
	}

	meta, ok := payload["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta object, got %#v", payload["meta"])
	}

	if got := meta["refresh_token"]; got != redactedLogValue {
		t.Fatalf("expected refresh_token to be redacted, got %#v", got)
	}

	if got := meta["safe"]; got != "ok" {
		t.Fatalf("expected safe field to stay visible, got %#v", got)
	}

	data, ok := payload["data"].([]any)
	if !ok || len(data) != 2 {
		t.Fatalf("expected data array with 2 elements, got %#v", payload["data"])
	}

	first, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first data element to be object, got %#v", data[0])
	}

	if got := first["client_secret"]; got != redactedLogValue {
		t.Fatalf("expected client_secret to be redacted, got %#v", got)
	}

	second, ok := data[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second data element to be object, got %#v", data[1])
	}

	if got := second["note"]; got != "visible" {
		t.Fatalf("expected note to stay visible, got %#v", got)
	}
}

func TestSanitizeLogPayloadPreservesNonJSONInput(t *testing.T) {
	input := []byte("plain text log line\n")
	output := sanitizeLogPayload(input)

	if string(output) != string(input) {
		t.Fatalf("expected non-json payload to be unchanged, got %q", string(output))
	}
}

func TestIsSensitiveLogKey(t *testing.T) {
	markers := loadSensitiveKeyMarkers("")

	if !isSensitiveLogKey("Authorization", markers) {
		t.Fatal("expected Authorization to be sensitive")
	}

	if !isSensitiveLogKey("refresh_token", markers) {
		t.Fatal("expected refresh_token to be sensitive")
	}

	if isSensitiveLogKey("duration_ms", markers) {
		t.Fatal("expected duration_ms to be non-sensitive")
	}
}

func TestLoadSensitiveKeyMarkersAddsFromEnv(t *testing.T) {
	markers := loadSensitiveKeyMarkers("pin, otp_code, session-id, token")
	set := map[string]struct{}{}
	for _, marker := range markers {
		set[marker] = struct{}{}
	}

	if _, ok := set["password"]; !ok {
		t.Fatal("expected default key marker password to exist")
	}

	if _, ok := set["pin"]; !ok {
		t.Fatal("expected env key marker pin to exist")
	}

	if _, ok := set["otpcode"]; !ok {
		t.Fatal("expected env key marker otp_code to be normalized")
	}

	if _, ok := set["sessionid"]; !ok {
		t.Fatal("expected env key marker session-id to be normalized")
	}
}
