# Falzo on Kubernetes

> Archived deployment option. Production now uses Docker Compose through
> `.github/workflows/deploy-docker.yml`; these manifests are not run by CI/CD.

This directory deploys Falzo in three explicit phases:

1. `platform`: namespace, configuration, PostgreSQL and Redis.
2. `migration`: one-shot database migration Job.
3. `app`: two backend Pods, two frontend Pods, Services and Ingress.

The staged flow prevents a new application release from serving traffic before
its database migration has completed.

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- `kubectl` with access to the cluster.
- A Traefik ingress controller with `web` and `websecure` entrypoints.
- A container registry reachable by the cluster.
- A TLS Secret or cert-manager certificate named `falzo-tls`.

For production, prefer managed PostgreSQL and Redis. The bundled single-replica
StatefulSets are appropriate for development, staging, or a small self-hosted
cluster, but they are not a highly available data tier.

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

- The public host is configured as `falzo.life` in `platform/configmap.yaml`
  and `app/ingress.yaml`.
- `GOOGLE_CLIENT_ID` in `platform/configmap.yaml`.
- CPU, memory and PVC sizes for the target cluster.
- `storageClassName` in each `volumeClaimTemplates` if the cluster has no
  default StorageClass.

The WebSocket origin pattern must be the public host without `https://`.

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
Traefik. On a bare-metal cluster, a `LoadBalancer` implementation such as
MetalLB or the K3s ServiceLB is also required:

```bash
kubectl get service -A | grep LoadBalancer
kubectl get ingress falzo -n falzo
```

The TLS Secret `falzo-tls` must contain a certificate valid for `falzo.life`.
It can be created manually as shown later in this guide or maintained by
cert-manager. In Google Cloud Console, add `https://falzo.life` to the OAuth
client's authorized JavaScript origins.

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

Create TLS using cert-manager or an existing certificate. A manual example is:

```bash
kubectl create secret tls falzo-tls \
  --cert=/path/to/fullchain.pem \
  --key=/path/to/private-key.pem \
  -n falzo
```

Then deploy and wait for both rollouts:

```bash
kubectl apply -k deploy/k8s/app
kubectl rollout status deployment/falzo-backend -n falzo --timeout=180s
kubectl rollout status deployment/falzo-frontend -n falzo --timeout=180s
kubectl get pods,svc,ingress -n falzo
```

Ingress sends `/api` traffic, including WebSocket upgrades, directly to the Go
backend. All other paths go to Next.js. The backend's Redis backplane lets both
backend replicas share room events and online presence.

## Health checks and diagnostics

The backend exposes:

- `/health/live`: process liveness only.
- `/health/ready`: verifies PostgreSQL and Redis within a two-second timeout.

Useful commands:

```bash
kubectl logs -f deployment/falzo-backend -n falzo
kubectl logs -f deployment/falzo-frontend -n falzo
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

These manifests are retained as a manual Kubernetes reference only. GitHub
Actions no longer applies them. Production deployment now uses Docker Compose;
see `deploy/docker/README.md` and `.github/workflows/deploy-docker.yml`.
