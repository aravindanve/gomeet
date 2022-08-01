package util

import "encoding/json"

const (
	LiveKitDataType_JSON LiveKitDataType = 1
)

type LiveKitDataType byte

func EncodeLiveKitDataJSON(d any) ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(LiveKitDataType_JSON)}, b...), nil
}
