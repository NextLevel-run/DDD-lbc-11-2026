package domain

import (
	"strings"
	"unicode"
)

// sansAccent maps the accented letters of French (and its neighbours) to
// their bare form. The standard library has no Unicode decomposition, and
// this short table covers what a vendeur types.
var sansAccent = map[rune]string{
	'à': "a", 'â': "a", 'ä': "a", 'á': "a", 'ã': "a", 'å': "a",
	'ç': "c",
	'è': "e", 'ê': "e", 'ë': "e", 'é': "e",
	'î': "i", 'ï': "i", 'í': "i", 'ì': "i",
	'ô': "o", 'ö': "o", 'ó': "o", 'ò': "o", 'õ': "o",
	'ù': "u", 'û': "u", 'ü': "u", 'ú': "u",
	'ÿ': "y", 'ý': "y",
	'ñ': "n",
	'œ': "oe", 'æ': "ae",
}

// normaliser folds a text for comparison: lower case, no accent, hyphens and
// apostrophes counted as spaces, runs of spaces collapsed to one.
func normaliser(texte string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(texte) {
		if bare, ok := sansAccent[r]; ok {
			b.WriteString(bare)
			continue
		}
		switch r {
		case '-', '\'', '’':
			b.WriteRune(' ')
		default:
			if unicode.IsSpace(r) {
				b.WriteRune(' ')
			} else {
				b.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
