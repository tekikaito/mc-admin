#!/bin/bash

# Script to set up Kubernetes port forwarding to Minecraft pod and start the Go server

set -e  # Exit on error

# Configuration
NAMESPACE="${K8S_NAMESPACE:-mc-red}"
POD_SELECTOR="${K8S_POD_SELECTOR:-app=mc-red-minecraft}"
RCON_PORT="${RCON_PORT:-25575}"
LOCAL_PORT="${LOCAL_PORT:-25575}"

echo "🔍 Finding Minecraft pod in namespace: $NAMESPACE"
POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l "$POD_SELECTOR" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -z "$POD_NAME" ]; then
    echo "❌ Error: No pod found with selector '$POD_SELECTOR' in namespace '$NAMESPACE'"
    echo "   You can set K8S_NAMESPACE and K8S_POD_SELECTOR environment variables to customize"
    exit 1
fi

echo "✅ Found pod: $POD_NAME"

# Cleanup function to kill port-forward on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up port-forward..."
    if [ ! -z "$PORT_FORWARD_PID" ]; then
        kill $PORT_FORWARD_PID 2>/dev/null || true
    fi
    exit
}

trap cleanup SIGINT SIGTERM EXIT

# Start port forwarding in the background
echo "🔌 Starting port-forward: localhost:$LOCAL_PORT -> $POD_NAME:$RCON_PORT"
kubectl port-forward -n "$NAMESPACE" "$POD_NAME" "$LOCAL_PORT:$RCON_PORT" &
PORT_FORWARD_PID=$!

# Wait for port-forward to be ready
echo "⏳ Waiting for port-forward to be ready..."
sleep 2

# Check if port-forward is still running
if ! kill -0 $PORT_FORWARD_PID 2>/dev/null; then
    echo "❌ Error: Port-forward failed to start"
    exit 1
fi

echo "✅ Port-forward established"

# Start the Go server
echo "🚀 Starting Go server..."
echo ""
go run main.go
