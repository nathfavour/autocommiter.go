# ✍️ Autocommiter

Generating commit messages manually? That's so yesterday. **Autocommiter** uses AI to analyze your staged changes and write meaningful commit messages for you.

Built with native Go for speed, portability, and zero-config vibe.

---

## 🚀 Installation

### Quick Install (Recommended)
```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/autocommiter.go/master/install.sh | bash
```

### Building from Source
Requires **Go 1.21** or later.
```bash
git clone https://github.com/nathfavour/autocommiter.go.git
cd autocommiter.go
go build -o autocommiter ./cmd/autocommiter
```

## 🛠️ Setup

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

#### 📁 Project-level Config
You can also create a `.autocommiter.json` in your repository root to override global settings for a specific project:
```json
{
  "selected_model": "gpt-4o",
  "enable_gitmoji": true
}
```

### 🧹 Maintenance
- `autocommiter update`: Self-update to the latest version
- `autocommiter clean`: Wipe all data and configuration
- `autocommiter uninstall`: Remove binary
- `autocommiter uninstall --clean`: Full wipe (binary + data)

---
*Clean commits. Zero friction.*
