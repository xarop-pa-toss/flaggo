#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://127.0.0.1:8080"
CURL_OPTS="--connect-timeout 3 --max-time 5 -s"

echo "=========================================="
echo "  Flaggo Stack Verification & API Test    "
echo "=========================================="

echo -e "\n1. Checking Health Check (/healthz)..."
if HEALTH_RESP=$(curl ${CURL_OPTS} "${BASE_URL}/healthz"); then
  echo "   Response: ${HEALTH_RESP}"
else
  echo "   [ERROR] Cannot connect to ${BASE_URL}. Ensure container stack is running!"
  exit 1
fi

echo -e "\n2. Creating test feature flag ('dark-mode')..."
CREATE_RESP=$(curl ${CURL_OPTS} -X POST "${BASE_URL}/api/flags" \
  -H "Content-Type: application/json" \
  -d '{"name":"dark-mode","description":"Dark mode feature flag","enabled":true,"rollout_percent":50}')
echo "   Response: ${CREATE_RESP}"

echo -e "\n3. Listing all active feature flags (/api/flags)..."
LIST_RESP=$(curl ${CURL_OPTS} "${BASE_URL}/api/flags")
echo "   Response: ${LIST_RESP}"

echo -e "\n4. Fetching 'dark-mode' flag details (/api/flags/dark-mode)..."
GET_RESP=$(curl ${CURL_OPTS} "${BASE_URL}/api/flags/dark-mode")
echo "   Response: ${GET_RESP}"

echo -e "\n5. Triggering Instant Kill-Switch (/api/flags/dark-mode/kill-switch)..."
KILL_RESP=$(curl ${CURL_OPTS} -X POST "${BASE_URL}/api/flags/dark-mode/kill-switch")
echo "   Response: ${KILL_RESP}"

echo -e "\n6. Verifying 'dark-mode' state after kill-switch..."
VERIFY_RESP=$(curl ${CURL_OPTS} "${BASE_URL}/api/flags/dark-mode")
echo "   Response: ${VERIFY_RESP}"

echo -e "\n7. Cleaning up 'dark-mode' flag (/api/flags/dark-mode)..."
DEL_RESP=$(curl ${CURL_OPTS} -X DELETE "${BASE_URL}/api/flags/dark-mode")
echo "   Response: ${DEL_RESP}"

echo -e "\n=========================================="
echo "  All Stack Verification Tests Completed! "
echo "=========================================="
