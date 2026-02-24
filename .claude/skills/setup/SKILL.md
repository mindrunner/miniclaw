---
name: setup
description: Interactive setup wizard for new miniclaw users who just forked the repo
disable-model-invocation: true
allowed-tools: "Read, Edit, Bash(go *), Bash(which *), Bash(claude *), Bash(mkdir *), Bash(ls *), Bash(systemctl *), Bash(loginctl *)"
---

# miniclaw Setup Wizard

You are helping a new user set up miniclaw after forking the repo. Walk through each step below **in order**. Check prerequisites first, then guide the user through configuration.

## Step 1: Check prerequisites

Run these checks silently and report the results as a checklist:

1. **Go** — run `which go` and `go version`. Require Go 1.23+.
2. **Claude CLI** — run `which claude` and `claude --version`. This is required for the agent runtime. Assume the user is authenticated since they are already using Claude to set it up - do not run any Claude commands yourself as it will fail.

If any prerequisite is missing, tell the user what to fix and stop. Do not continue until all checks pass.

## Step 2: Install Go dependencies

Run `go mod tidy` from the repo root to fetch all dependencies. Report success or failure.

## Step 3: Install binary

Run `go install ./cmd/miniclaw/` to compile and install the `miniclaw` binary to the user's `$GOPATH/bin`. Report success or failure.

## Step 4: Create runtime directories

Create `~/.miniclaw/` and its subdirectories by running:

```
mkdir -p ~/.miniclaw/{data/tasks,workspace}
```

Report that the runtime directory structure has been created.

## Step 5: Personalize agent

Read `agent/preferences.md` and walk the user through customizing it:

1. **Name** — Ask what they want to name their bot (default: Enki)
2. **Timezone** — Ask for their timezone in UTC offset format (default: UTC+8)

Update `agent/preferences.md` with their choices using the Edit tool. If they're happy with a default, skip that field.

## Step 6: Telegram bot token

Ask the user for their Telegram bot token. Tell them:

- Create a bot via [@BotFather](https://t.me/BotFather) on Telegram
- Use the `/newbot` command and follow the prompts
- Copy the token BotFather gives you

Once they provide the token, hold onto it for Step 9.

## Step 7: Agent directory

Determine the absolute path to the `agent/` directory in the current repo by running `ls` on it. Hold onto this path for Step 9 as the `MINICLAW_AGENT_DIR` value. This tells the bot where to find its CLAUDE.md and preferences.md files, so it can be run from any directory.

## Step 8: Allowed chat IDs

Ask the user for their allowed Telegram chat IDs (comma-separated). Tell them:

- After setting up, they can send `/chatid` to the bot from any chat to get the ID
- For now, they can leave this empty and add it later
- Group chats have negative IDs (e.g. `-1001234567890`)
- Private chats have positive IDs

Hold onto the value for Step 9.

## Step 9: Write .env file

Write `~/.miniclaw/.env` with the collected values:

```
TELEGRAM_BOT_TOKEN=<their token>
ALLOWED_CHAT_IDS=<their chat IDs, or empty>
MINICLAW_AGENT_DIR=<absolute path to agent/ from Step 7>
```

Use the Bash tool to write this file with `0600` permissions. Do NOT use the Write tool (the path is outside the project).

## Step 10: Systemd service (optional)

Ask the user if they want to run miniclaw as a systemd user service so it starts automatically and runs in the background. If they decline, skip to Step 11.

If they accept:

1. Determine the absolute path to the `miniclaw` binary by running `which miniclaw` or falling back to `ls ~/go/bin/miniclaw`.
2. Create the systemd user service directory: `mkdir -p ~/.config/systemd/user`
3. Write `~/.config/systemd/user/miniclaw.service` via the Bash tool with the following content:

```ini
[Unit]
Description=miniclaw Telegram Bot
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=%h/.miniclaw/.env
ExecStart=<absolute path to miniclaw binary>
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

4. Reload the systemd user daemon: `systemctl --user daemon-reload`
5. Enable the service so it starts on login: `systemctl --user enable miniclaw`
6. Enable lingering so it runs even when the user is not logged in: `loginctl enable-linger`
7. Ask if they want to start it now. If yes: `systemctl --user start miniclaw` and confirm it's running with `systemctl --user status miniclaw`.

## Step 11: Done

Print a summary:

```
Setup complete!
```

If they set up systemd, add:

```
miniclaw is running as a systemd user service.

  systemctl --user status miniclaw   — check status
  systemctl --user restart miniclaw  — restart after config changes
  journalctl --user -u miniclaw -f   — follow logs
```

If they skipped systemd, add:

```
To run miniclaw:

  miniclaw
```

In both cases, add:

```
To find your chat ID, send /chatid to your bot on Telegram,
then add it to ~/.miniclaw/.env as ALLOWED_CHAT_IDS.
```

If they left ALLOWED_CHAT_IDS empty, remind them to:

1. Start the bot without an allowlist (it will respond to anyone)
2. Send `/chatid` to the bot
3. Add the ID to `~/.miniclaw/.env`
4. Restart the bot (or `systemctl --user restart miniclaw` if using systemd)

## Rules

- Be concise and friendly
- Do NOT proceed past a failed step — fix it first
- Do NOT print raw commands unless the user asks to see them
- Do NOT modify any repo files except `agent/preferences.md` — only create `~/.miniclaw/.env` and edit preferences
