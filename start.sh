#!/bin/bash

# NAS AI Advisor Startup Script

set -e

echo "Starting NAS AI Advisor..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Check if Ollama is running
if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "Warning: Ollama is not running. Please start Ollama first."
    echo "You can start it with: ollama serve"
fi

# Build the application
echo "Building application..."
go build -o server ./cmd/server

# Create config if it doesn't exist
if [ ! -f config.json ]; then
    echo "Creating default configuration..."
    cat > config.json << EOF
{
  "llm": {
    "base_url": "http://localhost:11434",
    "model": "llama2",
    "timeout": 30,
    "max_tokens": 2048,
    "temperature": 0.7
  },
  "knowledge_path": "./knowledge",
  "max_messages": 50,
  "cleanup_interval": 60,
  "max_conversation_age": 24
}
EOF
fi

# Run the server
echo "Starting server on http://localhost:8080"
./server -config config.json
