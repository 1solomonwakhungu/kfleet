#!/usr/bin/env bash
set -euo pipefail

CLUSTERS=${KFLEET_CLUSTERS:-3}
USE_GHCR=${USE_GHCR:-0}
IMAGE_TAG=${IMAGE_TAG:-dev}
HUB_PORT=${KFLEET_HUB_PORT:-8080}
TOKEN=$(openssl rand -hex 16)
ADMIN_USERNAME=${KFLEET_ADMIN_USERNAME:-admin}
ADMIN_EMAIL=${KFLEET_ADMIN_EMAIL:-admin@kfleet.local}
ADMIN_PASSWORD=${KFLEET_ADMIN_PASSWORD:-$(openssl rand -hex 16)}
PF_PID_FILE="/tmp/kfleet-pf.pid"
COOKIE_FILE="/tmp/kfleet-quickstart-cookie.txt"

echo "Checking prerequisites..."
for cmd in kind kubectl helm docker; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "ERROR: $cmd not found. Please install $cmd first."
        exit 1
    fi
done
echo "All prerequisites found."

for i in $(seq 1 "$CLUSTERS"); do
    if kind get clusters 2>/dev/null | grep -q "kfleet-$i"; then
        echo "Cluster kfleet-$i already exists, skipping."
    else
        echo "Creating kind cluster kfleet-$i..."
        kind create cluster --name "kfleet-$i"
    fi
done

if [ "$USE_GHCR" = "1" ]; then
    HUB_IMAGE="ghcr.io/1solomonwakhungu/kfleet-hub:${IMAGE_TAG}"
    AGENT_IMAGE="ghcr.io/1solomonwakhungu/kfleet-agent:${IMAGE_TAG}"
else
    HUB_IMAGE="kfleet-hub:${IMAGE_TAG}"
    AGENT_IMAGE="kfleet-agent:${IMAGE_TAG}"
    echo "Building local images..."
    docker build -f Dockerfile.hub -t "$HUB_IMAGE" .
    docker build -f Dockerfile.agent -t "$AGENT_IMAGE" .
fi

for i in $(seq 1 "$CLUSTERS"); do
    if [ "$USE_GHCR" = "0" ]; then
        kind load docker-image "$HUB_IMAGE" --name "kfleet-$i"
        kind load docker-image "$AGENT_IMAGE" --name "kfleet-$i"
    fi
done

echo "Installing hub on kfleet-1..."
kubectl config use-context kind-kfleet-1
helm upgrade --install kfleet-hub charts/kfleet-hub \
    --set image.repository="${HUB_IMAGE%:*}" \
    --set image.tag="$IMAGE_TAG" \
    --set service.type=ClusterIP \
    --set persistence.enabled=true \
    --set registration.token="$TOKEN" \
    --set auth.sessionCookieInsecure=true \
    --set-string auth.bootstrapAdmin.username="$ADMIN_USERNAME" \
    --set-string auth.bootstrapAdmin.email="$ADMIN_EMAIL" \
    --set-string auth.bootstrapAdmin.password="$ADMIN_PASSWORD" \
    --wait --timeout 120s

kubectl rollout status deployment/kfleet-hub --timeout=120s

# Start port-forward (background process)
if [ -f "$PF_PID_FILE" ]; then
    kill "$(cat "$PF_PID_FILE")" 2>/dev/null || true
fi
kubectl port-forward svc/kfleet-hub "${HUB_PORT}:8080" > /dev/null 2>&1 &
echo $! > "$PF_PID_FILE"
sleep 3

HUB_URL="http://host.docker.internal:${HUB_PORT}"

curl --fail --silent --show-error \
    --cookie-jar "$COOKIE_FILE" \
    --header "Content-Type: application/json" \
    --data "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}" \
    "http://localhost:${HUB_PORT}/api/v1/auth/login" >/dev/null

for i in $(seq 1 "$CLUSTERS"); do
    echo "Installing agent on kfleet-$i..."
    kubectl config use-context "kind-kfleet-$i"
    helm upgrade --install kfleet-agent charts/kfleet-agent \
        --set image.repository="${AGENT_IMAGE%:*}" \
        --set image.tag="$IMAGE_TAG" \
        --set hub.url="$HUB_URL" \
        --set hub.token="$TOKEN" \
        --set cluster.name="kfleet-$i" \
        --wait --timeout 60s
done

echo "Waiting for agents to register..."
for attempt in $(seq 1 24); do
    sleep 5
    REGISTERED=$(curl -s --cookie "$COOKIE_FILE" "http://localhost:${HUB_PORT}/api/v1/clusters" 2>/dev/null | grep -o '"id"' | wc -l)
    echo "  Attempt $attempt: $REGISTERED/$CLUSTERS clusters registered"
    if [ "$REGISTERED" -ge "$CLUSTERS" ]; then
        break
    fi
done

echo ""
echo "=============================================="
echo " kfleet is ready!"
echo " Open http://localhost:${HUB_PORT}"
echo " Username: ${ADMIN_USERNAME}"
echo " Password: ${ADMIN_PASSWORD}"
echo "=============================================="
echo "Cleanup: ./hack/cleanup.sh"
