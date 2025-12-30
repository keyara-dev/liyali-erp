#!/bin/bash

echo "Testing all unit test files individually..."
echo "=========================================="

cd tests/unit

for file in *.go; do
    if [[ "$file" == "README.md" ]] || [[ "$file" == *".disabled" ]]; then
        continue
    fi
    
    echo "Testing: $file"
    if go test "../../tests/unit/$file" > /dev/null 2>&1; then
        echo "✅ PASS: $file"
    else
        echo "❌ FAIL: $file"
        echo "   Error details:"
        go test "../../tests/unit/$file" 2>&1 | head -3 | sed 's/^/   /'
    fi
    echo ""
done

echo ""
echo "Testing integration test files..."
echo "================================="

cd ../integration

for file in *.go; do
    if [[ "$file" == "helpers.go" ]] || [[ "$file" == "README.md" ]]; then
        continue
    fi
    
    echo "Testing: $file"
    if go test "../../tests/integration/$file" "../../tests/integration/helpers.go" > /dev/null 2>&1; then
        echo "✅ PASS: $file"
    else
        echo "❌ FAIL: $file"
        echo "   Error details:"
        go test "../../tests/integration/$file" "../../tests/integration/helpers.go" 2>&1 | head -3 | sed 's/^/   /'
    fi
    echo ""
done