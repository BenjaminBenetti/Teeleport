# Teeleport Setup Guide

You are helping a user set up Teeleport — a tool that creates a persistent "global home folder" across devcontainer workspaces. Walk the user through each step interactively. Ask questions, wait for answers, and adapt based on their responses. Do not rush ahead.

---

## Step 1: Dotfile Repository

Teeleport runs via the devcontainer dotfiles system. The user needs a dotfile repository on GitHub.

**Ask the user:**
> Do you already have a dotfile repository on GitHub? (A repo that VS Code/Codespaces clones into your devcontainers for personalization.)

### If they have one:
- Ask for the repository URL or name (e.g., `username/dotfiles`)
- Ask where it's cloned locally so you can work with it
- Verify it has an install script (e.g., `install.sh`, `setup.sh`, or `bootstrap.sh`). If not, you'll create one in Step 3.

### If they don't have one:
Walk them through creating one:

1. **Create the repo on GitHub:**
   - Suggest the name `dotfiles` (the conventional name)
   - Ask if they want it public or private
   - Have them create it via GitHub UI or `gh repo create`:
     ```bash
     gh repo create dotfiles --private --clone
     ```
   - If `gh` is not installed or not authenticated, walk them through the GitHub web UI instead

2. **Create a basic install script:**
   ```bash
   cd dotfiles
   cat > install.sh << 'EOF'
   #!/bin/bash
   set -euo pipefail

   # Teeleport setup will be added here in Step 3
   echo "dotfiles setup complete!"
   EOF
   chmod +x install.sh
   ```

3. **Initial commit and push:**
   ```bash
   git add -A
   git commit -m "Initial dotfiles setup"
   git push origin main
   ```

---

## Step 2: Configure VS Code Dotfiles

VS Code and GitHub Codespaces need to know about the dotfile repo so they clone it into every new devcontainer.

**Ask the user:**
> Which environment do you primarily use — VS Code with Dev Containers, GitHub Codespaces, or both?

### For VS Code Dev Containers:
Guide them to add this to their VS Code `settings.json` (User settings, not workspace):

```json
{
  "dotfiles.repository": "https://github.com/USERNAME/dotfiles.git",
  "dotfiles.installCommand": "install.sh"
}
```

Tell them to replace `USERNAME` with their GitHub username. They can open settings with `Ctrl+,` (or `Cmd+,` on Mac) and search for "dotfiles".

### For GitHub Codespaces:
Guide them to:
1. Go to https://github.com/settings/codespaces
2. Under "Dotfiles", select their dotfiles repository
3. Check "Automatically install dotfiles"

### For both:
Do both of the above.

---

## Step 3: Add Teeleport

Now add Teeleport to their dotfile repo. Ask questions to build their config.

Read the https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/README.md for full configuration details.

### 3a. Add the install command

Add this line to the **top** of their `install.sh` (or create the file if it doesn't exist):

```bash
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash
```

### 3b. Build the Teeleport config

**Ask the user these questions one at a time:**

#### Packages
> Are there any system packages you always want installed in your devcontainers? (e.g., `jq`, `ripgrep`, `htop`, `vim`)

If yes, note the list. If no, skip the packages section.

#### AI CLI Tools
> Which AI coding CLI tools do you use? 
Map their choices to `ai_cli` entries.

#### Mount Presets
> Do you want to persist your AI tool settings across workspaces? This requires an SSH-accessible server with FUSE support in your devcontainer.

If yes:
> What is the hostname of your SSH server? (e.g., `my-server.example.com`)
> What SSH user should be used? (Leave blank to use your current user)
> What SSH port? (Leave blank for default 22)

Then ask which presets to enable based on their AI CLI choices:
- If they chose Claude Code → suggest `claude` preset
- If they chose Codex → suggest `codex` preset
- If they chose Gemini → suggest `gemini` preset
- If they chose Copilot → suggest `copilot` preset
- Always offer the `gh` preset for GitHub CLI auth persistence

#### Config Files
> Do you have any config files in your dotfile repo you want copied into the container? (e.g., `.bashrc`, `.gitconfig`, `.vimrc`)

If yes, ask for each file:
- What's the source path (relative to the dotfile repo)?
- What's the target path? (e.g., `~/.bashrc`)
- Should it replace the target or append to it?

### 3c. Generate the config file

Based on their answers, generate a `teeleport.config.yaml` file in their dotfile repo. Use this template, including only the sections they need:

```yaml
# Teeleport configuration
# https://github.com/BenjaminBenetti/Teeleport

packages:
  - <package1>
  - <package2>

copies:
  - name: <name>
    source: <source>
    target: <target>
    mode: <replace|append>

mounts:
  ssh:
    host: <hostname>
    user: <username>
    port: <port>
  entries:
    - name: <preset-name>
      preset: <preset>

ai_cli:
  - tool: <tool-name>
```

Omit any sections the user doesn't need (e.g., if no packages, don't include the `packages:` section at all).

### 3d. Commit and push

After generating the config:

```bash
git add -A
git commit -m "Add Teeleport configuration"
git push origin main
```

---

## Step 4: Verify

Tell the user:

> Your setup is complete! To test it:
> 1. Open a new devcontainer or Codespace
> 2. Check `~/.teeleport/run.log` to see what Teeleport did
> 3. Verify your tools and config are in place

If they have mounts configured, remind them:
> For SSHFS mounts to work, your `devcontainer.json` needs FUSE access. Add one of:
> ```json
> "privileged": true
> ```
> or
> ```json
> "runArgs": ["--device=/dev/fuse"]
> ```

---

## Important Notes for the AI Agent

- Be conversational and patient. New users may not know what dotfiles are.
- If a user seems confused at any step, explain the concept before proceeding.
- Don't generate the config file until you've asked all the questions — the config should reflect their actual needs.
- If the user doesn't want mounts, that's fine — Teeleport works great with just packages, copies, and AI CLI tools.
- Always show the user what you're about to add/change and get confirmation before writing files.
