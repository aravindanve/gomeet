package util

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestEncodeLiveKitDataJSON(t *testing.T) {
	t.Parallel()

	a := map[string]string{
		"hello": "world",
	}
	m, err := EncodeLiveKitDataJSON(a)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	i, d := m[0], m[1:]
	if i != byte(LiveKitDataType_JSON) {
		t.Errorf("expected data type to be %v got %v", LiveKitDataType_JSON, i)
		return
	}

	var b map[string]string
	err = json.Unmarshal(d, &b)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected data to be %#v got %#v", a, b)
		return
	}
}
