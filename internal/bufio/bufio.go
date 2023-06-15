// Package bufio provides specialised packet buffers.
package bufio

import (
	"bytes"
	"errors"
	"io"
)

// ErrNegativeCount is returned when a negative number of bytes is requested.
var ErrNegativeCount = errors.New("bufio: negative count")

// PacketReader implements buffering for an io.Reader object.
// PacketReader will not read from the underlying reader unless its
// buffer is empty, reducing the risk of cross packet reading.
type PacketReader struct {
	rd   io.Reader
	buf  []byte
	r, w int
}

// NewPacketReader returns a new PacketReader whose buffer is size.
func NewPacketReader(rd io.Reader, size int) *PacketReader {
	r := &PacketReader{}
	r.reset(make([]byte, size), rd)
	return r
}

// Reset discards any buffered data, resets all state, and switches
// the buffered reader to read from rd.
// Calling Reset on the zero value of Reader will panic.
func (r *PacketReader) Reset(rd io.Reader) {
	if r.buf == nil {
		panic(errors.New("bufio: uninitialized PacketBuffer"))
	}
	r.reset(r.buf, rd)
}

func (r *PacketReader) reset(buf []byte, rd io.Reader) {
	*r = PacketReader{
		buf: buf,
		rd:  rd,
	}
}

const maxConsecutiveEmptyReads = 3

func (r *PacketReader) fill() error {
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
func (r *PacketReader) Buffered() int {
	return r.w - r.r
}

// Peek reads data in p without advancing the reader.
// It returns the number of bytes read into p.
func (r *PacketReader) Peek(p []byte) (int, error) {
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
func (r *PacketReader) Discard(n int) (int, error) {
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
func (r *PacketReader) Read(p []byte) (int, error) {
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
func (r *PacketReader) ReadByte() (byte, error) {
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
func (r *PacketReader) ReadSlice(delim byte) ([]byte, error) {
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

// PacketWriter implements buffering for an io.Writer object.
// PacketWriter will grow to the needed size and will not
// writer to the io.Writer until Flush is called.
type PacketWriter struct {
	wr  io.Writer
	buf []byte
}

// NewPacketWriter returns a new PacketWriter.
func NewPacketWriter(wr io.Writer) *PacketWriter {
	return &PacketWriter{wr: wr}
}

// Reset discards any buffered data, resets all state, and switches
// the buffered writer to write to wr.
func (w *PacketWriter) Reset(wr io.Writer) {
	w.wr = wr
	if w.buf != nil {
		w.buf = w.buf[:0]
	}
}

// Write writes p to the buffer.
func (w *PacketWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

// WriteByte writes the byte b to the buffer.
func (w *PacketWriter) WriteByte(b byte) error {
	w.buf = append(w.buf, b)
	return nil
}

// Flush writes the buffer to the writer and resets all state.
func (w *PacketWriter) Flush() error {
	if len(w.buf) == 0 {
		return nil
	}
	n, err := w.wr.Write(w.buf)
	if n < len(w.buf) && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		return err
	}
	w.buf = w.buf[:0]
	return nil
}
