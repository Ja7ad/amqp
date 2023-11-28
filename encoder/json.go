package encoder

import (
	"encoding/json"
	"strings"
)

type JsonEncoder struct{}

func (j *JsonEncoder) Encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (j *JsonEncoder) Decode(data []byte, vPtr any) error {
	switch arg := vPtr.(type) {
	case *string:
		str := string(data)
		if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
			*arg = str[1 : len(str)-1]
		} else {
			*arg = str
		}
	case *[]byte:
		*arg = data
	default:
		return json.Unmarshal(data, arg)
	}
	return nil
}
