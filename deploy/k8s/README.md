# Falzo on Kubernetes

Production deploys to a single-node K3s VPS through
`.github/workflows/deploy-k3s.yml`. The workflow builds immutable images in
GHCR, runs the migration Job and only then rolls out the application.

This directory deploys Falzo in three explicit phases:

1. `platform`: namespace, configuration, PostgreSQL, Redis and NATS JetStream.
2. `migration`: one-shot database migration Job.
3. `app`: backend, Telegram alert bot, frontend, Services and Ingress.

The staged flow prevents a new application release from serving traffic before
its database migration has completed.

## GitHub Actions configuration

Create a GitHub Environment named `production`. Configure an approval rule for
it when the repository plan supports protected environments. Add these
environment secrets:

- `VPS_HOST`
- `VPS_USER`
- `VPS_PORT` (optional; defaults to `22`)
- `VPS_SSH_KEY`
- `VPS_APP_DIR`, for example `/opt/falzo`
- `GHCR_PULL_USERNAME`
- `GHCR_PULL_TOKEN` with `read:packages`

Add the repository variable `NEXT_PUBLIC_GOOGLE_CLIENT_ID`. The frontend build
embeds this public OAuth client ID, while the deployment also writes it into the
backend ConfigMap in the immutable release bundle.

The deploy user must be able to write to `VPS_APP_DIR` and run `k3s kubectl`
either as root or through passwordless `sudo`. The workflow deliberately uses
the Kubernetes API locally over SSH, so port `6443` does not need to be exposed
to GitHub-hosted runners.

Before the first automated deployment, create the `falzo-secrets` Secret in
namespace `falzo`. The workflow installs pinned cert-manager `v1.20.3`, obtains
and renews the public `falzo-tls` certificate through Let's Encrypt HTTP-01,
checks the application Secret before changing the platform, and safely creates
or rotates `ghcr-pull` from the GitHub Environment credentials. It never
uploads production database, Redis, JWT or Telegram credentials from GitHub.
The Secret must contain `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` in addition
to the keys shown in `secret.example.yaml`.

Pushes to `main` that touch the backend, frontend, K3s manifests or the K3s
workflow build and deploy all three images under the Git commit SHA. Manual
runs are available through `workflow_dispatch`. The legacy Docker workflow is
manual-only and requires typing `DEPLOY_DOCKER` because running it stops K3s and
claims ports 80 and 443 for Caddy.

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- `kubectl` with access to the cluster.
- A Traefik ingress controller with `web` and `websecure` entrypoints.
- A container registry reachable by the cluster.
- Public DNS for both Falzo hosts and inbound ports 80/443 for the automated
  Let's Encrypt HTTP-01 certificate.

For production, prefer managed PostgreSQL, Redis and NATS. The bundled single-replica
StatefulSets are appropriate for development, staging, or a small self-hosted
cluster, but they are not a highly available data tier.

### Small VPS profile

The checked-in resource profile targets a small single-node K3s VPS with 2
vCPUs and 4 GiB RAM. It runs one backend Pod and one frontend Pod so Kubernetes,
Traefik, PostgreSQL and Redis retain enough memory headroom. This profile is not
highly available: losing the VPS takes down every replica and its local
persistent volumes. Increase the replica counts only after moving to a larger
node or a multi-node cluster and verifying the realtime Redis backplane across
backend replicas.

## 1. Build and push Linux images

Replace the registry and tag with values owned by your deployment:

```bash
export FALZO_REGISTRY=ghcr.io/your-org
export FALZO_TAG=$(git rev-parse --short HEAD)

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t "$FALZO_REGISTRY/falzo-backend:$FALZO_TAG" \
  --push be

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f be/Dockerfile.migrate \
  -t "$FALZO_REGISTRY/falzo-migrate:$FALZO_TAG" \
  --push be

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg BACKEND_URL=http://falzo-backend:8080 \
  --build-arg NEXT_PUBLIC_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com \
  -t "$FALZO_REGISTRY/falzo-frontend:$FALZO_TAG" \
  --push fe
```

`NEXT_PUBLIC_GOOGLE_CLIENT_ID` is embedded into the browser bundle during the
frontend build. Changing it requires a new frontend image.

Update `app/kustomization.yaml` and `migration/kustomization.yaml` with the
registry and immutable tag that were pushed. Avoid `latest` for a real release.

## 2. Configure host and non-secret values

Review these values before deployment:

- The public hosts are configured as `falzo.life` and `www.falzo.life` in
  `platform/configmap.yaml` and `app/ingress.yaml`.
- `GOOGLE_CLIENT_ID` in `platform/configmap.yaml`.
- CPU, memory and PVC sizes for the target cluster.
- `storageClassName` in each `volumeClaimTemplates` if the cluster has no
  default StorageClass.

The WebSocket origin patterns must contain every public host without
`https://`, separated by commas.

## Edge routing for falzo.life

Do not run a separate Nginx container inside the Falzo application. Install one
ingress controller for the cluster. Falzo uses the actively maintained Traefik
controller because the community ingress-nginx controller was retired in March
2026.

K3s normally includes Traefik. Verify it with:

```bash
kubectl get ingressclass
kubectl get pods,service -n kube-system -l app.kubernetes.io/name=traefik
```

For a cluster without an ingress controller, install Traefik once as a cluster
administrator. Pin the chart version selected for the cluster:

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm search repo traefik/traefik --versions

