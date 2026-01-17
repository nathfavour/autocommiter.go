package gitmoji

import (
	"math/rand"
	"strings"
	"time"
)

type Gitmoji struct {
	Emoji       string   `json:"emoji"`
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

var GITMOJIS = []Gitmoji{
	{Emoji: "🎨", Code: ":art:", Description: "Improve structure/format", Keywords: []string{"format", "structure", "style", "lint"}},
	{Emoji: "⚡", Code: ":zap:", Description: "Improve performance", Keywords: []string{"performance", "speed", "optimize", "fast"}},
	{Emoji: "🔥", Code: ":fire:", Description: "Remove code/files", Keywords: []string{"remove", "delete", "clean", "unused"}},
	{Emoji: "🐛", Code: ":bug:", Description: "Fix bug", Keywords: []string{"fix", "bug", "issue", "error", "crash"}},
	{Emoji: "✨", Code: ":sparkles:", Description: "New feature", Keywords: []string{"feature", "new", "add", "implement"}},
	{Emoji: "📝", Code: ":memo:", Description: "Add documentation", Keywords: []string{"docs", "documentation", "comment", "readme"}},
	{Emoji: "🚀", Code: ":rocket:", Description: "Deploy stuff", Keywords: []string{"deploy", "release", "publish", "launch"}},
	{Emoji: "💅", Code: ":nail_care:", Description: "Polish code", Keywords: []string{"polish", "refine", "improve"}},
	{Emoji: "✅", Code: ":white_check_mark:", Description: "Add tests", Keywords: []string{"test", "tests", "testing"}},
	{Emoji: "🔐", Code: ":lock:", Description: "Security fix", Keywords: []string{"security", "auth", "encrypt"}},
	{Emoji: "⬆️", Code: ":arrow_up:", Description: "Upgrade dependencies", Keywords: []string{"upgrade", "update", "dependency", "dependencies"}},
	{Emoji: "⬇️", Code: ":arrow_down:", Description: "Downgrade dependencies", Keywords: []string{"downgrade"}},
	{Emoji: "📦", Code: ":package:", Description: "Update packages", Keywords: []string{"package", "npm", "yarn", "bundler"}},
	{Emoji: "🔧", Code: ":wrench:", Description: "Configuration", Keywords: []string{"config", "configuration", "settings"}},
	{Emoji: "🌐", Code: ":globe_with_meridians:", Description: "i18n/localization", Keywords: []string{"i18n", "translation", "locale", "language"}},
	{Emoji: "♿", Code: ":wheelchair:", Description: "Accessibility", Keywords: []string{"accessibility", "a11y", "aria"}},
	{Emoji: "🚨", Code: ":rotating_light:", Description: "Fix warnings", Keywords: []string{"warning", "lint"}},
	{Emoji: "🔍", Code: ":mag:", Description: "SEO", Keywords: []string{"seo"}},
	{Emoji: "🍎", Code: ":apple:", Description: "macOS fix", Keywords: []string{"macos", "mac", "apple"}},
	{Emoji: "🐧", Code: ":penguin:", Description: "Linux fix", Keywords: []string{"linux", "ubuntu"}},
	{Emoji: "🐍", Code: ":snake:", Description: "Python changes", Keywords: []string{"python", "django", "flask", "pip", "pytorch"}},
	{Emoji: "📚", Code: ":books:", Description: "Node.js/JavaScript", Keywords: []string{"node", "npm", "javascript", "express", "typescript"}},
	{Emoji: "🦀", Code: ":crab:", Description: "Rust changes", Keywords: []string{"rust", "cargo", "tokio", "wasm"}},
	{Emoji: "☕", Code: ":coffee:", Description: "Java changes", Keywords: []string{"java", "spring", "maven", "gradle", "jvm"}},
	{Emoji: "🐳", Code: ":whale:", Description: "Docker changes", Keywords: []string{"docker", "container", "dockerfile", "image"}},
	{Emoji: "🐹", Code: ":hamster:", Description: "Go changes", Keywords: []string{"go", "golang", "mod"}},
}

func calculateFuzzyScore(commitMessage string, gitmoji Gitmoji) uint32 {
	msg := strings.ToLower(commitMessage)
	var score uint32 = 0

	for _, keyword := range gitmoji.Keywords {
		if strings.Contains(msg, keyword) {
			score += 40
		}
		if len(keyword) >= 3 && strings.Contains(msg, keyword[:3]) {
			score += 10
		}
	}

	for _, word := range strings.Fields(strings.ToLower(gitmoji.Description)) {
		if len(word) > 2 && strings.Contains(msg, word) {
			score += 15
		}
	}

	if score > 100 {
		return 100
	}
	return score
}

func FindBestGitmoji(commitMessage string) *Gitmoji {
	if strings.TrimSpace(commitMessage) == "" {
		return nil
	}

	var bestGitmoji *Gitmoji
	var bestScore uint32 = 30

	for i := range GITMOJIS {
		score := calculateFuzzyScore(commitMessage, GITMOJIS[i])
		if score > bestScore {
			bestScore = score
			bestGitmoji = &GITMOJIS[i]
		}
	}

	return bestGitmoji
}

func GetRandomGitmoji() Gitmoji {
	rand.Seed(time.Now().UnixNano())
	return GITMOJIS[rand.Intn(len(GITMOJIS))]
}

func GetGitmojifiedMessage(commitMessage string) string {
	bestMatch := FindBestGitmoji(commitMessage)
	var gitmoji Gitmoji
	if bestMatch != nil {
		gitmoji = *bestMatch
	} else {
		gitmoji = GetRandomGitmoji()
	}
	return gitmoji.Emoji + " " + commitMessage
}
