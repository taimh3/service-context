#!/bin/bash

# Quick validation script for k6 tests
# This script performs a simple validation of the test setup

echo "🔍 Validating K6 Test Setup..."

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo "❌ k6 is not installed"
    exit 1
fi

echo "✅ k6 is installed: $(k6 version)"

# Check if test files exist
if [ ! -f "test-get-task.js" ]; then
    echo "❌ test-get-task.js not found"
    exit 1
fi

echo "✅ Test file exists: test-get-task.js"

# Validate the test file syntax
echo "🔍 Validating test file syntax..."
if k6 inspect test-get-task.js > /dev/null 2>&1; then
    echo "✅ Test file syntax is valid"
else
    echo "❌ Test file has syntax errors"
    echo "Run: k6 inspect test-get-task.js"
    exit 1
fi

echo ""
echo "🎉 Setup validation complete!"
echo ""
echo "Next steps:"
echo "1. Start your API server:"
echo "   cd ../../"
echo "   go run main.go"
echo ""
echo "2. Run the tests:"
echo "   ./run-tests.sh"
echo "   # or"
echo "   k6 run test-get-task.js"