helm upgrade --install traefik traefik/traefik \
  --namespace traefik \
  --create-namespace \
  --version REPLACE_WITH_PINNED_VERSION \
  --set providers.kubernetesIngress.enabled=true \
  --set ports.web.redirections.entryPoint.to=websecure \
  --set ports.web.redirections.entryPoint.scheme=https
```

Point the DNS `A` record for `falzo.life` to the external IP assigned to
Traefik, and point `www.falzo.life` to it with either a `CNAME` or another `A`
record. On a bare-metal cluster, a `LoadBalancer` implementation such as
MetalLB or the K3s ServiceLB is also required:

```bash
kubectl get service -A | grep LoadBalancer
kubectl get ingress falzo -n falzo
```

The workflow creates and maintains the TLS Secret `falzo-tls` with a certificate
valid for both `falzo.life` and `www.falzo.life`. In Google Cloud Console, add
`https://falzo.life` and `https://www.falzo.life` to the OAuth client's
authorized JavaScript origins.

Traefik routes `/api`, including WebSocket upgrades, to the Go service. All
other paths are routed to Next.js. The production frontend derives `wss://`
from the browser origin, so no separate public WebSocket domain is required.

## 3. Create secrets

Never commit the production Secret. Copy the example outside the repository,
replace every value, and URL-encode reserved characters in the password inside
`DATABASE_URL`.

```bash
kubectl apply -f deploy/k8s/platform/namespace.yaml
cp deploy/k8s/secret.example.yaml /tmp/falzo-secret.yaml
# Edit /tmp/falzo-secret.yaml before applying it.
kubectl apply -f /tmp/falzo-secret.yaml
```

For a maintained cluster, use External Secrets, Sealed Secrets, SOPS, or the
cloud provider's secret manager instead of a plaintext Secret file.

If the registry is private, create an image pull Secret and add
`imagePullSecrets` to the migration and application Pod specs.

## 4. Deploy the platform

```bash
kubectl apply -k deploy/k8s/platform
kubectl rollout status statefulset/falzo-postgres -n falzo --timeout=180s
kubectl rollout status statefulset/falzo-redis -n falzo --timeout=180s
kubectl rollout status statefulset/falzo-nats -n falzo --timeout=180s
```

PostgreSQL 18 persists its versioned `PGDATA` under the mounted
`/var/lib/postgresql` parent directory.

## 5. Run migrations

The Job is intentionally separate from the backend Deployment. Delete the old
completed Job before each release, apply the new image tag, and wait for it to
finish:

```bash
kubectl delete job falzo-migrate -n falzo --ignore-not-found
kubectl apply -k deploy/k8s/migration
kubectl wait \
  --for=condition=complete \
  job/falzo-migrate \
  -n falzo \
  --timeout=300s
```

If the Job fails:

```bash
kubectl logs job/falzo-migrate -n falzo
kubectl describe job falzo-migrate -n falzo
```

Do not deploy the new backend until the migration Job is complete.

## 6. Deploy the application

The deployment workflow normally installs cert-manager and creates `falzo-tls`
automatically. To run the same TLS phase manually from the repository:

```bash
kubectl apply -f \
  https://github.com/cert-manager/cert-manager/releases/download/v1.20.3/cert-manager.yaml
kubectl wait --for=condition=Available deployment --all \
  -n cert-manager --timeout=180s
kubectl apply -f deploy/k8s/tls/cluster-issuer.yaml
kubectl apply -f deploy/k8s/tls/certificate.yaml
kubectl wait --for=condition=Ready certificate/falzo-tls \
  -n falzo --timeout=300s
```

For manual recovery with an existing certificate instead, run:

```bash
kubectl create secret tls falzo-tls \
  --cert=/path/to/fullchain.pem \
  --key=/path/to/private-key.pem \
  -n falzo
```

Then deploy and wait for all application rollouts:

```bash
kubectl apply -k deploy/k8s/app
kubectl rollout status deployment/falzo-backend -n falzo --timeout=180s
kubectl rollout status deployment/falzo-frontend -n falzo --timeout=180s
kubectl rollout status deployment/falzo-telegram-bot -n falzo --timeout=180s
kubectl get pods,svc,ingress -n falzo
```

Ingress sends `/api` traffic, including WebSocket upgrades, directly to the Go
backend. All other paths go to Next.js. The backend's Redis backplane supports
sharing room events and online presence when the backend is scaled above one
replica.

## Health checks and diagnostics

The backend exposes:

- `/health/live`: process liveness only.
- `/health/ready`: verifies PostgreSQL and Redis within a two-second timeout.

Useful commands:

```bash
kubectl logs -f deployment/falzo-backend -n falzo
kubectl logs -f deployment/falzo-frontend -n falzo
kubectl logs -f deployment/falzo-telegram-bot -n falzo
kubectl describe ingress falzo -n falzo
kubectl port-forward service/falzo-frontend 3000:3000 -n falzo
kubectl port-forward service/falzo-backend 8080:8080 -n falzo
```

Render manifests locally before applying changes:

```bash
kubectl kustomize deploy/k8s/platform >/dev/null
kubectl kustomize deploy/k8s/migration >/dev/null
kubectl kustomize deploy/k8s/app >/dev/null
kubectl diff -k deploy/k8s/app
```

## Deployment status

`.github/workflows/deploy-k3s.yml` is the active production deployment. The
manual commands in this guide remain useful for first-time bootstrap,
diagnostics and recovery when GitHub Actions is unavailable.
