#!/bin/sh
set -e

MODEL="nomic-embed-text"

# Start Ollama server in background.
ollama serve &
SERVER_PID=$!

# Wait for the server to be ready.
echo "Waiting for Ollama server..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:11434/api/tags >/dev/null 2>&1; then
    echo "Ollama server ready."
    break
  fi
  sleep 1
done

# Pull the model if not already present.
if ! ollama list | grep -q "$MODEL"; then
  echo "Pulling model: $MODEL ..."
  ollama pull "$MODEL"
  echo "Model $MODEL pulled successfully."
else
  echo "Model $MODEL already present."
fi

# Stop the background server.
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

# Start Ollama as PID 1.
exec ollama serve
