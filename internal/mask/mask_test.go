package mask_test

import (
	"testing"

	"github.com/panic-at/envx/internal/mask"
	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: ""},
		{name: "one char", value: "a", want: "***"},
		{name: "three chars", value: "abc", want: "***"},
		{name: "four chars", value: "abcd", want: "a***"},
		{name: "seven chars", value: "abcdefg", want: "a***"},
		{name: "eight chars", value: "abcdefgh", want: "ab***gh"},
		{name: "twenty chars", value: "abcdefghijklmnopqrst", want: "ab***st"},
		{name: "unicode short", value: "日本語", want: "***"},
		{name: "unicode medium", value: "café", want: "c***"},
		{name: "unicode long", value: "héllo wörld!", want: "hé***d!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mask.Mask(tt.value)
			assert.Equal(t, tt.want, got)
			if tt.value != "" {
				assert.NotEqual(t, tt.value, got, "a non-empty value must never be returned verbatim")
			}
		})
	}
}
