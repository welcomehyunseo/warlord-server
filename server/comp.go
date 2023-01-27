package server

import (
	"bytes"
	"compress/zlib"
	"io"
)

func Compress(
	arr []uint8,
) (*bytes.Buffer, error) {
	in := bytes.NewBuffer(nil)
	w := zlib.NewWriter(in)
	if _, err := w.Write(arr); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return in, nil
}

func Uncompress(
	arr []uint8,
) (*bytes.Buffer, error) {
	out := bytes.NewBuffer(nil)
	buf := bytes.NewBuffer(arr)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(out, r); err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}

	return out, nil
}
