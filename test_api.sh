#!/bin/bash

# Kalistheniks API Test Script
# This script tests all API endpoints with curl

set -e  # Exit on error

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="${EMAIL:-test_$(date +%s)@example.com}"
PASSWORD="${PASSWORD:-TestPass123!}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper function to print section headers
print_section() {
    echo -e "\n${BLUE}===================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}===================================================${NC}\n"
}

# Helper function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Helper function to print info
print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Helper function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if server is running
print_section "0. Health Check"
print_info "Checking if API is running at $BASE_URL..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
if [[ $HEALTH_RESPONSE == *"ok"* ]]; then
    print_success "API is healthy!"
    echo "Response: $HEALTH_RESPONSE"
else
    print_error "API is not responding correctly"
    exit 1
fi

# 1. Signup - Create a new user
print_section "1. Signup - Create New User"
print_info "Creating user with email: $EMAIL"

SIGNUP_RESPONSE=$(curl -s -X POST "$BASE_URL/signup" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$EMAIL\",
        \"password\": \"$PASSWORD\"
    }")

echo "Response: $SIGNUP_RESPONSE"

# Extract token from signup response
TOKEN=$(echo $SIGNUP_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
USER_ID=$(echo $SIGNUP_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    print_error "Failed to extract token from signup response"
    exit 1
fi

print_success "User created successfully!"
print_info "User ID: $USER_ID"
print_info "Token: ${TOKEN:0:20}..."

# 2. Login - Test authentication
print_section "2. Login - Authenticate User"
print_info "Logging in with email: $EMAIL"

LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$EMAIL\",
        \"password\": \"$PASSWORD\"
    }")

echo "Response: $LOGIN_RESPONSE"

# Extract new token from login
NEW_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$NEW_TOKEN" ]; then
    print_error "Failed to login"
    exit 1
fi

print_success "Login successful!"
print_info "New Token: ${NEW_TOKEN:0:20}..."

# Use the login token for subsequent requests
TOKEN=$NEW_TOKEN

# 3. Create Session
print_section "3. Create Training Session"
print_info "Creating a training session..."

SESSION_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"performed_at\": \"$SESSION_TIME\",
        \"session_type\": \"upper\",
        \"notes\": \"Push day - testing API\"
    }")

echo "Response: $SESSION_RESPONSE"

# Extract session ID
SESSION_ID=$(echo $SESSION_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -z "$SESSION_ID" ]; then
    print_error "Failed to create session"
    exit 1
fi

print_success "Session created successfully!"
print_info "Session ID: $SESSION_ID"

# 4. Add Sets to Session
print_section "4. Add Sets to Session"

# Generate some random UUIDs for exercise IDs (in a real scenario, these would be actual exercise IDs)
EXERCISE_ID_1="a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"  # Bench Press
EXERCISE_ID_2="b1ffcd88-8b1a-3de7-aa5c-5aa8ac291b22"  # Overhead Press

# Add first set
print_info "Adding Set 1: Bench Press - 80kg x 8 reps (RPE 7)"
SET1_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID/sets" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"exercise_id\": \"$EXERCISE_ID_1\",
        \"set_index\": 0,
        \"reps\": 8,
        \"weight_kg\": 80.0,
        \"rpe\": 7
    }")

echo "Response: $SET1_RESPONSE"
SET1_ID=$(echo $SET1_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$SET1_ID" ]; then
    print_success "Set 1 added successfully! (ID: $SET1_ID)"
else
    print_error "Failed to add Set 1"
fi

# Add second set
print_info "Adding Set 2: Bench Press - 85kg x 8 reps (RPE 8)"
SET2_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID/sets" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"exercise_id\": \"$EXERCISE_ID_1\",
        \"set_index\": 1,
        \"reps\": 8,
        \"weight_kg\": 85.0,
        \"rpe\": 8
    }")

echo "Response: $SET2_RESPONSE"
SET2_ID=$(echo $SET2_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$SET2_ID" ]; then
    print_success "Set 2 added successfully! (ID: $SET2_ID)"
else
    print_error "Failed to add Set 2"
fi

# Add third set
print_info "Adding Set 3: Bench Press - 90kg x 6 reps (RPE 9)"
SET3_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID/sets" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"exercise_id\": \"$EXERCISE_ID_1\",
        \"set_index\": 2,
        \"reps\": 6,
        \"weight_kg\": 90.0,
        \"rpe\": 9
    }")

