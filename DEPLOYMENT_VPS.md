# Deploy the server to a VPS

This reuses the same `compose.yaml` (game server + Caddy for HTTPS) already in
the repo, but on a rented VPS instead of your laptop. That avoids CGNAT/port-forwarding
issues and gives you a stable public IP. All world state is still in memory:
restarting the server resets the game and disconnects players.

## 1. Provision a VPS

Any small provider works (DigitalOcean, Hetzner, Linode, Vultr, ...). Minimum
useful size: 1 vCPU, 1 GB RAM, Ubuntu 22.04 or 24.04 LTS. Note the VPS's public
IPv4 address once it's created.

## 2. Point DNS at it

Create a DNS **A record** for your domain (e.g. `play.yourdomain.com`) pointing
to the VPS's public IPv4. If you don't own a domain, a free DDNS hostname (e.g.
[DuckDNS](https://www.duckdns.org)) pointed at the VPS's IP also works. Don't add
an AAAA record unless the VPS also has IPv6 configured. Wait for DNS to resolve
before continuing (`dig +short play.yourdomain.com` should return the VPS IP).

## 3. Basic server hardening

SSH in as root (or the provider's default user), then:

```bash
adduser deploy
usermod -aG sudo deploy
rsync --archive --chown=deploy:deploy ~/.ssh /home/deploy
```

Log back in as `deploy` from here on. Then lock down SSH and enable a firewall:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 443/udp   # Caddy negotiates HTTP/3 over UDP too
sudo ufw enable
```

**Important Docker + ufw gotcha:** Docker inserts its own `iptables` rules for
published container ports, and those rules take effect *ahead of* ufw's rules.
In practice this means a container publishing a port is reachable from the
internet even if you never ran `ufw allow` for it. Treat ufw here as
defense-in-depth, not the only gate — your cloud provider's own network/security-group
firewall (in its dashboard) is what actually restricts inbound traffic, so
configure it there too: allow TCP 22 (or your SSH port), 80, 443, and UDP 443;
deny everything else inbound.

## 4. Install Docker

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker deploy
```

Log out and back in so the group change takes effect, then confirm with
`docker run hello-world`.

## 5. Get the code onto the server

Easiest if the repo is pushed to GitHub/GitLab:

```bash
git clone <your-repo-url> raylib-2d-sandbox
cd raylib-2d-sandbox
```

No local repo remote yet? `rsync -avz --exclude .git --exclude dist ./ deploy@VPS_IP:raylib-2d-sandbox/`
from your machine works just as well — the server only needs `compose.yaml`,
`Dockerfile.server`, `deploy/Caddyfile`, `go.mod`, `go.sum`, `cmd/server`,
and the `internal` packages the server imports (`server`, `protocol`,
`simulation`, `worldmap`, `physics`, `parameters`) to build.

## 6. Configure and start

```bash
cp .env.example .env
```

Edit `.env` and set `GAME_DOMAIN` to the hostname from step 2 (no `https://`,
port, or path). Then:

```bash
docker compose up -d --build
docker compose ps
docker compose logs --tail=100 caddy server
```

Caddy requests and renews its TLS certificate automatically once DNS resolves
and 80/443 are reachable. Certificates persist in a Docker volume across
container restarts.

## 7. Verify

```bash
curl https://YOUR_HOSTNAME/health
```

Expect `{"status":"ok"}`. Then connect a client with:

```bash
./raylib-2d-sandbox -server wss://YOUR_HOSTNAME/ws
```

(build the client per the repo's build instructions — same source, `go build -o raylib-2d-sandbox ./cmd`)
and choose **Multiplayer**.

## Update, stop, and troubleshoot

```bash
# pull latest code and redeploy (resets world state)
git pull
docker compose up -d --build

docker compose logs --tail=100 caddy server
docker compose down          # stop; keeps certs
docker compose down -v       # stop AND wipes cert volumes — avoid unless intentional
```

- **Certificate errors:** check DNS resolves to the VPS IP, and that 80/443 are
  actually reachable (test with `curl -I http://YOUR_HOSTNAME` from outside the VPS).
- **Connection refused from outside but works via `curl` on the VPS itself:**
  almost always the cloud provider's firewall/security group, not ufw or Caddy.
- **502 response:** `docker compose ps` and `docker compose logs server`.
- **Health works but client fails:** double check you used `wss://HOST/ws` and
  selected Multiplayer, and that client/server are built from the same revision.

There's currently no password or invite restriction on `/ws` — anyone who can
reach the hostname can join. Fine for a playtest; add auth before treating this
as a public launch.

Reference: [Caddy reverse proxy and HTTPS setup](https://caddyserver.com/docs/quick-starts/reverse-proxy).
