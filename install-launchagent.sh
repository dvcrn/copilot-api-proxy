#!/bin/bash

set -e

# Get the absolute path of the current directory
PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLIST_NAME="com.copilot-oauth-proxy.plist"
PLIST_LOCAL="${PROJECT_DIR}/${PLIST_NAME}"
LAUNCHAGENTS_DIR="${HOME}/Library/LaunchAgents"
PLIST_SYMLINK="${LAUNCHAGENTS_DIR}/${PLIST_NAME}"

echo "Installing copilot-oauth-proxy LaunchAgent..."
echo "Project directory: ${PROJECT_DIR}"

# Create LaunchAgents directory if it doesn't exist
mkdir -p "${LAUNCHAGENTS_DIR}"

# Generate the plist file locally in project directory
cat > "${PLIST_LOCAL}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.copilot-oauth-proxy</string>

    <key>Program</key>
    <string>${PROJECT_DIR}/run_proxy.sh</string>

    <key>WorkingDirectory</key>
    <string>${PROJECT_DIR}</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
        <key>Crashed</key>
        <true/>
    </dict>

    <key>ThrottleInterval</key>
    <integer>30</integer>

    <key>StandardOutPath</key>
    <string>${HOME}/Library/Logs/copilot-oauth-proxy.log</string>

    <key>StandardErrorPath</key>
    <string>${HOME}/Library/Logs/copilot-oauth-proxy.error.log</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>${HOME}</string>
        <key>PATH</key>
        <string>${PATH}</string>
    </dict>
</dict>
</plist>
EOF

echo "Plist file created at: ${PLIST_LOCAL}"

# Validate plist was created
if [ ! -f "${PLIST_LOCAL}" ]; then
    echo "❌ Error: Failed to create plist file"
    exit 1
fi

# Remove old symlink if it exists
if [ -L "${PLIST_SYMLINK}" ]; then
    echo "Removing old symlink..."
    rm "${PLIST_SYMLINK}"
fi

# Unload the service if it's already running
if launchctl list | grep -q "com.copilot-oauth-proxy"; then
    echo "Unloading existing service..."
    launchctl unload "${PLIST_SYMLINK}" 2>/dev/null || true
fi

# Create symlink from LaunchAgents to local plist
echo "Creating symlink: ${PLIST_SYMLINK} -> ${PLIST_LOCAL}"
ln -sf "${PLIST_LOCAL}" "${PLIST_SYMLINK}"

# Validate symlink was created
if [ ! -L "${PLIST_SYMLINK}" ]; then
    echo "❌ Error: Failed to create symlink"
    exit 1
fi

# Load the new service
echo "Loading service..."
launchctl load "${PLIST_SYMLINK}"

# Check if the service is running
sleep 2
if launchctl list | grep -q "com.copilot-oauth-proxy"; then
    echo "✅ LaunchAgent installed and started successfully!"
    echo ""
    echo "Service management commands:"
    echo "  Check status:  launchctl list | grep copilot-oauth-proxy"
    echo "  View logs:     tail -f ~/Library/Logs/copilot-oauth-proxy.log"
    echo "  View errors:   tail -f ~/Library/Logs/copilot-oauth-proxy.error.log"
    echo "  Stop service:  launchctl unload ~/Library/LaunchAgents/${PLIST_NAME}"
    echo "  Start service: launchctl load ~/Library/LaunchAgents/${PLIST_NAME}"
    echo "  Uninstall:     ./uninstall-launchagent.sh"
else
    echo "⚠️  Service may not have started correctly. Check logs at:"
    echo "  ~/Library/Logs/copilot-oauth-proxy.error.log"
fi