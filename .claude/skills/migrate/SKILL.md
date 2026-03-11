---
name: migrate
description: Migrate your main session context into the current thread
---

# Migrate Session to Thread

This skill copies the main (non-threaded) Claude session into the current thread so you retain full conversation context.

## Step 1: Check thread context

Read the `MINICLAW_THREAD_ID` environment variable.

If it is `0` or empty, tell the user:

> This command only works inside a thread. Open a thread first, then send /migrate there.

Stop here.

## Step 2: Load sessions

Read the file `~/.miniclaw/data/sessions.json`. This is a JSON object mapping session keys to Claude CLI session UUIDs.

Key format:
- `"<chatID>"` — the main (non-threaded) session
- `"<chatID>:<threadID>"` — a thread-specific session

Look up the main session using the plain `MINICLAW_CHAT_ID` value (no colon, no thread suffix).

If no main session exists, tell the user:

> No main session found to migrate. There's nothing to copy.

Stop here.

## Step 3: Check for existing thread session

Build the thread key: `"<MINICLAW_CHAT_ID>:<MINICLAW_THREAD_ID>"`.

If a session already exists for this thread key, tell the user:

> This thread already has a session. Migrating will replace it with the main session's context. Do you want to proceed?

Wait for confirmation before continuing. If the user declines, stop here.

## Step 4: Confirm migration

Tell the user what will happen:

> I'll copy the main session into this thread. The main session will remain intact so it can still be used in non-threaded mode.
>
> Proceed?

Wait for the user to confirm.

## Step 5: Execute migration

1. Read `~/.miniclaw/data/sessions.json` again (fresh read in case it changed)
2. Copy the main session UUID to the thread key: set `sessions["<chatID>:<threadID>"]` = `sessions["<chatID>"]`
3. Do NOT delete the main session key — keep it so non-threaded mode still works
4. Write the updated JSON back to `~/.miniclaw/data/sessions.json`

## Step 6: Confirm

Tell the user:

> Done. This thread now has the same context as your main session. Your next message here will resume with full history.
