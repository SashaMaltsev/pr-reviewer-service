#!/bin/bash

# End-to-end test runner for PR Reviewer Service
# Starts dependencies, runs tests, and cleans up

set -e

echo "Starting E2E tests..."

# Start services
docker-compose up -d

# Wait for services
echo "Waiting for services to be ready..."
sleep 10

# Run E2E tests
export API_BASE_URL="http://localhost:8080"
go test -v ./tests/e2e/...

# Cleanup
docker-compose down

echo "E2E tests completed!"