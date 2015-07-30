package cli

import (
	"io"
	"time"
	mrand "math/rand"
	crand "crypto/rand"
)

type RandReaders struct {
	buf []byte
}

func NewRandReaders(size int) (p *RandReaders) {
	buf := make([]byte, size)
	io.ReadFull(crand.Reader, buf)
	return &RandReaders{buf}
}

func (p *RandReaders) NewRandReader(size int64) *RandReader {
	from := Rand(0, len(p.buf))
	return &RandReader{p.buf[from:], p.buf[:from], 0, len(p.buf), 0, size, int64(from)}
}

type RandReader struct {
	p1 []byte
	p2 []byte
	curr int
	blen int
	pos int64
	size int64
	from int64
}

func (rr *RandReader) Reset() {
	rr.pos = 0
}

func (rr *RandReader) Read(p []byte) (read int, err error) {
	pcap := len(p)

	min := func(a, b int) int {
		if a > b {
			return b
		}
		return a
	}

	get := func(b []byte) {
		r := min(int(rr.size - rr.pos), pcap - read)
		r = min(r, len(b))
		copy(p[read:], b[:r])
		read += r
		rr.pos += int64(r)
	}

	fetch := func() {
		curr := int(rr.pos) % rr.blen
		if curr < len(rr.p1) {
			get(rr.p1[curr:])
		} else {
			get(rr.p2[curr - len(rr.p1):])
		}
	}

	for {
		if rr.size == rr.pos {
			if pcap != read {
				return read, io.EOF
			} else {
				return
			}
		}
		if pcap == read {
			return
		}
		fetch()
	}
}

func Rand(min, max int) int {
	if min > max {
		min, max = max, min
	}
	val := max
	if max != min {
		val = _rand.Intn(max - min) + min
	}
	return val
}

var _rand = mrand.New(mrand.NewSource(time.Now().UnixNano()))
