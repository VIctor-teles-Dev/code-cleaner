#!/usr/bin/env bash
# Deploy no cluster k3s da VPS: builda as imagens, importa no containerd do
# k3s, aplica os manifests e reinicia os deployments. Idempotente — pode rodar
# quantas vezes for preciso. Chamado pelo workflow de CD (runner self-hosted)
# ou à mão na VPS.
#
# Requisitos na VPS: docker, k3s instalado, e `sudo` sem senha para o `k3s`.
set -euo pipefail

cd "$(dirname "$0")/.."

echo ">> build das imagens"
docker build -t ccl/backend-api:dev apps/backend-api
docker build -t ccl/url-shortener:dev apps/url-shortener
docker build -f apps/web-app/Dockerfile -t ccl/web-app:dev .

echo ">> importa as imagens no containerd do k3s"
for img in backend-api url-shortener web-app; do
  docker save "ccl/${img}:dev" | sudo k3s ctr images import -
done

echo ">> aplica os manifests"
sudo k3s kubectl apply -k infra/k8s

echo ">> rollout (força pods novos com as imagens recém-importadas)"
for dep in backend-api url-shortener web-app; do
  sudo k3s kubectl -n ccl rollout restart "deployment/${dep}"
done
for dep in backend-api url-shortener web-app; do
  sudo k3s kubectl -n ccl rollout status "deployment/${dep}" --timeout=180s
done

echo ">> deploy concluído"
