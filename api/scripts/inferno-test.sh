#!/usr/bin/env bash
#
# inferno-test.sh — Automated Inferno g(10) test runner
#
# Usage:
#   ./scripts/inferno-test.sh                       # SMART Standalone Patient App
#   ./scripts/inferno-test.sh --single-patient-api   # Run SMART + Single Patient API
#   ./scripts/inferno-test.sh --loop 300             # Run every 300 seconds (5 min)
#   ./scripts/inferno-test.sh --ci                   # CI mode: exit 1 on any non-TLS failure
#
set -euo pipefail

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------
EHR_HOST="localhost"
EHR_PORT="8000"
INFERNO_HOST="localhost"
INFERNO_PORT="4567"
DOCKER_HOST="host.docker.internal"

EHR_URL="http://${EHR_HOST}:${EHR_PORT}"
INFERNO_URL="http://${INFERNO_HOST}:${INFERNO_PORT}"
FHIR_URL="http://${DOCKER_HOST}:${EHR_PORT}/fhir"

CLIENT_ID="test-client"
CLIENT_SECRET="test-secret"
SCOPES="launch/patient openid fhirUser offline_access patient/*.read"

SMART_TEST_GROUP="g10_certification-g10_smart_standalone_patient_app"
SINGLE_PATIENT_API_GROUP="g10_certification-g10_smart_single_patient_api"
RUN_SINGLE_PATIENT_API=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
log()  { echo -e "${CYAN}[$(date +%H:%M:%S)]${NC} $*"; }
pass() { echo -e "  ${GREEN}✓${NC} $*"; }
fail() { echo -e "  ${RED}✗${NC} $*"; }
warn() { echo -e "  ${YELLOW}!${NC} $*"; }

check_service() {
    local name="$1" url="$2"
    if ! curl -sf --max-time 5 "$url" > /dev/null 2>&1; then
        echo -e "${RED}ERROR:${NC} $name is not running at $url"
        return 1
    fi
}

# ---------------------------------------------------------------------------
# Preflight checks
# ---------------------------------------------------------------------------
preflight() {
    log "Preflight checks..."

    check_service "EHR Server" "${EHR_URL}/fhir/metadata" || exit 1
    check_service "Inferno"    "${INFERNO_URL}/api" || {
        # Inferno returns 404 on /api but that means it's up
        curl -sf --max-time 5 "${INFERNO_URL}" > /dev/null 2>&1 || {
            echo -e "${RED}ERROR:${NC} Inferno is not running at ${INFERNO_URL}"
            exit 1
        }
    }

    # Verify OIDC discovery is available
    ISSUER=$(curl -s "${EHR_URL}/.well-known/openid-configuration" | python3 -c "import sys,json; print(json.load(sys.stdin).get('issuer',''))" 2>/dev/null || true)
    if [ -z "$ISSUER" ]; then
        echo -e "${RED}ERROR:${NC} OIDC discovery endpoint not responding"
        exit 1
    fi
    log "EHR issuer: $ISSUER"
}

