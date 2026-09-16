# Play together over ZeroTier

For a few friends, ZeroTier is simpler than the VPS setup in
[DEPLOYMENT_VPS.md](DEPLOYMENT_VPS.md): everyone gets a private IP on a
virtual LAN, no public hostname, no TLS cert, no port-forwarding/CGNAT
issues. The host just runs the plain `-addr` server and shares their
ZeroTier IP.

## 1. Install ZeroTier (each person, Ubuntu)

```bash
curl -s https://install.zerotier.com | sudo bash
```

This installs a background service (`zerotier-one`) and a CLI
(`zerotier-cli`), not a GUI app. `zerotier-cli` lives in `/usr/sbin`,
which usually isn't on your `PATH`:

```bash
echo 'export PATH="$PATH:/usr/sbin"' >> ~/.bashrc
source ~/.bashrc
```

All `zerotier-cli` commands below need `sudo`.

## 2. Create or join a network

**Whoever is hosting the network** (only needs to be done once, by one
person) creates it at https://my.zerotier.com (free account, web GUI —
ZeroTier has no desktop GUI on Linux). Note the 16-char **Network ID**.

**Everyone** (host included) joins it:

```bash
sudo zerotier-cli join <NETWORK_ID>
```

If the network is private (default), the network owner must approve
each new member in the member list at https://my.zerotier.com before
they get an IP.

Confirm you're connected and get your assigned IP:

```bash
sudo zerotier-cli listnetworks
```

Status must say `OK`, not `ACCESS_DENIED`. Note the IP shown (e.g.
`10.x.x.x` or `172.2x.x.x`) — that's your address on the virtual LAN.

## 3. Host: start the server

```bash
go run ./cmd/server -addr 0.0.0.0:8080
```

Binding `0.0.0.0` is fine here since the game has no public exposure —
only people on your ZeroTier network can reach it. (If your machine also
has a public IP you don't trust, bind to your ZeroTier IP specifically
instead, e.g. `-addr 10.50.239.15:8080`.)

Sanity-check it's up:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## 4. Everyone else: connect

```bash
go run ./cmd -server ws://HOST_ZEROTIER_IP:8080/ws
```

(or the built binary: `./raylib-2d-sandbox -server ws://HOST_ZEROTIER_IP:8080/ws`)

Then pick **Multiplayer** in the menu.

## Troubleshooting

- **`curl http://HOST_ZEROTIER_IP:8080/health` works, game doesn't
  connect / hangs:** the game logs the real error to the terminal
  (`log.Printf("multiplayer: %v", ...)`) even though the in-game message
  is generic. Check there first.
- **`context deadline exceeded` in that log line:** the ZeroTier tunnel
  itself isn't passing traffic, not a server/firewall problem. Check:
  ```bash
  sudo zerotier-cli listpeers
  ```
  Look at the `PATH` column for the other person's node — should say
  `DIRECT`. `RELAY` or no path at all means NAT traversal failed (usually
  a restrictive router on one side). Try `sudo systemctl restart
  zerotier-one` on both machines first; persistent relay/no-path issues
  are ZeroTier connectivity, not something to fix in this repo.
- **`curl` itself fails / connection refused:** check `sudo ufw status`
  on the host — if active, allow the game port on the ZeroTier interface
  specifically (find the interface name with `ip addr show | grep zt`):
  ```bash
  sudo ufw allow in on <zt-interface> to any port 8080
  ```
- **Wrong IP:** re-run `sudo zerotier-cli listnetworks` — the assigned
  IP can differ from what you last used if you left and rejoined.

There's no auth on `/ws` — anyone who can reach the host's ZeroTier IP
can join. Fine among friends on a private network; don't expose this
setup beyond that.
