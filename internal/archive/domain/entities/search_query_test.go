package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizeSearchQuery(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "empty", raw: "  ", want: nil},
		{name: "phrase", raw: "cita médica", want: []string{"cita", "médica"}},
		{name: "certificado", raw: "certificado", want: []string{"certificado"}},
		{name: "year", raw: "1998", want: []string{"1998"}},
		{name: "drops single letter", raw: "a factura", want: []string{"factura"}},
		{name: "keeps single digit", raw: "1 predial", want: []string{"1", "predial"}},
		{name: "collapses spaces", raw: "  gas   efigas  ", want: []string{"gas", "efigas"}},
		{name: "caps at eight", raw: "aa bb cc dd ee ff gg hh ii jj", want: []string{"aa", "bb", "cc", "dd", "ee", "ff", "gg", "hh"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TokenizeSearchQuery(tt.raw))
		})
	}
}

func TestIsYearToken(t *testing.T) {
	assert.True(t, IsYearToken("1998"))
	assert.True(t, IsYearToken("2026"))
	assert.False(t, IsYearToken("98"))
	assert.False(t, IsYearToken("199a"))
	assert.False(t, IsYearToken("19980"))
	assert.False(t, IsYearToken(""))
}
