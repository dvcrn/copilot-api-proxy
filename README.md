# Copilot API Proxy

A FAST reverse proxy server written in Go that forwards `/v1/chat/completions` requests to the GitHub Copilot API, to expose the Copilot API to other tools

> [!WARNING]
> This is a reverse-engineered proxy of GitHub Copilot API. It is not supported by GitHub, and may break unexpectedly. Use at your own risk.

> [!WARNING]
> **GitHub Security Notice:**
> Excessive automated or scripted use of Copilot (including rapid or bulk requests, such as via automated tools) may trigger GitHub's abuse-detection systems.
> You may receive a warning from GitHub Security, and further anomalous activity could result in temporary suspension of your Copilot access.
>
> GitHub prohibits use of their servers for excessive automated bulk activity or any activity that places undue burden on their infrastructure.
>
> Please review:
>
> - [GitHub Acceptable Use Policies](https://docs.github.com/site-policy/acceptable-use-policies/github-acceptable-use-policies#4-spam-and-inauthentic-activity-on-github)
> - [GitHub Copilot Terms](https://docs.github.com/site-policy/github-terms/github-terms-for-additional-products-and-features#github-copilot)
>
> Use this proxy responsibly to avoid account restrictions.

## Setup

```
go install github.com/dvcrn/copilot-api-proxy/cmd/copilot-api-proxy@latest
```

### Authenticate

Run

```
copilot-api-proxy auth
```

which will start the Copilot auth flow

Then run

```
copilot-api-proxy server
```

to run the server

**Optional Config**

- Set the `COPILOT_TOKEN` environment variable with your GitHub Copilot authentication token.
- Optionally set `PORT` (defaults to 9871).

### Auto-start on Boot (macOS)

To automatically start the proxy when your system boots:

```bash
# Install the launch agent
./install-launchagent.sh

# Check status
launchctl list | grep copilot-api-proxy

# View logs
tail -f ~/Library/Logs/copilot-api-proxy.log

# Stop the service
launchctl unload ~/Library/LaunchAgents/com.copilot-api-proxy.plist

# Uninstall
./uninstall-launchagent.sh
```
