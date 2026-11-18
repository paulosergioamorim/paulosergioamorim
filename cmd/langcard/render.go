package main

import (
	"fmt"
	"html"
	"strings"
)

const (
	width     = 495
	padding   = 24
	rowHeight = 30
	labelW    = 110
	barW      = 270
	headerH   = 56
)

var colors = map[string]string{
	"C":          "#555555",
	"C++":        "#f34b7d",
	"Go":         "#00ADD8",
	"Python":     "#3572A5",
	"JavaScript": "#f1e05a",
	"TypeScript": "#3178c6",
	"Java":       "#b07219",
	"Shell":      "#89e051",
	"Rust":       "#dea584",
	"Lua":        "#000080",
	"Vue":        "#41b883",
	"Dockerfile": "#384d54",
	"PLpgSQL":    "#336790",
	"SQL":        "#e38c00",
	"Makefile":   "#427819",
	"Other":      "#8b949e",
}

var fallback = []string{"#6e7781", "#a475f9", "#e16f24", "#0e8a16", "#bf3989", "#1f6feb"}

func colorFor(name string, index int) string {
	if c, ok := colors[name]; ok {
		return c
	}
	return fallback[index%len(fallback)]
}

const styles = `.t{font:600 16px system-ui,-apple-system,"Segoe UI",sans-serif;fill:#1f2328}
.l{font:13px system-ui,-apple-system,"Segoe UI",sans-serif;fill:#1f2328}
.p{font:12px system-ui,-apple-system,"Segoe UI",sans-serif;fill:#656d76}
.bg{fill:#ffffff;stroke:#d0d7de}
.tr{fill:#eaeef2}
@media (prefers-color-scheme: dark){
.t,.l{fill:#e6edf3}
.p{fill:#8d96a0}
.bg{fill:#0d1117;stroke:#30363d}
.tr{fill:#21262d}
}`

func render(title string, entries []entry) string {
	height := headerH + len(entries)*rowHeight + padding/2
	maxPct := entries[0].Pct
	for _, e := range entries {
		if e.Pct > maxPct {
			maxPct = e.Pct
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`, width, height, width, height, html.EscapeString(title))
	b.WriteString("\n<style>\n" + styles + "\n</style>\n")
	fmt.Fprintf(&b, `<rect class="bg" x="0.5" y="0.5" width="%d" height="%d" rx="6"/>`+"\n", width-1, height-1)
	fmt.Fprintf(&b, `<text class="t" x="%d" y="34">%s</text>`+"\n", padding, html.EscapeString(title))

	for i, e := range entries {
		y := headerH + i*rowHeight
		fill := float64(barW) * e.Pct / maxPct
		fmt.Fprintf(&b, `<text class="l" x="%d" y="%d">%s</text>`+"\n", padding, y+14, html.EscapeString(e.Name))
		fmt.Fprintf(&b, `<rect class="tr" x="%d" y="%d" width="%d" height="8" rx="4"/>`+"\n", padding+labelW, y+6, barW)
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%.1f" height="8" rx="4" fill="%s"/>`+"\n", padding+labelW, y+6, fill, colorFor(e.Name, i))
		fmt.Fprintf(&b, `<text class="p" x="%d" y="%d">%.1f%%</text>`+"\n", padding+labelW+barW+10, y+14, e.Pct)
	}
	b.WriteString("</svg>\n")
	return b.String()
}
