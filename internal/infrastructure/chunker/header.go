package chunker

import (
	"fmt"
	"path/filepath"
	"strings"
)

const maxTitleRunes = 80

// DocTitle derives a display title: the first non-empty line of the first
// page with leading Markdown "#" and spaces removed, whitespace collapsed and
// truncated to 80 runes. It falls back to the file name without extension.
func DocTitle(firstPage, filename string) string {
	for _, line := range strings.Split(firstPage, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		title := strings.Join(strings.Fields(strings.TrimLeft(strings.TrimSpace(line), "# \t")), " ")
		if r := []rune(title); len(r) > maxTitleRunes {
			title = string(r[:maxTitleRunes])
		}
		if title != "" {
			return title
		}
		break
	}
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

// ContextHeader returns the one-line header "[title — p. N]", or "[title]"
// when page is not positive. The caller prepends it plus a newline to a chunk.
func ContextHeader(title string, page int) string {
	if page > 0 {
		return fmt.Sprintf("[%s — p. %d]", title, page)
	}
	return fmt.Sprintf("[%s]", title)
}
