#!/bin/bash

# Configuration
BASE_URL="http://localhost:9527"
RANDOM_SUFFIX=$((RANDOM))
EMAIL="user_${RANDOM_SUFFIX}@test.com"
PASSWORD="password123"
NAME="Test User"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper Functions
log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

log_info() {
    echo -e "${CYAN}$1${NC}"
}

# JSON Parser using Python
parse_json() {
    echo "$1" | python3 -c "import sys, json; print(json.load(sys.stdin)$2)" 2>/dev/null
}

echo "🚀 Starting Comprehensive MVP 1.0 Test Suite (Linux)..."
echo "------------------------------------------------"

# ---------------------------------------------------------
# 1. User System Tests
# ---------------------------------------------------------
log_info "\n[1/6] Testing User System (Signup & Login)..."

# Signup
log_info "Attempting Signup with $EMAIL..."
SIGNUP_RESP=$(curl -s -X POST "$BASE_URL/signup" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$NAME\", \"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

USER_ID=$(parse_json "$SIGNUP_RESP" "['id']")

if [ "$USER_ID" != "" ] && [ "$USER_ID" != "None" ]; then
    log_success "User registered: $EMAIL (ID: $USER_ID)"
else
    log_error "Signup failed: $SIGNUP_RESP"
fi

# Login
log_info "Attempting Login..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

TOKEN=$(parse_json "$LOGIN_RESP" "['token']")

if [ "$TOKEN" != "" ] && [ "$TOKEN" != "None" ]; then
    log_success "Login successful, Token received."
else
    log_error "Login failed: $LOGIN_RESP"
fi

# Me
log_info "Checking /api/me..."
ME_RESP=$(curl -s -X GET "$BASE_URL/api/me" \
    -H "Authorization: Bearer $TOKEN")

ME_NAME=$(parse_json "$ME_RESP" "['name']")

if [ "$ME_NAME" == "$NAME" ]; then
    log_success "Authenticated as: $ME_NAME"
else
    log_error "Get Me failed: $ME_RESP"
fi

# ---------------------------------------------------------
# 1.4 Update Profile
# ---------------------------------------------------------
log_info "Updating profile fields..."
UPDATE_RESP=$(curl -s -X PATCH "$BASE_URL/api/me" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$NAME\",\"school\":\"Test School\",\"student_id\":\"SID123\",\"birthdate\":\"2000-01-01\",\"address\":\"Test Address\",\"gender\":\"OTHER\"}")

UPD_SCHOOL=$(parse_json "$UPDATE_RESP" "['school']")
if [ "$UPD_SCHOOL" == "Test School" ]; then
    log_success "Profile updated (school=$UPD_SCHOOL)"
else
    log_error "Profile update failed: $UPDATE_RESP"
fi

# ---------------------------------------------------------
# 2. Public Resource Tests (Anonymous Upload)
# ---------------------------------------------------------
log_info "\n[2/6] Testing Public Resources (Anonymous Upload)..."

TEST_FILE="test_doc_${RANDOM_SUFFIX}.txt"
TEST_TITLE="Lecture Notes ${RANDOM_SUFFIX}"
TEST_DESC="Anonymous upload test"
echo "This is a unique content for testing hash ${RANDOM_SUFFIX}" > "$TEST_FILE"

log_info "Uploading file $TEST_FILE..."

UPLOAD_RESP=$(curl -s -X POST "$BASE_URL/api/public/resources" \
    -F "title=$TEST_TITLE" \
    -F "description=$TEST_DESC" \
    -F "file=@$TEST_FILE")

# Clean up temp file
rm "$TEST_FILE"

STATUS=$(parse_json "$UPLOAD_RESP" "['status']")
FILE_HASH=$(parse_json "$UPLOAD_RESP" "['file_hash']")
RESOURCE_ID=$(parse_json "$UPLOAD_RESP" "['id']")

if [ "$STATUS" == "PENDING" ] && [ "$FILE_HASH" != "" ] && [ "$FILE_HASH" != "None" ]; then
    log_success "Anonymous upload successful (ID: $RESOURCE_ID, Status: $STATUS)"
else
    log_error "Upload failed: $UPLOAD_RESP"
fi

# ---------------------------------------------------------
# 3. List Resources
# ---------------------------------------------------------
log_info "\n[3/6] Testing List Resources..."

LIST_RESP=$(curl -s -X GET "$BASE_URL/api/public/resources")
COUNT=$(echo "$LIST_RESP" | python3 -c "import sys, json; print(len(json.load(sys.stdin)))" 2>/dev/null)

if [ "$COUNT" -ge 1 ]; then
    log_success "List resources successful. Found $COUNT resources."
else
    log_error "List resources failed or empty: $LIST_RESP"
fi

echo "------------------------------------------------"
log_success "🎉 All tests passed!"
