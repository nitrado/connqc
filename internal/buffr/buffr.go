// Package buffr provides a buffered reader.
package buffr

import (
	"bytes"
	"errors"
	"io"
)

// ErrNegativeCount is returned when a negative number of bytes is requested.
var ErrNegativeCount = errors.New("buffr: negative count")

// Reader implements buffering for an io.Reader object.
// Reader does not read from the underlying reader unless its
// buffer is empty, reducing the risk of cross-packet reading.
type Reader struct {
	rd   io.Reader
	buf  []byte
	r, w int
}

// NewReader returns a new Reader with a buffer of the length specified by size.
func NewReader(rd io.Reader, size int) *Reader {
	r := &Reader{}
	r.reset(make([]byte, size), rd)
	return r
}

// Reset discards any buffered data, resets all state, and switches
// the buffered reader to read from rd.
func (r *Reader) Reset(rd io.Reader) {
	if r.buf == nil {
		return
	}
	r.reset(r.buf, rd)
}

func (r *Reader) reset(buf []byte, rd io.Reader) {
	*r = Reader{
		buf: buf,
		rd:  rd,
	}
}

const maxConsecutiveEmptyReads = 3

func (r *Reader) fill() error {
	// Ignore any data in the buffer.
	r.r = 0
	r.w = 0

	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := r.rd.Read(r.buf)
		r.w += n
		if err != nil {
			return err
		}
		if n > 0 {
			return nil
		}
	}
	return io.ErrNoProgress
}

// Buffered returns the number of bytes that can be read from the current buffer.
func (r *Reader) Buffered() int {
	return r.w - r.r
}

// Peek reads data in p without advancing the reader.
// It returns the number of bytes read into p.
func (r *Reader) Peek(p []byte) (int, error) {
	if r.r == r.w {
		err := r.fill()
		if err != nil {
			return 0, err
		}
	}

	n := copy(p, r.buf[r.r:r.w])
	return n, nil
}

// Discard skips the next n bytes, returning the number of bytes discarded.
// If Discard skips fewer than n bytes, it also returns an error.
func (r *Reader) Discard(n int) (int, error) {
	if n < 0 {
		return 0, ErrNegativeCount
	}
	if n == 0 {
		return 0, nil
	}

	var err error
	skip := n
	if bn := r.Buffered(); skip > bn {
		skip = bn
		err = io.EOF
	}
	r.r += skip
	return skip, err
}

// Read reads data in p.
// It returns the number of bytes read into p.
func (r *Reader) Read(p []byte) (int, error) {
	if r.r == r.w {
		err := r.fill()
		if err != nil {
			return 0, err
		}
	}

	n := copy(p, r.buf[r.r:r.w])
	r.r += n
	return n, nil
}

// ReadByte reads the next byte.
func (r *Reader) ReadByte() (byte, error) {
	if r.r == r.w {
		err := r.fill()
		if err != nil {
			return 0, err
		}
	}

	c := r.buf[r.r]
	r.r++
	return c, nil
}

// ReadSlice reads until the first occurrence of delim in the input,
// returning a slice pointing at the bytes in the buffer.
// If ReadSlice encounters the end of the buffer before finding the delimiter,
// it returns an io.EOF error.
func (r *Reader) ReadSlice(delim byte) ([]byte, error) {
	if r.r == r.w {
		err := r.fill()
		if err != nil {
			return nil, err
		}
	}

	i := bytes.IndexByte(r.buf[r.r:r.w], delim)
	if i < 0 {
		r.r = r.w
		return nil, io.EOF
	}

	line := r.buf[r.r : r.r+i+1]
	r.r += i + 1
	return line, nil
}
