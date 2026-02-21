---
description: Update changelog, commit all changes, and push to GitHub
---

# Commit & Push to GitHub

## Prerequisites
- Working directory: `/opt/proxy-server-local`
- Git remote: `github.com:Chinsusu/proxy-manager.git` (main branch)
- SSH agent must have key loaded (uses SSH, not HTTPS)

---

## Steps

### 1. Check current status
// turbo
```bash
cd /opt/proxy-server-local && git status --short && git log --oneline -3
```
Review what files have changed before writing changelog.

---

### 2. Determine next version
Look at the latest version in `CHANGELOG.md` and increment:
- **patch** (1.4.x): bug fixes only
- **minor** (1.x.0): new features
- **major** (x.0.0): breaking changes

// turbo
```bash
head -5 /opt/proxy-server-local/CHANGELOG.md
```

---

### 3. Update CHANGELOG.md
Add new entry at the top of `CHANGELOG.md`, after the `# Changelog` heading, using this format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- ...

### Changed
- ...

### Fixed
- ...
```

Use today's date. Only include sections that have entries.

---

### 4. Stage all changes
// turbo
```bash
cd /opt/proxy-server-local && git add -A && git status --short
```

Review the staged files — make sure nothing unexpected is included.

---

### 5. Commit
**IMPORTANT:** Use `timeout 30` to prevent git commit from hanging. The `-m` flag + `GIT_EDITOR=true` should be enough, but background execution can sometimes cause a hang. With `timeout 30`, the command exits after 30 seconds if it doesn't complete — but **the commit often succeeds silently** even if the tool reports "running". Always check `git log --oneline -2` after.

// turbo
```bash
cd /opt/proxy-server-local && timeout 30 git commit -m "feat/fix/chore: short description (vX.Y.Z)

- Bullet point detail 1
- Bullet point detail 2" 2>&1 || true
```

After running, immediately verify:
// turbo
```bash
cd /opt/proxy-server-local && git log --oneline -2
```

If HEAD shows the new commit message → commit succeeded. The `|| true` prevents tool failure even if timeout fires.

**If commit appears stuck (> 10s):**
// turbo
```bash
pkill -f "git commit" 2>/dev/null; cd /opt/proxy-server-local && git log --oneline -2
```
If the commit shows in the log → it already committed, just push.

**If index.lock exists (from crashed commit):**
// turbo
```bash
rm -f /opt/proxy-server-local/.git/index.lock && cd /opt/proxy-server-local && timeout 30 git commit -m "your message" 2>&1 || true
```

---

### 6. Push to GitHub
// turbo
```bash
cd /opt/proxy-server-local && \
  GIT_SSH_COMMAND="ssh -o ConnectTimeout=10 -o ServerAliveInterval=10 -o ServerAliveCountMax=3" \
  timeout 60 git push origin main 2>&1 && \
  echo "PUSHED OK" || echo "PUSH FAILED"
```

Expected output:
```
To github.com:Chinsusu/proxy-manager.git
   <old_hash>..<new_hash>  main -> main
PUSHED OK
```

After push, verify commit is on remote:
// turbo
```bash
cd /opt/proxy-server-local && git log --oneline origin/main -3
```



## Troubleshooting

### `git commit` hangs / no output for > 30s
This happens when git tries to open an editor (e.g., missing `$GIT_EDITOR`). Fix:
```bash
# Terminate the hanging command, then run:
cd /opt/proxy-server-local && GIT_EDITOR=true git commit -m "your message"
```
Or check if commit already succeeded silently:
```bash
git status --short  # empty = already committed
```

### `git push` fails with auth error
SSH key may not be loaded:
```bash
ssh -T git@github.com  # should say: Hi Chinsusu! You've successfully authenticated
eval $(ssh-agent -s) && ssh-add ~/.ssh/id_ed25519
```

### `Text file busy` when copying binary
Stop the service first:
```bash
systemctl stop pgw-ui && cp bin/pgw-ui /usr/local/bin/pgw-ui && systemctl start pgw-ui
```

### Git index.lock exists
```bash
rm /opt/proxy-server-local/.git/index.lock
```
