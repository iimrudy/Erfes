package utils

import "io"

type PassThru struct {
	io.Reader
	TransferredBytes int64       `json:"-"` /*ignore from json*/
	CallBack         func(int64) `json:"-"` /*ignore from json*/
}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.TransferredBytes += int64(n)
	if pt.CallBack != nil {
		pt.CallBack(pt.TransferredBytes)
	}
	return n, err
}
