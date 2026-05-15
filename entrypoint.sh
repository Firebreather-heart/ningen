#!/bin/bash
set -e

echo "Starting ETL Worker Entrypoint..."

# Pull the lightweight embedding model
echo "Pulling all-minilm model via Ollama..."
curl -X POST "$OLLAMA_URL/api/pull" -d '{"name": "all-minilm"}'

echo "Model pulled successfully. Starting ETL pipeline..."
exec /app/etl-worker
