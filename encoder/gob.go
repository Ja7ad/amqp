package encoder

import (
	"bytes"
	"encoding/gob"
)

// GobEncoder is a Go specific GOB Encoder implementation for EncodedConn.
// This encoder will use the builtin encoding/gob to Marshal
// and Unmarshal most types, including structs.
type GobEncoder struct{}

func (g *GobEncoder) Encode(v any) ([]byte, error) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (g *GobEncoder) Decode(data []byte, vPtr any) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	return dec.Decode(vPtr)
}
