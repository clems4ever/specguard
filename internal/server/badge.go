package server

import (
	"fmt"
	"html"
)

// badgeSVG renders a minimal two-part shields-style badge (label | message).
// Widths are approximated from character count — good enough for a status badge
// without bundling a font metrics table.
func badgeSVG(label, message, color string) []byte {
	lw := textWidth(label)
	mw := textWidth(message)
	total := lw + mw
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">`+
		`<linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>`+
		`<rect rx="3" width="%d" height="20" fill="#555"/>`+
		`<rect rx="3" x="%d" width="%d" height="20" fill="%s"/>`+
		`<rect rx="3" width="%d" height="20" fill="url(#s)"/>`+
		`<g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">`+
		`<text x="%d" y="14">%s</text>`+
		`<text x="%d" y="14">%s</text>`+
		`</g></svg>`,
		total, html.EscapeString(label), html.EscapeString(message),
		lw, lw, mw, color, total,
		lw/2, html.EscapeString(label),
		lw+mw/2, html.EscapeString(message),
	)
	return []byte(svg)
}

// textWidth approximates a rendered label width in pixels (11px sans).
func textWidth(s string) int {
	return len(s)*7 + 12
}
