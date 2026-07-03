package constants

const (
	ConfigDirName    = ".burrow"
	AuthFileName     = "auth.json"
	PIDsFileName     = "pids.json"
	BurrowConfigFile = ".burrow.yaml"
	ConfigDirEnvVar  = "BURROW_CONFIG_DIR"

	CloudflareAPIBase      = "https://api.cloudflare.com/client/v4"
	CloudflareVerifyPath   = "/user/tokens/verify"
	CloudflareTunnelDomain = "cfargotunnel.com"

	CloudflareAPITokenURL = "https://dash.cloudflare.com/profile/api-tokens"
	CloudflareAccountURL  = "https://dash.cloudflare.com"

	CloudflaredGitHubBase  = "https://github.com/cloudflare/cloudflared/releases/download"
	CloudflaredReleasesURL = "https://api.github.com/repos/cloudflare/cloudflared/releases/latest"
	CloudflaredBinDir      = "bin"
	LogsDir                = "logs"
)
