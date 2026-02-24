---
name: restart
description: Rebuild miniclaw and restart the systemd service
disable-model-invocation: true
allowed-tools: "Bash(source *), Bash(go install *), Bash(date *), Bash(cat *), Bash(systemctl --user restart miniclaw)"
---

# Restart miniclaw

This skill rebuilds the miniclaw binary and restarts the systemd service. The restart will kill the current process (including this claude subprocess), so a scheduled task is used to confirm success afterwards.

Follow these steps **in order**. Do NOT skip steps or reorder them.

## Step 1: Load environment

Run `source ~/.miniclaw/.env` to ensure `MINICLAW_AGENT_DIR` and other variables are available.

## Step 2: Build the binary

Run `go install $MINICLAW_AGENT_DIR/../cmd/miniclaw/` to compile and install the updated binary. Report success or failure. Do NOT continue if the build fails.

## Step 3: Schedule a post-restart confirmation

Create a one-time scheduled task for **each** chat ID in `$ALLOWED_CHAT_IDS` (comma-separated). Each task fires 10 seconds from now and sends a confirmation message.

1. Compute the timestamp: `date -u -d '+10 seconds' --iso-8601=seconds`
2. Loop over each chat ID and write a task file per chat:

```bash
TIMESTAMP=$(date -u -d '+10 seconds' --iso-8601=seconds)
IFS=',' read -ra CHAT_IDS <<< "$ALLOWED_CHAT_IDS"
for CHAT_ID in "${CHAT_IDS[@]}"; do
  CHAT_ID=$(echo "$CHAT_ID" | tr -d ' ')
  cat > ~/.miniclaw/data/tasks/restart-confirmation-${CHAT_ID}.json << EOF
{
    "prompt": "Send a short message confirming that miniclaw has been rebuilt and restarted successfully. Keep it to one or two lines.",
    "chat_id": ${CHAT_ID},
    "type": "once",
    "value": "",
    "status": "active",
    "next_run": "${TIMESTAMP}"
}
EOF
done
```

It is fine to overwrite these files if they already exist from a previous restart.

## Step 4: Restart the service

Run `systemctl --user restart miniclaw`.

**This command kills the current miniclaw process and this claude subprocess.** This is expected behaviour. Nothing after this command will execute.
