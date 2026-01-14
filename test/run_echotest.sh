#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAKE_LOG="/tmp/fakeipapi.log"
ECHO_LOG="/tmp/echotest.log"
GEO_LOG="/tmp/geoproxy.log"
CONFIG_PATH="/tmp/geoproxy_integration.yaml"

cleanup() {
  if [[ -n "${GEO_PID:-}" ]]; then
    kill "${GEO_PID}" 2>/dev/null || true
  fi
  if [[ -n "${ECHO_PID:-}" ]]; then
    kill "${ECHO_PID}" 2>/dev/null || true
  fi
  if [[ -n "${FAKE_PID:-}" ]]; then
    kill "${FAKE_PID}" 2>/dev/null || true
  fi
  wait "${GEO_PID:-}" "${ECHO_PID:-}" "${FAKE_PID:-}" 2>/dev/null || true
}
trap cleanup EXIT

cat > "${CONFIG_PATH}" <<'EOF'
servers:
  - listenIP: "127.0.0.1"
    listenPort: "5555"
    backendIP: "127.0.0.1"
    backendPort: "5556"
    allowedCountries:
      - "US"
EOF

(
  cd "${ROOT_DIR}/test/fakeipapi"
  go run . -countryCode US -region WA >"${FAKE_LOG}" 2>&1
) &
FAKE_PID=$!

(
  cd "${ROOT_DIR}/test/echotest"
  go run . -server -port 5556 >"${ECHO_LOG}" 2>&1
) &
ECHO_PID=$!

(
  cd "${ROOT_DIR}"
  go run . -config "${CONFIG_PATH}" -ipapi http://127.0.0.1:8181/json/ >"${GEO_LOG}" 2>&1
) &
GEO_PID=$!

sleep 1

(
  cd "${ROOT_DIR}/test/echotest"
  timeout 30s go run . -client -serverIP 127.0.0.1 -port 5555 -n 100
) || true

echo "Integration run complete."
echo "Logs: ${FAKE_LOG} ${ECHO_LOG} ${GEO_LOG}"
