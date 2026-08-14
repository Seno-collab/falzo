# Falzo monitoring

The backend exposes Prometheus metrics at `GET /metrics`. The public Ingress
does not route this path to the backend; scrape the backend Service or Pods
from inside the cluster.

Load `falzo-alerts.yml` through Prometheus `rule_files`, or translate it into a
`PrometheusRule` when using Prometheus Operator. Validate changes with:

```bash
promtool check rules deploy/monitoring/falzo-alerts.yml
```

Alert thresholds are safe initial values and should be tuned after production
traffic establishes a baseline.
