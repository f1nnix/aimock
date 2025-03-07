#!/bin/bash

# Test script for OpenAI API Mock Service

echo "Testing OpenAI API Mock Service..."
echo

# Test the models endpoint
echo "Testing /v1/models endpoint..."
curl -s http://localhost:8080/v1/models | jq .
echo

# Test the chat completions endpoint
echo "Testing /v1/chat/completions endpoint..."
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }' | jq .
echo

# Test the embeddings endpoint
echo "Testing /v1/embeddings endpoint..."
curl -s -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-ada-002",
    "input": ["The food was delicious and the service was excellent."]
  }' | jq '.data[0].embedding | length'
echo

echo "All tests completed!"