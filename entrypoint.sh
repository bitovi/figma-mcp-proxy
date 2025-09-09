#!/bin/sh
set -e

/usr/local/bin/figma-mcp-proxy "$@" &
PROXY_PID=$!

/opt/figma/app/figma --no-sandbox --disable-gpu --disable-dev-shm-usage &
FIGMA_PID=$!

term() {
  kill -TERM "$PROXY_PID" "$FIGMA_PID" 2>/dev/null || true
}

trap term INT TERM

# Wait until either exits
while kill -0 "$PROXY_PID" 2>/dev/null && kill -0 "$FIGMA_PID" 2>/dev/null; do
  sleep 1
done

# One exited: stop the other and wait them out
term
wait "$PROXY_PID" 2>/dev/null || true
wait "$FIGMA_PID" 2>/dev/null || true