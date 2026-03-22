#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
REGISTER_URL="$BASE_URL/auth/register"
LOGIN_URL="$BASE_URL/auth/login"
REFRESH_URL="$BASE_URL/auth/refresh"
ME_URL="$BASE_URL/auth/me"
LOGOUT_URL="$BASE_URL/auth/logout"

timestamp="$(date +%s)"
username="admin_${timestamp}"
email="${username}@example.com"
password="admin123"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

request() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  local token="${4:-}"
  local response_file="$tmp_dir/response.json"

  local -a headers=(-H "Content-Type: application/json")
  if [[ -n "$token" ]]; then
    headers+=(-H "Authorization: Bearer $token")
  fi

  local status
  if [[ -n "$body" ]]; then
    status="$(curl -sS -o "$response_file" -w '%{http_code}' -X "$method" "${headers[@]}" -d "$body" "$url")"
  else
    status="$(curl -sS -o "$response_file" -w '%{http_code}' -X "$method" "${headers[@]}" "$url")"
  fi

  printf '%s\n' "$status"
  printf '%s\n' "$response_file"
}

assert_json() {
  local file="$1"
  local py="$2"
  python3 - "$file" "$py" <<'PY'
import json
import sys

path = sys.argv[1]
code = sys.argv[2]
with open(path, "r", encoding="utf-8") as fh:
    payload = json.load(fh)

namespace = {"payload": payload}
exec(code, {}, namespace)
PY
}

assert_status() {
  local actual="$1"
  local expected="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "[$label] expected status $expected, got $actual"
    exit 1
  fi
}

run_step() {
  local label="$1"
  local expected_status="$2"
  local method="$3"
  local url="$4"
  local body="${5:-}"
  local token="${6:-}"

  mapfile -t result < <(request "$method" "$url" "$body" "$token")
  local status="${result[0]}"
  local file="${result[1]}"

  assert_status "$status" "$expected_status" "$label"
  echo "[$label] status=$status"
  cat "$file"
  echo
  LAST_RESPONSE_FILE="$file"
}

echo "Testing auth API against $BASE_URL"

run_step "register" "201" "POST" "$REGISTER_URL" "{\"username\":\"$username\",\"email\":\"$email\",\"password\":\"$password\"}"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is True
assert payload["message"] == "Account created successfully"
assert payload["data"]["message"] == "account created"
assert "meta" in payload
assert "timestamp" in payload["meta"]
'

run_step "login" "200" "POST" "$LOGIN_URL" "{\"username\":\"$username\",\"password\":\"$password\"}"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is True
assert payload["message"] == "Login successful"
assert payload["data"]["token_type"] == "Bearer"
assert payload["data"]["access_token"]
assert payload["data"]["refresh_token"]
'
TOKEN_AND_REFRESH="$(python3 - "$LAST_RESPONSE_FILE" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as fh:
    payload = json.load(fh)
print(payload["data"]["access_token"])
print(payload["data"]["refresh_token"])
PY
)"
TOKEN="$(printf '%s\n' "$TOKEN_AND_REFRESH" | sed -n '1p')"
REFRESH_TOKEN="$(printf '%s\n' "$TOKEN_AND_REFRESH" | sed -n '2p')"

run_step "me-before-refresh" "200" "GET" "$ME_URL" "" "$TOKEN"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is True
assert payload["data"]["username"]
'

run_step "refresh" "200" "POST" "$REFRESH_URL" "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is True
assert payload["message"] == "Token refreshed successfully"
assert payload["data"]["access_token"]
assert payload["data"]["refresh_token"]
'
TOKEN_AND_REFRESH="$(python3 - "$LAST_RESPONSE_FILE" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as fh:
    payload = json.load(fh)
print(payload["data"]["access_token"])
print(payload["data"]["refresh_token"])
PY
)"
TOKEN="$(printf '%s\n' "$TOKEN_AND_REFRESH" | sed -n '1p')"
ROTATED_REFRESH_TOKEN="$(printf '%s\n' "$TOKEN_AND_REFRESH" | sed -n '2p')"

run_step "old-refresh-rejected" "401" "POST" "$REFRESH_URL" "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is False
assert payload["errors"][0]["code"] == "UNAUTHORIZED"
'

run_step "me-after-refresh" "200" "GET" "$ME_URL" "" "$TOKEN"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is True
assert payload["data"]["username"]
'

run_step "logout" "200" "POST" "$LOGOUT_URL" "" "$TOKEN"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is True
assert payload["message"] == "Logout acknowledged"
'

run_step "me-after-logout" "401" "GET" "$ME_URL" "" "$TOKEN"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is False
assert payload["errors"][0]["code"] == "UNAUTHORIZED"
'

run_step "refresh-after-logout" "401" "POST" "$REFRESH_URL" "{\"refresh_token\":\"$ROTATED_REFRESH_TOKEN\"}"
assert_json "$LAST_RESPONSE_FILE" '
assert payload["success"] is False
assert payload["errors"][0]["code"] == "UNAUTHORIZED"
'

echo "Auth API smoke test passed"
