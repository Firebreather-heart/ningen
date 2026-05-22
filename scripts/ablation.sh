#!/usr/bin/env bash
# ablation.sh — runs all 6 ablation variants and writes results to OUT_FILE.
#
# Usage:
#   ./scripts/ablation.sh                         # stdout only
#   OUT_FILE=ablation_results.txt ./scripts/ablation.sh   # save to file
#   nohup OUT_FILE=ablation_results.txt ./scripts/ablation.sh &  # background
#
# Env vars (all optional):
#   PROVIDER         gemini          LLM backend
#   SEEDS_PER_DOMAIN 5               seeds per domain per variant
#   GT_THRESHOLD     0.45            cosine distance ground-truth threshold
#   DB_URL           postgres://...  override DB connection
#   EMBEDDER_URL     http://...      override embedder sidecar
#   API_URL          http://...      override API URL
#   OUT_FILE                         path to write results (appended)

set -euo pipefail

PROVIDER="${PROVIDER:-gemini}"
SEEDS="${SEEDS_PER_DOMAIN:-5}"
GT="${GT_THRESHOLD:-0.45}"
OUT="${OUT_FILE:-}"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

log() { echo "$*" | tee -a "${OUT:-/dev/null}"; }

run_variant() {
  local label="$1"
  local skip="$2"

  log ""
  log "════════════════════════════════════════════════════════════"
  log "VARIANT: ${label}  skip=${skip:-none}"
  log "════════════════════════════════════════════════════════════"

  PROVIDER="$PROVIDER" \
  SEEDS_PER_DOMAIN="$SEEDS" \
  GT_THRESHOLD="$GT" \
  SKIP_STAGES="$skip" \
    go run ./cmd/holdout_eval 2>&1 | tee -a "${OUT:-/dev/null}"
}

log "ablation.sh  started=${TIMESTAMP}  provider=${PROVIDER}  seeds=${SEEDS}  gt=${GT}"
log "out=${OUT:-stdout only}"

run_variant "A: Full SIGNAL pipeline"          ""
run_variant "B: Pure RAG (no Gate, no Reranker)" "gate,reranker"
run_variant "C: RAG + Reranker (no Gate)"      "gate"
run_variant "D: RAG + Gate (no Reranker)"      "reranker"
run_variant "E: Single-query (no multi-vector)" "multi_vector"
run_variant "F: No corpus grounding"           "grounding"

log ""
log "════════════════════════════════════════════════════════════"
log "ablation.sh  done  $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
log "════════════════════════════════════════════════════════════"
