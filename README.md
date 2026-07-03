# burrow

Persistent, named tunnel URLs for development teams — powered by Cloudflare.

```
burrow up
  [up] frontend  -> https://frontend.myapp.com
  [up] api       -> https://api.myapp.com

All tunnels running. Press Ctrl+C to stop.
```

## Why burrow?

Most tunnel tools give you a random URL that changes every restart. This breaks webhooks, breaks shared links, and means your team is constantly asking "what's your ngrok URL?"

Burrow gives you **named, persistent URLs** defined in a `.burrow.yaml` file committed to your repo. Every teammate runs `burrow up` and gets the exact same domains every time.

## Requirements

- A [Cloudflare account](https://dash.cloudflare.com/sign-up) (free)
- A domain managed by Cloudflare (~$10/year) — for named persistent URLs
- No domain? Use `burrow share` for a free temporary URL with no account needed

## Installation

### macOS (Homebrew)
```bash
brew install frank-chris/tap/burrow
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
scoop bucket add burrow https://github.com/frank-chris/scoop-burrow
scoop install burrow
```

### Windows (manual)
Download `burrow_windows_amd64.zip` from the [latest release](https://github.com/frank-chris/burrow/releases/latest), extract, and add to your PATH.

### Build from source
```bash
go install github.com/frank-chris/burrow@latest
```

## Quick start

**1. Set up credentials**
```bash
burrow init
```
This walks you through entering your Cloudflare API token and Account ID, validates them, and downloads the cloudflared binary automatically. You only need to do this once.

**2. Quick share (no config needed)**
```bash
burrow share 3000
# → https://random-words.trycloudflare.com
```

**3. Named persistent tunnels**

Create a `.burrow.yaml` in your project:
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

Then:
```bash
burrow up
```

Commit `.burrow.yaml` to your repo. Every teammate gets the same URLs.

## Commands

| Command | Description |
|---------|-------------|
| `burrow init` | First-time setup — credentials and cloudflared install |
| `burrow up` | Start all tunnels defined in `.burrow.yaml` |
| `burrow up <name>` | Start a single named tunnel |
| `burrow down` | Stop all running tunnels |
| `burrow share <port>` | Quick one-off public URL (no config needed) |
| `burrow share <port> --password <pw>` | Password-protected share |
| `burrow share <port> --ttl 2h` | Share that expires after 2 hours |
| `burrow status` | Show status of running tunnels |
| `burrow logs` | Show tunnel request logs |
| `burrow logs <name>` | Show logs for a specific tunnel |
| `burrow logs -f` | Follow logs in real time |
| `burrow doctor` | Diagnose setup issues |
| `burrow uninstall` | Remove all burrow data from your system |

## Configuration

`.burrow.yaml` supports multiple tunnels, each with a name, port, and domain:

```yaml
provider: cloudflare

tunnels:
  - name: frontend
    port: 3000
    domain: frontend.myapp.com

  - name: api
    port: 8080
    domain: api.myapp.com

  - name: docs
    port: 4000
    domain: docs.myapp.com
```

The file is designed to be committed to your repo. Teammates clone and run `burrow up` — no setup beyond `burrow init` with their own credentials.

## Environment variables

| Variable | Description |
|----------|-------------|
| `BURROW_CONFIG_DIR` | Override the default config directory (`~/.burrow`) |

## Cloudflare setup

1. [Create a Cloudflare account](https://dash.cloudflare.com/sign-up)
2. Add your domain to Cloudflare and update your nameservers
3. Generate an API token at [dash.cloudflare.com/profile/api-tokens](https://dash.cloudflare.com/profile/api-tokens)
   - Use the **Edit zone DNS** template
   - Scope it to your domain
4. Find your Account ID on the right sidebar at [dash.cloudflare.com](https://dash.cloudflare.com)
5. Run `burrow init` and enter both values

## License

MIT
