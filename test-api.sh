#!/bin/bash

# API Test Script for NAS AI Advisor

BASE_URL="http://localhost:8080"

echo "Testing NAS AI Advisor API..."
echo "=============================="

# Test health endpoint
echo -e "\n1. Testing Health Check..."
curl -s "$BASE_URL/health" | jq .

# Test status endpoint
echo -e "\n2. Testing System Status..."
curl -s "$BASE_URL/api/v1/status" | jq .

# Test query endpoint
echo -e "\n3. Testing Query..."
curl -s -X POST "$BASE_URL/api/v1/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I set up RAID?",
    "user_id": "test-user"
  }' | jq .

# Test recommendations endpoint
echo -e "\n4. Testing Recommendations..."
curl -s -X POST "$BASE_URL/api/v1/recommendations" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user",
    "limit": 3
  }' | jq .

# Test knowledge articles
echo -e "\n5. Testing Knowledge Articles..."
curl -s "$BASE_URL/api/v1/knowledge/articles" | jq .

# Test categories
echo -e "\n6. Testing Categories..."
curl -s "$BASE_URL/api/v1/knowledge/categories" | jq .

echo -e "\n=============================="
echo "API testing complete!"
