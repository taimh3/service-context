#!/bin/bash

# Quick validation script for k6 tests
# This script performs a simple validation of the test setup

echo "🔍 Validating K6 Test Setup for Person Service..."

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo "❌ k6 is not installed"
    exit 1
fi

echo "✅ k6 is installed: $(k6 version)"

# Check if test files exist
if [ ! -f "test-person.js" ]; then
    echo "❌ test-person.js not found"
    exit 1
fi

echo "✅ Test file exists: test-person.js"

# Validate the test file syntax
echo "🔍 Validating test file syntax..."
if k6 inspect test-person.js > /dev/null 2>&1; then
    echo "✅ Test file syntax is valid"
else
    echo "❌ Test file has syntax errors"
    echo "Run: k6 inspect test-person.js"
    exit 1
fi

# Check if config file exists
if [ ! -f "config.json" ]; then
    echo "⚠️  config.json not found (optional)"
else
    echo "✅ Config file exists: config.json"
fi

# Check if run script exists
if [ ! -f "run-tests.sh" ]; then
    echo "❌ run-tests.sh not found"
    exit 1
fi

echo "✅ Run script exists: run-tests.sh"

# Check if run script is executable
if [ ! -x "run-tests.sh" ]; then
    echo "⚠️  run-tests.sh is not executable, fixing..."
    chmod +x run-tests.sh
    echo "✅ run-tests.sh is now executable"
else
    echo "✅ run-tests.sh is executable"
fi

echo ""
echo "🎉 Setup validation complete!"
echo ""
echo "Next steps:"
echo "1. Start your API server:"
echo "   cd ../../"
echo "   go run main.go"
echo ""
echo "2. Run tests:"
echo "   ./run-tests.sh               # All test scenarios"
echo "   ./run-tests.sh smoke         # Quick smoke test"
echo "   ./run-tests.sh stress        # Stress test"
echo "   ./run-tests.sh spike         # Spike test"
echo ""
echo "3. Or run directly with k6:"
echo "   k6 run test-person.js        # Default test"
echo "   k6 run --vus 5 --duration 1m test-person.js  # Custom settings"
