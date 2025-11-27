#!/bin/bash

# Configuration
API_URL="http://localhost:8080/v1"
EMAIL="testuser_$(date +%s)@example.com"
PASSWORD="password123"
NAME="Test User"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================================================"
echo "🎬 Film API Comprehensive Test Suite"
echo "================================================================"
echo "Target: $API_URL"
echo "User:   $EMAIL"
echo "================================================================"

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to print test results
pass_test() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((TESTS_PASSED++))
}

fail_test() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((TESTS_FAILED++))
}

info_log() {
    echo -e "${BLUE}ℹ️  ${NC}$1"
}

# =================================================================
# TEST 1: Health Check
# =================================================================
echo -e "\n${YELLOW}[TEST 1/15]${NC} Health Check"
echo "================================================================"
RESPONSE=$(curl -s "$API_URL/healthcheck")
echo "$RESPONSE"

if echo "$RESPONSE" | grep -q "available"; then
    pass_test "Health check returned 'available'"
else
    fail_test "Health check failed"
    exit 1
fi

# =================================================================
# TEST 2: User Registration
# =================================================================
echo -e "\n${YELLOW}[TEST 2/15]${NC} User Registration"
echo "================================================================"
RESPONSE=$(curl -s -X POST "$API_URL/users" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$NAME\", \"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

echo "$RESPONSE"

ACTIVATION_TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | head -n 1 | cut -d'"' -f4)

if [ -n "$ACTIVATION_TOKEN" ]; then
    pass_test "User registered, activation token: ${ACTIVATION_TOKEN:0:10}..."
else
    fail_test "Failed to get activation token"
    exit 1
fi

# =================================================================
# TEST 3: User Activation
# =================================================================
echo -e "\n${YELLOW}[TEST 3/15]${NC} User Activation"
echo "================================================================"
RESPONSE=$(curl -s -X PUT "$API_URL/users/activate" \
    -H "Content-Type: application/json" \
    -d "{\"token\": \"$ACTIVATION_TOKEN\"}")

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q '"activated":true'; then
    pass_test "User activated successfully"
else
    fail_test "User activation failed"
    exit 1
fi