echo "Response: $SET3_RESPONSE"
SET3_ID=$(echo $SET3_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$SET3_ID" ]; then
    print_success "Set 3 added successfully! (ID: $SET3_ID)"
else
    print_error "Failed to add Set 3"
fi

# Add fourth set (different exercise)
print_info "Adding Set 4: Overhead Press - 50kg x 10 reps (RPE 7)"
SET4_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID/sets" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"exercise_id\": \"$EXERCISE_ID_2\",
        \"set_index\": 0,
        \"reps\": 10,
        \"weight_kg\": 50.0,
        \"rpe\": 7
    }")

echo "Response: $SET4_RESPONSE"
SET4_ID=$(echo $SET4_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$SET4_ID" ]; then
    print_success "Set 4 added successfully! (ID: $SET4_ID)"
else
    print_error "Failed to add Set 4"
fi

# Add fifth set with high reps (for testing progression logic)
print_info "Adding Set 5: Overhead Press - 52.5kg x 12 reps (RPE 8)"
SET5_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID/sets" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"exercise_id\": \"$EXERCISE_ID_2\",
        \"set_index\": 1,
        \"reps\": 12,
        \"weight_kg\": 52.5,
        \"rpe\": 8
    }")

echo "Response: $SET5_RESPONSE"
SET5_ID=$(echo $SET5_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$SET5_ID" ]; then
    print_success "Set 5 added successfully! (ID: $SET5_ID)"
else
    print_error "Failed to add Set 5"
fi

# 5. List All Sessions
print_section "5. List All Sessions"
print_info "Fetching all sessions with sets..."

SESSIONS_RESPONSE=$(curl -s -X GET "$BASE_URL/sessions" \
    -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$SESSIONS_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$SESSIONS_RESPONSE"

SESSION_COUNT=$(echo "$SESSIONS_RESPONSE" | grep -o '"id":"' | wc -l)
print_success "Found $SESSION_COUNT session(s)"

# 6. Get Next Plan Suggestion
print_section "6. Get Next Workout Plan"
print_info "Requesting next workout suggestion based on history..."

PLAN_RESPONSE=$(curl -s -X GET "$BASE_URL/plan/next" \
    -H "Authorization: Bearer $TOKEN")

echo "Response:"
echo "$PLAN_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$PLAN_RESPONSE"

if [[ $PLAN_RESPONSE == *"exercise_id"* ]]; then
    print_success "Plan suggestion received!"
else
    print_error "Failed to get plan suggestion"
fi

# 7. Create Another Session (Lower Body)
print_section "7. Create Second Session (Lower Body)"
print_info "Creating a lower body session to test progression logic..."

SESSION_TIME_2=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
SESSION_RESPONSE_2=$(curl -s -X POST "$BASE_URL/sessions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"performed_at\": \"$SESSION_TIME_2\",
        \"session_type\": \"lower\",
        \"notes\": \"Leg day - testing API\"
    }")

echo "Response: $SESSION_RESPONSE_2"

SESSION_ID_2=$(echo $SESSION_RESPONSE_2 | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -z "$SESSION_ID_2" ]; then
    print_error "Failed to create second session"
else
    print_success "Second session created successfully!"
    print_info "Session ID: $SESSION_ID_2"

    # Add a set to the lower body session
    EXERCISE_ID_3="c2ffde77-7a2b-4ef8-bb6d-7cc0cd492c33"  # Squat

    print_info "Adding Set: Squat - 100kg x 5 reps (RPE 8)"
    SET6_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID_2/sets" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"exercise_id\": \"$EXERCISE_ID_3\",
            \"set_index\": 0,
            \"reps\": 5,
            \"weight_kg\": 100.0,
            \"rpe\": 8
        }")

    echo "Response: $SET6_RESPONSE"
    SET6_ID=$(echo $SET6_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

    if [ -n "$SET6_ID" ]; then
        print_success "Set added to lower body session! (ID: $SET6_ID)"
    fi
fi

# Final Summary
print_section "Test Summary"
print_success "All API tests completed!"
echo -e "\n${YELLOW}Generated Test Data:${NC}"
echo "  Email: $EMAIL"
echo "  Password: $PASSWORD"
echo "  User ID: $USER_ID"
echo "  Token: ${TOKEN:0:30}..."
echo "  Session 1 ID: $SESSION_ID (upper)"
echo "  Session 2 ID: $SESSION_ID_2 (lower)"
echo ""
echo -e "${YELLOW}You can use this token for manual testing:${NC}"
echo "  export TOKEN=\"$TOKEN\""
echo ""
echo -e "${YELLOW}Example manual curl commands:${NC}"
echo "  curl -H \"Authorization: Bearer \$TOKEN\" $BASE_URL/sessions"
echo "  curl -H \"Authorization: Bearer \$TOKEN\" $BASE_URL/plan/next"
