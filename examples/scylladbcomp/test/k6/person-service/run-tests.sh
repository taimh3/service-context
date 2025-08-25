#!/bin/bash

# K6 Load Testing Script for Persons API
# This script runs various k6 test scenarios for the ScyllaDB persons API

set -e

# Default configuration
BASE_URL=${BASE_URL:-"http://localhost:8080"}
TEST_FILE="test-person.js"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting K6 Load Tests for Persons API${NC}"
echo -e "${YELLOW}Base URL: ${BASE_URL}${NC}"

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}❌ k6 is not installed. Please install k6 first.${NC}"
    echo -e "${YELLOW}Installation instructions:${NC}"
    echo "  - macOS: brew install k6"
    echo "  - Ubuntu/Debian: sudo apt update && sudo apt install k6"
    echo "  - Or download from: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

echo -e "${GREEN}✅ k6 found: $(k6 version)${NC}"

# Function to run a test scenario
run_test() {
    local test_name="$1"
    local test_options="$2"
    
    echo -e "\n${YELLOW}📊 Running: ${test_name}${NC}"
    echo "----------------------------------------"
    
    if [ -n "$test_options" ]; then
        k6 run $test_options --env BASE_URL=$BASE_URL $TEST_FILE
    else
        k6 run --env BASE_URL=$BASE_URL $TEST_FILE
    fi
    
    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ ${test_name} completed successfully${NC}"
    else
        echo -e "${RED}❌ ${test_name} failed with exit code $exit_code${NC}"
    fi
    
    return $exit_code
}

# Function to check if API is running
check_api() {
    echo -e "${YELLOW}🔍 Checking if API is accessible...${NC}"
    
    if curl -s -f "${BASE_URL}/ping" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ API is accessible at ${BASE_URL}${NC}"
        return 0
    else
        echo -e "${RED}❌ API is not accessible at ${BASE_URL}${NC}"
        echo -e "${YELLOW}💡 Make sure your API server is running:${NC}"
        echo "   cd /path/to/your/golang-clean-architecture"
        echo "   make run  # or go run main.go"
        return 1
    fi
}

# Main execution
main() {
    # Check if API is accessible
    if ! check_api; then
        exit 1
    fi
    
    echo -e "\n${GREEN}🏁 Starting test scenarios...${NC}"
    
    # Scenario 1: Default load test
    echo -e "\n${YELLOW}═══════════════════════════════════════${NC}"
    echo -e "${YELLOW}🎯 Scenario 1: Default Load Test${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════${NC}"
    run_test "Default Load Test" ""
    
    # Scenario 2: Smoke test (quick verification)
    echo -e "\n${YELLOW}═══════════════════════════════════════${NC}"
    echo -e "${YELLOW}💨 Scenario 2: Smoke Test${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════${NC}"
    run_test "Smoke Test" "--vus 1 --duration 30s"
    
    # Scenario 3: Stress test
    echo -e "\n${YELLOW}═══════════════════════════════════════${NC}"
    echo -e "${YELLOW}💪 Scenario 3: Stress Test${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════${NC}"
    run_test "Stress Test" "--vus 20 --duration 2m"
    
    # Scenario 4: Spike test
    echo -e "\n${YELLOW}═══════════════════════════════════════${NC}"
    echo -e "${YELLOW}⚡ Scenario 4: Spike Test${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════${NC}"
    run_test "Spike Test" "--stage 10s:1,10s:20,30s:20,10s:1"
    
    echo -e "\n${GREEN}🎉 All test scenarios completed!${NC}"
    echo -e "${YELLOW}📊 Check the detailed results above for performance metrics.${NC}"
}

# Handle command line arguments
case "${1:-}" in
    "smoke")
        check_api && run_test "Smoke Test" "--vus 1 --duration 30s"
        ;;
    "stress")
        check_api && run_test "Stress Test" "--vus 20 --duration 2m"
        ;;
    "spike")
        check_api && run_test "Spike Test" "--stage 10s:1,10s:20,30s:20,10s:1"
        ;;
    "default"|"")
        main
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [test_type]"
        echo ""
        echo "Test types:"
        echo "  default  - Run all test scenarios (default)"
        echo "  smoke    - Quick smoke test with 1 user for 30s"
        echo "  stress   - Stress test with 20 users for 2m"
        echo "  spike    - Spike test with varying load"
        echo "  help     - Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  BASE_URL - API base URL (default: http://localhost:8080)"
        echo ""
        echo "Examples:"
        echo "  $0                           # Run all scenarios"
        echo "  $0 smoke                     # Run smoke test only"
        echo "  BASE_URL=http://api.example.com $0 stress"
        ;;
    *)
        echo -e "${RED}❌ Unknown test type: $1${NC}"
        echo "Run '$0 help' for usage information."
        exit 1
        ;;
esac
