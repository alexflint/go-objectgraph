package memusage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type T struct {
	X int
	Y *U
}

type U struct {
	A string
	B []int64
}

func TestMemusage(t *testing.T) {
	obj := T{
		X: 3,
		Y: &U{
			A: "abababa",
			B: []int64{3, 2, 1},
		},
	}
	assert.Equal(t, 143, int(Bytes(obj)))
}
