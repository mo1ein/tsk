#!/bin/bash

BASE_URL="http://localhost:8080"
CONCURRENT=10
REQUESTS=100

echo "=== Task Manager Load Test ==="
echo "Base URL: $BASE_URL"
echo "Concurrent: $CONCURRENT"
echo "Total Requests: $REQUESTS"
echo ""

echo "1. Creating tasks..."
for i in $(seq 1 20); do
  curl -s -X POST "$BASE_URL/tasks" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"Task $i\",\"assignee\":\"user$((i % 5))\"}" > /dev/null
done
echo "Created 20 tasks"

echo ""
echo "2. Listing tasks (benchmark)..."
ab -n $REQUESTS -c $CONCURRENT "$BASE_URL/tasks"

echo ""
echo "3. Getting single task (benchmark)..."
ab -n $REQUESTS -c $CONCURRENT "$BASE_URL/tasks/1"

echo ""
echo "4. pprof CPU profile (5 seconds)..."
curl -s "http://localhost:8080/debug/pprof/profile?seconds=5" > /tmp/cpu.prof 2>/dev/null
if [ -f /tmp/cpu.prof ]; then
  echo "CPU profile saved to /tmp/cpu.prof"
  echo "Analyze with: go tool pprof -http=:8081 /tmp/cpu.prof"
else
  echo "pprof endpoint not available (enable with _ import net/http/pprof)"
fi

echo ""
echo "=== Benchmark Complete ==="
