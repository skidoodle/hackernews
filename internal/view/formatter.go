package view

import (
	"html/template"
	"regexp"
	"strings"
)

var (
	italicRegex = regexp.MustCompile(`(^|[^\*])\*([^\*]+)\*([^\*]|$)`)
	urlRegex    = regexp.MustCompile(`(<(https?:\/\/[^\s<>"]+)>|https?:\/\/[^\s<>"]+)`)
)

func FormatText(text string) template.HTML {
	if text == "" {
		return ""
	}

	var result strings.Builder
	paragraphs := strings.SplitSeq(text, "\n\n")

	for p := range paragraphs {
		if strings.TrimSpace(p) == "" {
			continue
		}

		if after, ok := strings.CutPrefix(p, "  "); ok {
			codeContent := after
			escapedCode := template.HTMLEscapeString(codeContent)
			result.WriteString("<pre><code>" + escapedCode + "</code></pre>")
			continue
		}

		p = urlRegex.ReplaceAllStringFunc(p, func(match string) string {
			url := strings.Trim(match, "<>")
			return `<a href="` + url + `" rel="nofollow">` + template.HTMLEscapeString(url) + `</a>`
		})

		p = italicRegex.ReplaceAllString(p, "$1<i>$2</i>$3")

		p = strings.ReplaceAll(p, `\*`, `*`)

		p = strings.ReplaceAll(p, "\n", "<br>")

		result.WriteString("<p>" + p + "</p>")
	}

	return template.HTML(result.String())
}
