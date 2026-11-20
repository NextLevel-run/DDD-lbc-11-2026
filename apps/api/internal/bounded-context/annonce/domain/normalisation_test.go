package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormaliser(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Vélo Électrique", "velo electrique"},
		{"  ÇA   va  ", "ca va"},
		{"Œuvre d'art", "oeuvre d art"},
		{"Saint-Étienne", "saint etienne"},
		{"àâäáãå èêëé îïíì ôöóòõ ùûüú ÿý ñ æ", "aaaaaa eeee iiii ooooo uuuu yy n ae"},
		{"Lit 140×190", "lit 140×190"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, normaliser(tt.in))
		})
	}
}
