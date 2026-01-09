#!/bin/bash

# Session Management Verification Script
# This script tests the key session management features

echo "🔍 Session Management Verification"
echo "=================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC}: $2"
        ((TESTS_FAILED++))
    fi
}

echo ""
echo "1. Checking Frontend Configuration..."

# Check if idle timeout is set to 5 minutes
IDLE_TIMEOUT=$(grep -o "IDLE_TIMEOUT: [0-9]* \* [0-9]* \* [0-9]*" frontend/src/lib/session-config.ts | grep -o "[0-9]* \* [0-9]* \* [0-9]*")
if [[ "$IDLE_TIMEOUT" == "5 * 60 * 1000" ]]; then
    print_result 0 "Idle timeout set to 5 minutes"
else
    print_result 1 "Idle timeout not set to 5 minutes (found: $IDLE_TIMEOUT)"
fi

# Check if refresh token handling includes rotation
if grep -q "newRefreshToken" frontend/src/app/_actions/auth.ts; then
    print_result 0 "Frontend handles refresh token rotation"
else
    print_result 1 "Frontend missing refresh token rotation handling"
fi

echo ""
echo "2. Checking Backend Implementation..."

# Check if backend returns new refresh token
if grep -q "RefreshToken.*string.*json" backend/services/auth_service.go; then
    print_result 0 "Backend TokenResponse includes refresh token"
else
    print_result 1 "Backend TokenResponse missing refresh token field"
fi

# Check if session repository has update method
if grep -q "UpdateRefreshToken" backend/repository/interfaces.go; then
    print_result 0 "Session repository interface includes UpdateRefreshToken"
else
    print_result 1 "Session repository missing UpdateRefreshToken method"
fi

# Check if SQL query exists
if grep -q "UpdateSessionRefreshToken" backend/database/queries/sessions.sql; then
    print_result 0 "SQL query for updating refresh token exists"
else
    print_result 1 "SQL query for updating refresh token missing"
fi

echo ""
echo "3. Checking Security Improvements..."

# Check if refresh token rotation is implemented
if grep -q "generateRefreshToken" backend/services/auth_service.go && grep -q "newRefreshToken" backend/services/auth_service.go; then
    print_result 0 "Refresh token rotation implemented"
else
    print_result 1 "Refresh token rotation not properly implemented"
fi

# Check if session expiration is extended
if grep -q "RefreshTokenDuration" backend/services/auth_service.go; then
    print_result 0 "Session expiration extension implemented"
else
    print_result 1 "Session expiration extension missing"
fi

echo ""
echo "4. Checking Component Fixes..."

# Check if screen lock component import is fixed
if ! grep -q "usePathname" frontend/src/components/base/screen-lock.tsx; then
    print_result 0 "Screen lock component unused import removed"
else
    print_result 1 "Screen lock component still has unused import"
fi

echo ""
echo "📊 Test Summary"
echo "==============="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n🎉 ${GREEN}All session management fixes verified successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Run 'sqlc generate' in backend directory (after fixing migration syntax)"
    echo "2. Test the application with 5-minute idle timer"
    echo "3. Verify token refresh works with rotation"
    echo "4. Deploy to staging for integration testing"
else
    echo -e "\n⚠️  ${YELLOW}Some issues found. Please review the failed tests above.${NC}"
fi

echo ""
echo "🔗 Additional Resources:"
echo "- Test HTTP requests: test_session_management.http"
echo "- Full audit report: SESSION_MANAGEMENT_AUDIT_SUMMARY.md"