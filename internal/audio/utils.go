package audio

import (
	"bytes"
	"fmt"
	"io"
)

func toReadSeeker(r io.Reader) (io.ReadSeeker, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error while converting to byte slice: %w", err)
	}
	seekableReader := bytes.NewReader(data)
	return seekableReader, nil
}
