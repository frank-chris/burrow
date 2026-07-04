# burrow

Persistent, named tunnel URLs for development teams - powered by Cloudflare.

```
burrow up
  [up] frontend  -> https://frontend.myapp.com
  [up] api       -> https://api.myapp.com

All tunnels running. Press Ctrl+C to stop.
```

## Why burrow?

Most tunnel tools give you a random URL that changes every restart. This breaks webhooks, breaks shared links, and means your team is constantly asking "what's your ngrok URL?"

Burrow gives you **named, persistent URLs** defined in a `.burrow.yaml` file committed to your repo. Every teammate runs `burrow up` and gets the exact same domains every time.

No domain? Burrow works without a Cloudflare account too - tunnels without a domain fall back to free temporary trycloudflare.com URLs automatically.

## Requirements

For quick tunnels (`burrow share`, or `burrow up` without domains):
- Nothing. No account needed.

For persistent named URLs (`burrow up` with domains):
- A [Cloudflare account](https://dash.cloudflare.com/sign-up) (free)
- A domain managed by Cloudflare (~$10/year)

## Installation

### macOS (Homebrew)
```bash
brew tap frank-chris/tap
brew install burrow
```

### Linux
```bash
curl -sSL https://github.com/frank-chris/burrow/releases/latest/download/burrow_linux_amd64.tar.gz | tar -xz
sudo mv burrow /usr/local/bin/
```

For ARM64:
```bash
curl -sSL https://github.com/frank-chris/burrow/releases/latest/download/burrow_linux_arm64.tar.gz | tar -xz
sudo mv burrow /usr/local/bin/
```

### Windows (Scoop)
```powershell
scoop bucket add frank-chris https://github.com/frank-chris/scoop-burrow
scoop install burrow
```

### Windows (manual)
Download `burrow_windows_amd64.zip` from the [latest release](https://github.com/frank-chris/burrow/releases/latest), extract, and add to your PATH.

### Build from source
```bash
go install github.com/frank-chris/burrow@latest
```

## Quick start

### No account needed

Share a local port instantly:
```bash
burrow share 3000
# → https://random-words.trycloudflare.com
```

Or run multiple tunnels without any config:
```yaml
# .burrow.yaml
tunnels:
  - name: frontend
    port: 3000
  - name: api
    port: 8080
```
```bash
burrow up
# → [up] frontend -> https://random1.trycloudflare.com
# → [up] api      -> https://random2.trycloudflare.com
```

cloudflared is downloaded automatically on first use.

### Persistent named URLs (requires Cloudflare account)

Run `burrow init` once to store your credentials:
```bash
burrow init
```

Then add domains to your `.burrow.yaml`:
```yaml
provider: cloudflare

tunnels:
  - name: frontend
    port: 3000
    domain: frontend.myapp.com
  - name: api
    port: 8080
    domain: api.myapp.com
```

```bash
burrow up
# → [up] frontend -> https://frontend.myapp.com
# → [up] api      -> https://api.myapp.com
```

Commit `.burrow.yaml` to your repo. Every teammate gets the same URLs every time.

## Commands

| Command | Description |
|---------|-------------|
| `burrow init` | Set up Cloudflare credentials (only needed for persistent URLs) |
| `burrow up` | Start all tunnels defined in `.burrow.yaml` |
| `burrow up <name>` | Start a single named tunnel |
| `burrow down` | Stop all running tunnels |
| `burrow down <name>` | Stop a single named tunnel |
| `burrow share <port>` | Quick one-off public URL (no account needed) |
| `burrow share <port> --password <pw>` | Password-protected share |
| `burrow share <port> --ttl 2h` | Share that expires after 2 hours |
| `burrow share <port> --qr` | Print a QR code for the tunnel URL |
| `burrow status` | Show status of running tunnels |
| `burrow logs` | Show tunnel request logs |
| `burrow logs <name>` | Show logs for a specific tunnel |
| `burrow logs -f` | Follow logs in real time |
| `burrow doctor` | Diagnose setup issues |
| `burrow uninstall` | Remove all burrow data from your system |

`burrow down`, `burrow status`, and `burrow logs` only apply to tunnels started with `burrow up`. `burrow share` runs in the foreground - use Ctrl+C to stop it. `burrow logs` requires a domain in `.burrow.yaml`; quick tunnels (no domain) do not write logs.

## Configuration

`.burrow.yaml` supports multiple tunnels. `domain` is optional - omit it to use a free temporary trycloudflare.com URL:

```yaml
provider: cloudflare  # only needed if any tunnel has a domain

tunnels:
  - name: frontend
    port: 3000
    domain: frontend.myapp.com  # persistent URL - requires burrow init

  - name: api
    port: 8080
    domain: api.myapp.com

  - name: docs
    port: 4000
    # no domain - gets a free temporary URL
```

## WebSocket and HMR

Burrow inherits WebSocket support from cloudflared. Frameworks that use WebSockets for hot module replacement - Vite, Next.js, webpack-dev-server - work without any extra configuration.

## Environment variables

| Variable | Description |
|----------|-------------|
| `BURROW_CONFIG_DIR` | Override the default config directory (`~/.burrow`) |

## Cloudflare setup

Only needed for persistent named URLs:

1. [Create a Cloudflare account](https://dash.cloudflare.com/sign-up)
2. Add your domain to Cloudflare and update your nameservers
3. Generate an API token at [dash.cloudflare.com/profile/api-tokens](https://dash.cloudflare.com/profile/api-tokens)
   - Use the **Edit zone DNS** template
   - Scope it to your domain
4. Find your Account ID on the right sidebar at [dash.cloudflare.com](https://dash.cloudflare.com)
5. Run `burrow init` and enter both values

## License

MIT
