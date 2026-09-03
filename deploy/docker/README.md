# Falzo on Docker Compose

> Legacy manual rollback option. Production uses K3s through
> `.github/workflows/deploy-k3s.yml`. Running the Docker workflow stops K3s and
> transfers ports 80 and 443 to Caddy.

The legacy stack runs as one Docker Compose project on the VPS:

1. PostgreSQL and Redis start and pass health checks.
2. The one-shot migration container completes successfully.
3. Backend, frontend and Caddy start without waiting for notification services.
4. NATS JetStream and the Telegram alert bot start as optional workers; their
   failure is reported as a warning and does not fail the main deployment.
5. Caddy terminates TLS for `falzo.life`, redirects `www.falzo.life` to the
   apex domain, sends `/api` to the Go backend, and sends all other requests to
   Next.js. WebSocket upgrades work through the same `/api` route.

PostgreSQL and Redis do not publish host ports. Only Caddy publishes ports 80
and 443.

## VPS prerequisites

- Linux VPS with Docker Engine and the Docker Compose v2 plugin.
- The deployment SSH user can run Docker, either directly or with passwordless
  `sudo`.
- DNS `A`/`AAAA` records for `falzo.life` point to the VPS, with
  `www.falzo.life` configured as a `CNAME` to `falzo.life`.
- TCP ports 22, 80 and 443 and UDP port 443 are allowed by the VPS firewall.

Verify Docker before the first deployment:

```bash
docker version
docker compose version
docker compose up --help | grep -- --wait
```

The workflow stops and disables `k3s.service` when it is active so that K3s
Traefik does not compete with Caddy for ports 80 and 443. It does not uninstall
K3s or delete `/var/lib/rancher/k3s`, so the old cluster data remains recoverable.

Docker uses new named volumes and does not automatically import PostgreSQL or
Redis data from K3s volumes. If the old cluster contains production data, export
and verify that data before the first Docker deployment.

## GitHub configuration

Create a GitHub Environment named `production` and add these secrets:

- `VPS_HOST`
- `VPS_USER`
- `VPS_PORT` (optional; defaults to `22`)
- `VPS_SSH_KEY`
- `VPS_APP_DIR`, for example `/opt/falzo`
- `GHCR_PULL_USERNAME`
- `GHCR_PULL_TOKEN` with `read:packages`
- `TELEGRAM_BOT_TOKEN`, created through BotFather (optional)
- `TELEGRAM_CHAT_ID`, the target user, group or channel ID (optional)

Set both Telegram values to enable the independent alert worker. If either is
missing, the main application still deploys and the Telegram worker is skipped.

Add the repository variable `NEXT_PUBLIC_GOOGLE_CLIENT_ID`. The OAuth client ID
is public configuration and is embedded in the frontend build.

The first deployment creates `$VPS_APP_DIR/shared/secrets.env` on the VPS with
mode `600`. It contains randomly generated PostgreSQL, Redis and JWT secrets and
is reused by later deployments. It is never uploaded to GitHub or committed.

To inspect the deployment on the VPS without printing secrets:

```bash
cd /opt/falzo/docker
sudo docker compose \
  --project-name falzo \
  --env-file /opt/falzo/shared/secrets.env \
  --env-file release.env \
  ps
```

Do not run `docker compose config` without `--quiet` in shared logs because the
rendered output contains interpolated secret values.
