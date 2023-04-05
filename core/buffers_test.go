package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffers(t *testing.T) {
	bufs := NewBuffers(1, 2)
	assert.Equal(t, 1, *bufs.Current())

	ref := bufs.Current()
	bufs.Swap()
	*bufs.Current() = 9
	assert.Equal(t, 1, *ref)

	bufs.Swap()
	*bufs.Current() = 8
	assert.Equal(t, 8, *ref)
}