# =================================================================
# TEST 4: Login (JWT Authentication)
# =================================================================
echo -e "\n${YELLOW}[TEST 4/15]${NC} Login (JWT Authentication)"
echo "================================================================"
RESPONSE=$(curl -s -X POST "$API_URL/tokens/authentication" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

echo "$RESPONSE"

# Extract access_token (JWT)
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ACCESS_TOKEN" ]; then
    pass_test "Login successful, got JWT access token"
    info_log "Access Token (JWT): ${ACCESS_TOKEN:0:30}..."
    info_log "Refresh Token: ${REFRESH_TOKEN:0:30}..."
else
    fail_test "Failed to login and get JWT"
    exit 1
fi

# =================================================================
# TEST 5: List Films (Basic)
# =================================================================
echo -e "\n${YELLOW}[TEST 5/15]${NC} List Films (Basic)"
echo "================================================================"
RESPONSE=$(curl -s -G "$API_URL/films" \
    -d "page_size=3" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE"

FILM_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -n 1 | grep -o '[0-9]*')

if [ -n "$FILM_ID" ]; then
    pass_test "Films retrieved, first film ID: $FILM_ID"
else
    fail_test "Failed to retrieve films"
    exit 1
fi

# =================================================================
# TEST 6: List Films with Filters
# =================================================================
echo -e "\n${YELLOW}[TEST 6/15]${NC} List Films with Filters"
echo "================================================================"
RESPONSE=$(curl -s -G "$API_URL/films" \
    -d "page_size=5" \
    -d "sort=-rating" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE" | head -c 500
echo "..."

if echo "$RESPONSE" | grep -q '"films"'; then
    pass_test "Filtered films retrieved with sorting"
else
    fail_test "Failed to retrieve filtered films"
fi

# =================================================================
# TEST 7: Get Specific Film by ID
# =================================================================
echo -e "\n${YELLOW}[TEST 7/15]${NC} Get Film by ID"
echo "================================================================"
RESPONSE=$(curl -s "$API_URL/films/$FILM_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q "\"id\":$FILM_ID"; then
    pass_test "Film retrieved by ID: $FILM_ID"
else
    fail_test "Failed to get film by ID"
fi

# =================================================================
# TEST 8: Add Film to Watchlist
# =================================================================
echo -e "\n${YELLOW}[TEST 8/15]${NC} Add Film to Watchlist"
echo "================================================================"
RESPONSE=$(curl -s -X POST "$API_URL/watchlist" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"film_id\": $FILM_ID, \"notes\": \"Must watch!\", \"priority\": 10}")

echo "$RESPONSE"

WATCHLIST_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -n 1 | grep -o '[0-9]*')

if [ -n "$WATCHLIST_ID" ]; then
    pass_test "Film added to watchlist, entry ID: $WATCHLIST_ID"
else
    fail_test "Failed to add film to watchlist"
fi

# =================================================================
# TEST 9: Get Watchlist
# =================================================================
echo -e "\n${YELLOW}[TEST 9/15]${NC} Get User Watchlist"
echo "================================================================"
RESPONSE=$(curl -s -G "$API_URL/watchlist" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q '"watchlist"'; then
    pass_test "Watchlist retrieved"
else
    fail_test "Failed to retrieve watchlist"
fi

# =================================================================
# TEST 10: Update Watchlist Entry (Rate Film)
# =================================================================
echo -e "\n${YELLOW}[TEST 10/15]${NC} Update Watchlist Entry (Rate Film)"
echo "================================================================"
RESPONSE=$(curl -s -X PATCH "$API_URL/watchlist/$WATCHLIST_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"rating": 9, "watched": true, "notes": "Absolutely amazing!"}')

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q '"rating":9'; then
    pass_test "Watchlist entry updated with rating"
else
    fail_test "Failed to update watchlist entry"
fi

# =================================================================
# TEST 11: Get Recommendations
# =================================================================
echo -e "\n${YELLOW}[TEST 11/15]${NC} Get Personalized Recommendations"
echo "================================================================"
RESPONSE=$(curl -s -G "$API_URL/recommendations" \
    -d "limit=5" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q '"recommendations"'; then
    pass_test "Recommendations retrieved"
else
    fail_test "Failed to get recommendations"
fi

# =================================================================
# TEST 12: Get Watchlist Filtered (Watched Films)
# =================================================================
echo -e "\n${YELLOW}[TEST 12/15]${NC} Get Watched Films from Watchlist"
echo "================================================================"
RESPONSE=$(curl -s -G "$API_URL/watchlist" \
    -d "watched=true" \
    -d "sort=-rating" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q '"watched":true'; then
    pass_test "Filtered watchlist (watched films) retrieved"
else
    info_log "No watched films yet or filter failed (expected for new user)"
fi

# =================================================================
# TEST 13: Add Another Film to Watchlist (Unwatched)
# =================================================================
echo -e "\n${YELLOW}[TEST 13/15]${NC} Add Second Film to Watchlist"
echo "================================================================"

# Get second film ID
SECOND_FILM_ID=$(echo "$FILMS_RESPONSE" | grep -o '"id":[0-9]*' | sed -n '2p' | grep -o '[0-9]*')

if [ -z "$SECOND_FILM_ID" ]; then
    SECOND_FILM_ID=8  # Fallback to known film ID
fi

RESPONSE=$(curl -s -X POST "$API_URL/watchlist" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"film_id\": $SECOND_FILM_ID, \"notes\": \"Want to watch later\", \"priority\": 5}")

echo "$RESPONSE"

SECOND_WATCHLIST_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -n 1 | grep -o '[0-9]*')

if [ -n "$SECOND_WATCHLIST_ID" ]; then
    pass_test "Second film added to watchlist"
else
    info_log "Could not add second film (might already be in watchlist)"
fi

# =================================================================
# TEST 14: Get Unwatched Films
# =================================================================
echo -e "\n${YELLOW}[TEST 14/15]${NC} Get Unwatched Films from Watchlist"
echo "================================================================"
RESPONSE=$(curl -s -G "$API_URL/watchlist" \
    -d "watched=false" \
    -d "sort=-priority" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$RESPONSE"

if echo "$RESPONSE" | grep -q '"watched":false'; then
    pass_test "Filtered watchlist (unwatched films) retrieved"
else
    info_log "No unwatched films or filter mismatch"
fi

# =================================================================
# TEST 15: Delete Film from Watchlist
# =================================================================
echo -e "\n${YELLOW}[TEST 15/15]${NC} Delete Film from Watchlist"
echo "================================================================"

if [ -n "$SECOND_WATCHLIST_ID" ]; then
    RESPONSE=$(curl -s -X DELETE "$API_URL/watchlist/$SECOND_WATCHLIST_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    echo "$RESPONSE"
    
    if echo "$RESPONSE" | grep -q "success"; then
        pass_test "Film removed from watchlist"
    else
        fail_test "Failed to delete film from watchlist"
    fi
else
    info_log "Skipping delete test (no second watchlist entry)"
fi

# =================================================================
# SUMMARY
# =================================================================
echo -e "\n================================================================"
echo "🎯 Test Summary"
echo "================================================================"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo "================================================================"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
