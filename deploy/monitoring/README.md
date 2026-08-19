# Falzo monitoring

Falzo exposes Prometheus metrics at `GET /metrics` on the backend's internal
HTTP port. The public application Ingress deliberately does not route this
path. Prometheus scrapes the `falzo-backend` ClusterIP Service and Grafana is
the only monitoring component exposed publicly, at
`https://monitor.falzo.life`.

The checked-in profile targets the existing single-node K3s VPS:

- Prometheus retention: 7 days, with a 10 GiB PVC.
- Grafana storage: 2 GiB PVC.
- Prometheus, its Operator and Grafana remain `ClusterIP` Services.
- Alertmanager and unreachable K3s control-plane monitors are disabled to keep
  the memory footprint and false-positive noise down.
- The `Falzo Overview` dashboard and Falzo alert rules are provisioned from
  Kubernetes resources in this directory.

## Prerequisites

- Helm and `kubectl` can access the K3s cluster.
- The cluster has the `local-path` StorageClass (the K3s default).
- Traefik and cert-manager are installed.
- The `letsencrypt-production` ClusterIssuer from `deploy/k8s/tls` is Ready.
- DNS contains an `A` record from `monitor.falzo.life` to the VPS public IP.

Verify the prerequisites:

```bash
kubectl get nodes
kubectl get storageclass local-path
kubectl get ingressclass traefik
kubectl get clusterissuer letsencrypt-production
```

## Install

Create the namespace first:

```bash
kubectl apply -f deploy/monitoring/k8s/namespace.yaml
```

Create the Grafana administrator credential without writing the password to
Git. Choose a unique password when prompted:

```bash
read -rsp "Grafana admin password: " GRAFANA_PASSWORD
printf '\n'
kubectl create secret generic falzo-grafana-admin \
  --namespace monitoring \
  --from-literal=admin-user=admin \
  --from-literal=admin-password="$GRAFANA_PASSWORD" \
  --dry-run=client \
  -o yaml | kubectl apply -f -
unset GRAFANA_PASSWORD
```

Install the pinned monitoring chart. Review and deliberately update the
version instead of silently following `latest` during later upgrades:

```bash
helm repo add prometheus-community \
  https://prometheus-community.github.io/helm-charts
helm repo update

helm template falzo-monitoring \
  prometheus-community/kube-prometheus-stack \
  --version 87.21.0 \
  --namespace monitoring \
  --values deploy/monitoring/kube-prometheus-stack-values.yaml \
  >/dev/null

helm upgrade --install falzo-monitoring \
  prometheus-community/kube-prometheus-stack \
  --version 87.21.0 \
  --namespace monitoring \
  --values deploy/monitoring/kube-prometheus-stack-values.yaml \
  --wait \
  --timeout 10m
```

Apply the Falzo scrape target, rules, dashboard, certificate and Ingress. The
Prometheus Operator CRDs must exist first, which is why this happens after the
Helm installation:

```bash
kubectl apply --dry-run=server -f deploy/monitoring/k8s
kubectl apply -f deploy/monitoring/k8s
```

Wait for the public endpoint:

```bash
kubectl rollout status deployment/falzo-grafana \
  --namespace monitoring \
  --timeout=180s
kubectl wait --for=condition=Ready certificate/falzo-monitor-tls \
  --namespace monitoring \
  --timeout=300s
kubectl get pods,svc,ingress,pvc -n monitoring
kubectl get servicemonitor,prometheusrule -n falzo
```

Open `https://monitor.falzo.life`, sign in as `admin`, then open
**Dashboards → Falzo Overview**. In **Explore**, this query should return `1`:

```promql
max(up{namespace="falzo", service="falzo-backend"})
```

If it returns no data, inspect the selected target without publishing
Prometheus on the Internet:

```bash
kubectl port-forward \
  --namespace monitoring \
  service/falzo-monitoring-kube-prom-prometheus \
  9090:9090
```

Then open `http://127.0.0.1:9090/targets` on the machine running the command.

## Upgrade or remove

Upgrade by changing the pinned chart version only after reviewing its release
notes, then rerun the `helm upgrade --install` and `kubectl apply` commands.

Removing the Helm release does not automatically remove all PVCs. This is
intentional so an accidental uninstall does not silently destroy monitoring
history:

```bash
helm uninstall falzo-monitoring --namespace monitoring
```

Do not delete the `monitoring` namespace or its PVCs unless loss of Prometheus
history and Grafana state is explicitly intended.

## Alert rule source

`falzo-alerts.yml` remains the plain Prometheus rule source. The equivalent
Prometheus Operator resource is `k8s/prometheus-rules.yaml`. Keep the rule
expressions synchronized when changing thresholds.
