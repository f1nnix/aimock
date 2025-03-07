#!/bin/bash

# Test script for malformed requests to OpenAI API Mock Service

echo "Testing OpenAI API Mock Service with malformed requests..."
echo

# Test the chat completions endpoint with malformed JSON
echo "Testing /v1/chat/completions endpoint with malformed JSON..."
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
  ' | jq .
echo

# Test the chat completions endpoint with missing required fields
echo "Testing /v1/chat/completions endpoint with missing required fields..."
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo"
  }' | jq .
echo

# Test the embeddings endpoint with malformed JSON
echo "Testing /v1/embeddings endpoint with malformed JSON..."
curl -s -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-ada-002",
    "input": ["The food was delicious and the service was excellent."
  }' | jq .
echo

# Test the embeddings endpoint with missing required fields
echo "Testing /v1/embeddings endpoint with missing required fields..."
curl -s -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-ada-002"
  }' | jq .
echo

echo "All malformed request tests completed!"