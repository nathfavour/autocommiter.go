# ✍️ Autocommiter

Generating commit messages manually? That's so yesterday. **Autocommiter** uses AI to analyze your staged changes and write meaningful commit messages for you.

Built with native Go for speed, portability, and zero-config vibe.

---

### 🚀 Quick Install

Install the latest binary for your platform (Linux, macOS, Windows, Android/Termux):

```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/autocommiter.go/master/install.sh | bash
```

### 🛠️ Getting Started

1. **Set your API Key** (GitHub Models API):
   ```bash
   autocommiter set-api-key
   ```
2. **Commit with style**:
   ```bash
   git add .
   autocommiter
   ```

### ⚙️ Config
- `autocommiter toggle-gitmoji` - Enable/disable emojis ✨
- `autocommiter select-model` - Choose your favorite AI model
- `autocommiter toggle-auto-update` - Keep it fresh

### 🧹 Uninstallation

- **Remove binary**: `autocommiter uninstall`
- **Full wipe**: `autocommiter uninstall --clean`

---
*Clean commits. Zero friction.*
