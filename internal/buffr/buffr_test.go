package buffr_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nitrado/connqc/internal/buffr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPacketReader(t *testing.T) {
	rd := bytes.NewReader([]byte("this is a test:"))

	r := buffr.NewReader(rd, 20)
	r.Reset(rd)

	var pb [5]byte
	n, err := r.Peek(pb[:])
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "this ", string(pb[:]))

	n, err = r.Discard(2)
	require.NoError(t, err)
	assert.Equal(t, n, 2)

	var rb [5]byte
	n, err = r.Read(rb[:])
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "is is", string(rb[:]))

	n = r.Buffered()
	assert.Equal(t, 8, n)

	b, err := r.ReadByte()
	require.NoError(t, err)
	assert.Equal(t, byte(' '), b)

	sb, err := r.ReadSlice(':')
	require.NoError(t, err)
	assert.Equal(t, []byte("a test:"), sb)

	n = r.Buffered()
	assert.Equal(t, 0, n)
}

func TestPacketReader_DiscardHandlesShortBuffer(t *testing.T) {
	rd := bytes.NewReader([]byte("test"))

	r := buffr.NewReader(rd, 20)

	var pb [5]byte
	_, err := r.Peek(pb[:])
	require.NoError(t, err)

	n, err := r.Discard(5)

	require.ErrorIs(t, err, io.EOF)
	assert.Equal(t, n, 4)
}

func TestPacketReader_ReadSliceHandlesShortBuffer(t *testing.T) {
	rd := bytes.NewReader([]byte("test"))

	r := buffr.NewReader(rd, 20)

	sb, err := r.ReadSlice(':')

	require.ErrorIs(t, err, io.EOF)
	assert.Nil(t, sb)
}