# ---------------------------------------------------------------------------
# Run one test cycle
# ---------------------------------------------------------------------------
run_tests() {
    local start_time=$(date +%s)
    log "${BOLD}Starting Inferno g(10) SMART Standalone Patient App tests${NC}"

    # Build auth info JSON
    AUTH_INFO=$(python3 -c "
import json
print(json.dumps({
    'auth_type': 'symmetric',
    'pkce_support': 'enabled',
    'pkce_code_challenge_method': 'S256',
    'token_url': 'http://${DOCKER_HOST}:${EHR_PORT}/auth/token',
    'auth_url': 'http://${DOCKER_HOST}:${EHR_PORT}/auth/authorize',
    'client_id': '${CLIENT_ID}',
    'client_secret': '${CLIENT_SECRET}',
    'requested_scopes': '${SCOPES}'
}))
")

    # 1. Create test session
    log "Creating test session..."
    SESSION=$(curl -sf -X POST "${INFERNO_URL}/api/test_sessions" \
        -H 'Content-Type: application/json' \
        -d '{"test_suite_id": "g10_certification"}' \
        | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
    log "Session: ${SESSION}"

    # 2. Start test run
    log "Starting test run..."
    INPUTS=$(python3 -c "
import json
inputs = [
    {'name': 'url', 'value': '${FHIR_URL}'},
    {'name': 'standalone_client_id', 'value': '${CLIENT_ID}'},
    {'name': 'standalone_client_secret', 'value': '${CLIENT_SECRET}'},
    {'name': 'standalone_requested_scopes', 'value': '${SCOPES}'},
    {'name': 'standalone_smart_auth_info', 'value': ${AUTH_INFO}}
]
print(json.dumps({
    'test_session_id': '${SESSION}',
    'test_group_id': '${SMART_TEST_GROUP}',
    'inputs': inputs
}))
")

    RUN_ID=$(curl -sf -X POST "${INFERNO_URL}/api/test_runs" \
        -H 'Content-Type: application/json' \
        -d "$INPUTS" \
        | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
    log "Run: ${RUN_ID}"

    # 3. Wait for OAuth waiting state
    log "Waiting for OAuth redirect..."
    for i in $(seq 1 30); do
        sleep 2
        STATUS=$(curl -sf "${INFERNO_URL}/api/test_runs/${RUN_ID}" \
            | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null || echo "unknown")
        if [ "$STATUS" = "waiting" ]; then
            break
        fi
        if [ "$STATUS" = "done" ] || [ "$STATUS" = "error" ]; then
            log "Tests finished early with status: $STATUS"
            break
        fi
    done

    if [ "$STATUS" = "waiting" ]; then
        # 4. Extract authorize URL from waiting test
        AUTH_URL=$(curl -sf "${INFERNO_URL}/api/test_sessions/${SESSION}/results" | python3 -c "
import sys, json, re
d = json.load(sys.stdin)
for r in d:
    if r.get('result') == 'wait':
        for msg in (r.get('messages') or []):
            m = msg.get('message','')
            match = re.search(r'(http://${DOCKER_HOST}:${EHR_PORT}/auth/authorize\S+)', m)
            if match:
                print(match.group(1).rstrip('.'))
                sys.exit(0)
sys.exit(1)
" 2>/dev/null)

        if [ -z "$AUTH_URL" ]; then
            log "${RED}Could not extract authorize URL${NC}"
            return 1
        fi

        # 5. Complete OAuth flow
        log "Completing OAuth flow..."
        LOCAL_URL=$(echo "$AUTH_URL" | sed "s/${DOCKER_HOST}/${EHR_HOST}/g")
        REDIRECT=$(curl -sf -D - "$LOCAL_URL" 2>/dev/null \
            | grep -i "^location:" | tr -d '\r' | sed 's/^[Ll]ocation: //')

        if [ -z "$REDIRECT" ]; then
            log "${RED}No redirect from authorize endpoint${NC}"
            return 1
        fi

        # Hit Inferno callback
        curl -sf "$REDIRECT" > /dev/null 2>&1 || true
        log "OAuth callback sent"

        # 6. Wait for tests to complete
        log "Waiting for tests to finish..."
        for i in $(seq 1 120); do
            sleep 2
            STATUS=$(curl -sf "${INFERNO_URL}/api/test_runs/${RUN_ID}" \
                | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null || echo "unknown")
            if [ "$STATUS" = "done" ] || [ "$STATUS" = "error" ]; then
                break
            fi
        done
    fi

    # 7. Collect and display results
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    echo ""
    echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}  Inferno g(10) Results — $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"

    RESULT=$(curl -sf "${INFERNO_URL}/api/test_sessions/${SESSION}/results" | python3 -c "
import sys, json

d = json.load(sys.stdin)
passes = []
fails = []
tls_fails = []
skips = 0
errors = 0

for r in d:
    st = r.get('result', '')
    tid = r.get('test_id', '').split('-')[-1] if r.get('test_id') else ''
    msg = r.get('result_message', '')

    if st == 'pass':
        passes.append(tid)
    elif st == 'fail':
        if not tid:
            continue  # skip group-level aggregates
        if 'TLS' in msg or 'tls' in tid.lower():
            tls_fails.append(f'{tid}: {msg[:80]}')
        else:
            fails.append(f'{tid}: {msg[:80]}')
    elif st == 'skip':
        skips += 1
    elif st == 'error':
        errors += 1

total_real = len(passes) + len(fails) + len(tls_fails)
functional_pass = len(passes)
functional_total = functional_pass + len(fails)

# Print passes
for t in passes:
    print(f'PASS:{t}')

# Print TLS fails
for t in tls_fails:
    print(f'TLS_FAIL:{t}')

# Print real fails
for t in fails:
    print(f'FAIL:{t}')

# Summary line
print(f'SUMMARY:{functional_pass}:{len(fails)}:{len(tls_fails)}:{skips}:{errors}:{total_real}:{functional_total}')
")

    # Parse and display
    local func_pass=0 func_fail=0 tls_fail=0
    while IFS= read -r line; do
        case "$line" in
            PASS:*)   pass "${line#PASS:}" ;;
            TLS_FAIL:*) warn "${line#TLS_FAIL:} (TLS — expected in HTTP)" ;;
            FAIL:*)   fail "${line#FAIL:}" ;;
            SUMMARY:*)
                IFS=: read -r _ func_pass func_fail tls_fail skips errors total func_total <<< "$line"
                ;;
        esac
    done <<< "$RESULT"

    echo ""
    echo -e "${BOLD}───────────────────────────────────────────────────────────${NC}"

    if [ "$func_fail" -eq 0 ]; then
        echo -e "  ${GREEN}${BOLD}ALL FUNCTIONAL TESTS PASSING${NC}  (${func_pass}/${func_total})"
    else
        echo -e "  ${RED}${BOLD}${func_fail} FUNCTIONAL TEST(S) FAILING${NC}  (${func_pass}/${func_total})"
    fi

    if [ "$tls_fail" -gt 0 ]; then
        echo -e "  ${YELLOW}${tls_fail} TLS test(s) skipped (HTTP dev mode)${NC}"
    fi

    echo -e "  Duration: ${duration}s | Session: ${SESSION}"
    echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
    echo ""

    # Optionally run Single Patient API tests in the same session
    if [ "$RUN_SINGLE_PATIENT_API" = "1" ] && [ "$func_fail" -eq 0 ]; then
        run_single_patient_api "$SESSION"
    fi

    # Return exit code for CI mode
    if [ "${CI_MODE:-0}" = "1" ] && [ "$func_fail" -gt 0 ]; then
        return 1
    fi
    return 0
}

# ---------------------------------------------------------------------------
# Run Single Patient API tests (group 2) in an existing session
# ---------------------------------------------------------------------------
run_single_patient_api() {
    local session_id="$1"
    local start_time=$(date +%s)
    log "${BOLD}Starting Inferno US Core Single Patient API tests${NC}"

    # Group 2 uses the token from the SMART test (same session).
    # We only need to pass the test_session_id and test_group_id.
    local run_payload
    run_payload=$(python3 -c "
import json
print(json.dumps({
    'test_session_id': '${session_id}',
    'test_group_id': '${SINGLE_PATIENT_API_GROUP}',
    'inputs': [
        {'name': 'url', 'value': '${FHIR_URL}'}
    ]
}))
")

    local run_id
    run_id=$(curl -sf -X POST "${INFERNO_URL}/api/test_runs" \
        -H 'Content-Type: application/json' \
        -d "$run_payload" \
        | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
    log "Single Patient API Run: ${run_id}"

    # Wait for tests to complete
    log "Waiting for Single Patient API tests..."
    local status="unknown"
    for i in $(seq 1 180); do
        sleep 2
        status=$(curl -sf "${INFERNO_URL}/api/test_runs/${run_id}" \
            | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null || echo "unknown")
        if [ "$status" = "done" ] || [ "$status" = "error" ]; then
            break
        fi
    done

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    echo ""
    echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}  US Core Single Patient API — $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"

    local result
    result=$(curl -sf "${INFERNO_URL}/api/test_sessions/${session_id}/results" | python3 -c "
import sys, json

d = json.load(sys.stdin)
passes = []
fails = []
skips = 0
errors = 0

# Filter to only Single Patient API results (group 2)
for r in d:
    gid = r.get('test_group_id', '') or ''
    tid = r.get('test_id', '') or ''
    # Only include results from the single patient API group
    if 'single_patient_api' not in gid and 'single_patient_api' not in tid:
        continue
    st = r.get('result', '')
    short_id = tid.split('-')[-1] if tid else ''
    msg = r.get('result_message', '')

    if st == 'pass':
        passes.append(short_id)
    elif st == 'fail':
        if not short_id:
            continue
        fails.append(f'{short_id}: {msg[:100]}')
    elif st == 'skip':
        skips += 1
    elif st == 'error':
        errors += 1

total = len(passes) + len(fails)

for t in passes:
    print(f'PASS:{t}')
for t in fails:
    print(f'FAIL:{t}')
print(f'SUMMARY:{len(passes)}:{len(fails)}:0:{skips}:{errors}:{total}:{total}')
")

    local func_pass=0 func_fail=0
    while IFS= read -r line; do
        case "$line" in
            PASS:*)   pass "${line#PASS:}" ;;
            FAIL:*)   fail "${line#FAIL:}" ;;
            SUMMARY:*)
                IFS=: read -r _ func_pass func_fail _ skips errors total func_total <<< "$line"
                ;;
        esac
    done <<< "$result"

    echo ""
    echo -e "${BOLD}───────────────────────────────────────────────────────────${NC}"

    if [ "$func_fail" -eq 0 ]; then
        echo -e "  ${GREEN}${BOLD}ALL SINGLE PATIENT API TESTS PASSING${NC}  (${func_pass}/${func_total:-$func_pass})"
    else
        echo -e "  ${RED}${BOLD}${func_fail} SINGLE PATIENT API TEST(S) FAILING${NC}  (${func_pass}/${func_total:-0})"
    fi

    echo -e "  Duration: ${duration}s | Session: ${session_id}"
    echo -e "${BOLD}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
LOOP_INTERVAL=0
CI_MODE=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --loop)
            LOOP_INTERVAL="${2:-300}"
            shift 2
            ;;
        --ci)
            CI_MODE=1
            shift
            ;;
        --single-patient-api)
            RUN_SINGLE_PATIENT_API=1
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --single-patient-api  Also run US Core Single Patient API tests after SMART"
            echo "  --loop SECONDS        Run tests every N seconds (default: 300)"
            echo "  --ci                  CI mode: exit 1 on any non-TLS failure"
            echo "  --help                Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

preflight

if [ "$LOOP_INTERVAL" -gt 0 ]; then
    log "Running in loop mode (every ${LOOP_INTERVAL}s). Ctrl+C to stop."
    RUN_COUNT=0
    FAIL_COUNT=0
    while true; do
        RUN_COUNT=$((RUN_COUNT + 1))
        log "${BOLD}=== Run #${RUN_COUNT} ===${NC}"
        if ! run_tests; then
            FAIL_COUNT=$((FAIL_COUNT + 1))
        fi
        log "Stats: ${RUN_COUNT} runs, ${FAIL_COUNT} with failures. Next run in ${LOOP_INTERVAL}s..."
        sleep "$LOOP_INTERVAL"
    done
else
    run_tests
fi
