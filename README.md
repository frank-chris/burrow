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

**vs ngrok:** ngrok charges $8-20/month for persistent URLs. Burrow runs on Cloudflare Tunnel (no usage fees), so the only cost is your domain (~$10/year), shared across the whole team with no per-seat pricing.

**vs raw cloudflared:** Cloudflare Tunnel is the right infrastructure, but the raw CLI takes 7 manual steps to set up and produces credentials files that cannot be safely committed to git. Burrow automates DNS record creation, wraps the config into a single committable `.burrow.yaml`, and handles multi-tunnel process management. Each user still needs to authenticate once with `burrow init`, but everything after that is automated.

## Requirements

For quick tunnels (`burrow share`, or `burrow up` without domains):
- Nothing. No account needed.

For persistent named URLs (`burrow up` with domains):
- A [Cloudflare account](https://dash.cloudflare.com/sign-up) (free)
- A domain managed by Cloudflare (~$10/year)

## Installation

### macOS and Linux
```bash
curl -sSL https://frank-chris.github.io/burrow/install.sh | sh
```

Also available via Homebrew: `brew tap frank-chris/tap && brew install burrow`

### Windows (Scoop)
```powershell
scoop bucket add frank-chris https://github.com/frank-chris/scoop-burrow
scoop install burrow
```

### GitHub Releases
Download a prebuilt binary for any platform from the [releases page](https://github.com/frank-chris/burrow/releases/latest).

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

Only needed for persistent named URLs. For a step-by-step walkthrough see the [Cloudflare setup guide](https://frank-chris.github.io/burrow/setup.html).

**Summary:**
1. [Create a Cloudflare account](https://dash.cloudflare.com/sign-up) and add your domain
2. Find your Account ID at [dash.cloudflare.com](https://dash.cloudflare.com) - click the &#8942; icon next to the + Add button and select "Copy account ID"
3. Create a **Custom Token** at [dash.cloudflare.com/profile/api-tokens](https://dash.cloudflare.com/profile/api-tokens) with these permissions:
   - Account / Cloudflare Tunnel / Edit
   - Zone / DNS / Edit
   - Zone / Zone / Read
4. Run `burrow init` and enter your token and Account ID

## License

MIT
