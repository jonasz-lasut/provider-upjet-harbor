#!/usr/bin/env bash
set -aeuo pipefail

echo "Running setup.sh"
echo "Creating cloud credential secret..."
${KUBECTL} -n crossplane-system create secret generic provider-secret --from-literal=credentials="${UPTEST_CLOUD_CREDENTIALS}" --dry-run=client -o yaml | ${KUBECTL} apply -f -

echo "Waiting until provider is healthy..."
${KUBECTL} wait provider.pkg --all --for condition=Healthy --timeout 5m

echo "Waiting for all pods to come online..."
${KUBECTL} -n crossplane-system wait --for=condition=Available deployment --all --timeout=5m

echo "Installing Harbor via Helm..."
${HELM} repo add harbor https://helm.goharbor.io
${HELM} repo update
${HELM} upgrade --install harbor harbor/harbor \
  --namespace harbor \
  --create-namespace \
  --set expose.type=clusterIP \
  --set expose.tls.enabled=false \
  --set externalURL=http://harbor.harbor.svc.cluster.local \
  --set harborAdminPassword=Harbor12345 \
  --set persistence.enabled=false \
  --wait \
  --atomic

echo "Creating Harbor provider credentials secret..."
${KUBECTL} -n crossplane-system create secret generic harbor-credentials \
  --from-literal=credentials='{"username":"admin","password":"Harbor12345"}' \
  --dry-run=client -o yaml | ${KUBECTL} apply -f -

echo "Creating a default cluster provider config (v2-style)..."
cat <<EOF | ${KUBECTL} apply -f -
apiVersion: harbor.m.crossplane.io/v1beta1
kind: ClusterProviderConfig
metadata:
  name: default
spec:
  url: http://harbor.harbor.svc.cluster.local
  credentials:
    source: Secret
    secretRef:
      name: harbor-credentials
      namespace: crossplane-system
      key: credentials
EOF

${KUBECTL} wait provider.pkg --all --for condition=Healthy --timeout 5m
${KUBECTL} -n crossplane-system wait --for=condition=Available deployment --all --timeout=5m
