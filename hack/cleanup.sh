#!/usr/bin/env bash
set -euo pipefail
CLUSTERS=${KFLEET_CLUSTERS:-3}
PF_PID_FILE="/tmp/kfleet-pf.pid"

echo "Cleaning up kfleet..."
if [ -f "$PF_PID_FILE" ]; then
    kill "$(cat "$PF_PID_FILE")" 2>/dev/null || true
    rm -f "$PF_PID_FILE"
    echo "Port-forward stopped."
fi
for i in $(seq 1 "$CLUSTERS"); do
    kubectl config use-context "kind-kfleet-$i" 2>/dev/null || true
    helm uninstall kfleet-agent 2>/dev/null || true
    helm uninstall kfleet-hub 2>/dev/null || true
done
for i in $(seq 1 "$CLUSTERS"); do
    kind delete cluster --name "kfleet-$i" 2>/dev/null || true
done
echo "Cleanup complete."
