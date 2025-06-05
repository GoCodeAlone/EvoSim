#!/bin/bash
# Web Interface End-to-End Tests
# Tests the web interface functionality without browser automation

set -e

echo "Starting EvoSim Web Interface Tests..."

# Test configuration
WEB_PORT=8080
PID_FILE="/tmp/evosim_test.pid"
LOG_FILE="/tmp/evosim_test.log"

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID"
            echo "Stopped EvoSim web server (PID: $PID)"
        fi
        rm -f "$PID_FILE"
    fi
    rm -f "$LOG_FILE"
}

# Set trap for cleanup
trap cleanup EXIT

# Start EvoSim web server in background
echo "Starting EvoSim web server on port $WEB_PORT..."
cd /home/runner/work/EvoSim/EvoSim
GOWORK=off go run . -web -web-port $WEB_PORT -pop-size 5 > "$LOG_FILE" 2>&1 &
echo $! > "$PID_FILE"

# Wait for server to start
echo "Waiting for server to start..."
sleep 5

# Test 1: Check if web server is responding
echo "Test 1: Web server responsiveness..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$WEB_PORT/")
if [[ "$HTTP_CODE" == "200" ]]; then
    echo "✓ Web server is responding with HTTP 200"
else
    echo "✗ Web server is not responding properly"
    cat "$LOG_FILE"
    exit 1
fi

# Test 2: Check if homepage contains expected content
echo "Test 2: Homepage content..."
HOMEPAGE_CONTENT=$(curl -s "http://localhost:$WEB_PORT/")
if [[ "$HOMEPAGE_CONTENT" == *"EvoSim"* ]]; then
    echo "✓ Homepage contains 'EvoSim' title"
else
    echo "✗ Homepage does not contain expected title"
    echo "Content received:"
    echo "$HOMEPAGE_CONTENT"
    exit 1
fi

# Test 3: Check for essential HTML elements
echo "Test 3: HTML structure..."
if [[ "$HOMEPAGE_CONTENT" == *"simulation-view"* ]] && \
   [[ "$HOMEPAGE_CONTENT" == *"info-panel"* ]] && \
   [[ "$HOMEPAGE_CONTENT" == *"controls"* ]]; then
    echo "✓ Essential HTML elements are present"
else
    echo "✗ Missing essential HTML elements"
    echo "Checking for: simulation-view, info-panel, controls"
    exit 1
fi

# Test 4: Check API status endpoint
echo "Test 4: API status endpoint..."
API_RESPONSE=$(curl -s "http://localhost:$WEB_PORT/api/status")
if [[ "$API_RESPONSE" == *'"status"'* ]]; then
    echo "✓ API status endpoint is working"
else
    echo "✗ API status endpoint is not working properly"
    echo "Response: $API_RESPONSE"
    exit 1
fi

# Test 5: Check if WebSocket endpoint is available
echo "Test 5: WebSocket endpoint availability..."
# We can't easily test WebSocket with curl, but we can check if the endpoint exists
WS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$WEB_PORT/ws")
if [ "$WS_RESPONSE" != "000" ]; then
    echo "✓ WebSocket endpoint is available (HTTP $WS_RESPONSE)"
else
    echo "✗ WebSocket endpoint is not available"
    exit 1
fi

# Test 6: Check for JavaScript functionality
echo "Test 6: JavaScript files and functionality..."
if [[ "$HOMEPAGE_CONTENT" == *"WebSocket"* ]] && \
   [[ "$HOMEPAGE_CONTENT" == *"connect"* ]]; then
    echo "✓ JavaScript WebSocket functionality is present"
else
    echo "✗ JavaScript WebSocket functionality is missing"
    exit 1
fi

# Test 7: Check server logs for errors
echo "Test 7: Server error checking..."
if grep -q "ERROR\|FATAL\|panic" "$LOG_FILE"; then
    echo "✗ Server logs contain errors:"
    grep "ERROR\|FATAL\|panic" "$LOG_FILE"
    exit 1
else
    echo "✓ No critical errors in server logs"
fi

# Test 8: Performance test - multiple requests
echo "Test 8: Performance test..."
start_time=$(date +%s%N)
for i in {1..10}; do
    curl -s -o /dev/null "http://localhost:$WEB_PORT/"
done
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
echo "✓ Completed 10 requests in ${duration}ms (avg: $((duration/10))ms per request)"

# Test 9: Run Playwright tests
echo "Test 9: Running Playwright end-to-end tests..."
echo "Server is running on port $WEB_PORT, executing playwright tests..."

# Set environment variable to prevent playwright from starting its own server
export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0
export CI=true

# Run playwright tests with timeout
timeout 300s npm test
PLAYWRIGHT_EXIT_CODE=$?

if [ $PLAYWRIGHT_EXIT_CODE -eq 0 ]; then
    echo "✓ All Playwright tests passed"
elif [ $PLAYWRIGHT_EXIT_CODE -eq 124 ]; then
    echo "✗ Playwright tests timed out after 5 minutes"
    exit 1
else
    echo "✗ Playwright tests failed with exit code $PLAYWRIGHT_EXIT_CODE"
    exit 1
fi

echo ""
echo "All web interface tests passed! ✓"
echo "Web interface is working correctly on port $WEB_PORT"