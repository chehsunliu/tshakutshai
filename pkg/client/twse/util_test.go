package twse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuffixDuplicateFields(t *testing.T) {
	fields1 := []string{"a", "b", "c"}
	assert.Equal(t, []string{"a", "b", "c"}, suffixDuplicateFields(fields1))

	fields2 := []string{"a", "b", "c", "b", "b", "c"}
	assert.Equal(t, []string{"a", "b", "c", "b2", "b3", "c2"}, suffixDuplicateFields(fields2))
}
