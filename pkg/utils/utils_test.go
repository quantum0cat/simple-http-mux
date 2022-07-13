package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "default",
			input: []string{"a", "b", "a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "empty",
			input: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, RemoveDuplicates(tt.input), tt.want, "Lists doesn't match")
		})
	}
}

func TestWithRid(t *testing.T) {

	tests := []struct {
		name  string
		input string
		rid   uint32
		want  string
	}{
		{
			name:  "default",
			input: "test",
			rid:   0,
			want:  "test [rid=0]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WithRid(tt.input, tt.rid), "strings don't match")
		})
	}
}
