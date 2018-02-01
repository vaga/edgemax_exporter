package edgemax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type stat struct {
	Name string `json:"name"`
}

type connectRequest struct {
	Subscribe   []stat `json:"SUBSCRIBE"`
	Unsubscribe []stat `json:"UNSUBSCRIBE"`
	SessionID   string `json:"SESSION_ID"`
}

func marshalWS(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	blen := []byte(strconv.Itoa(len(b)) + "\n")
	return append(blen, b...)
}

func unmarshalWS(data []byte, v interface{}) error {
	if data[0] == '{' {
		return json.Unmarshal(data, v)
	}

	bb := bytes.SplitN(data, []byte("\n"), 2)
	if l := len(bb); l != 2 {
		return fmt.Errorf("incorrect number of elements in websocket message: %d", l)
	}

	if len(bb[1]) == 0 {
		return nil
	}

	return json.Unmarshal(bb[1], v)
}
