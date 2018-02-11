package main

import (
	"io"
	"errors"
)

type ReadableClosableBytes struct {
	s        []byte
	i        int64 // current reading index
}

func (r *ReadableClosableBytes) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

func (r *ReadableClosableBytes) Close() error {
	return nil
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (r *ReadableClosableBytes) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case io.SeekStart:
	case io.SeekCurrent:
		offset += r.i
	case io.SeekEnd:
		offset += int64(len(r.s))
	}

	if offset < 0 {
		return 0, errOffset
	}

	r.i = offset

	return offset, nil
}

func openAsset(name string) *ReadableClosableBytes {
	asset, err := Asset(name)
	fatalIfFailed(err)
	return &ReadableClosableBytes{ s: asset, i: 0 }
}
