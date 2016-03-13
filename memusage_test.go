package memusage

import (
	"log"
	"testing"

	"github.com/dustin/go-humanize"
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
			A: "123456",
			B: []int64{3, 2, 1},
		},
	}
	p := NewProfile(&obj)
	for t, sz := range p.sizeByType {
		log.Printf("  %8s %s", humanize.Bytes(sz), t.String())
	}
	assert.Equal(t, 56, int(p.TotalBytes))
}
