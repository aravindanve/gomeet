package config

import (
	"bytes"
	"encoding/base64"
	"os"
	"strconv"
	"testing"
)

func TestGetenvStringUnset(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e == nil {
			t.Errorf("expected GetenvString to panic")
			return
		}
	}()

	GetenvString("UNSET")
}

func TestGetenvStringWithDefault(t *testing.T) {
	t.Parallel()
	a := "hello"

	if b := GetenvStringWithDefault("UNSET", a); b != a {
		t.Errorf("expected UNSET to be %v got %v", a, b)
	}
}

func TestGetenvStringSet(t *testing.T) {
	t.Parallel()
	a := "hello"
	os.Setenv("TEST_STRING", a)

	if b := GetenvString("TEST_STRING"); b != a {
		t.Errorf("expected TEST_STRING to be %v got %v", a, b)
	}
}

func TestGetenvIntUnset(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e == nil {
			t.Errorf("expected GetenvInt to panic")
			return
		}
	}()

	GetenvInt("UNSET")
}

func TestGetenvIntWithDefault(t *testing.T) {
	t.Parallel()
	a := 42

	if b := GetenvIntWithDefault("UNSET", a); b != a {
		t.Errorf("expected UNSET to be %v got %v", a, b)
	}
}

func TestGetenvIntSet(t *testing.T) {
	t.Parallel()
	a := 42
	os.Setenv("TEST_INT", strconv.Itoa(a))

	if b := GetenvInt("TEST_INT"); b != a {
		t.Errorf("expected TEST_INT to be %v got %v", a, b)
	}
}

func TestGetenvBoolUnset(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e == nil {
			t.Errorf("expected GetenvBool to panic")
			return
		}
	}()

	GetenvBool("UNSET")
}

func TestGetenvBoolWithDefault(t *testing.T) {
	t.Parallel()
	a := true

	if b := GetenvBoolWithDefault("UNSET", a); b != a {
		t.Errorf("expected UNSET to be %v got %v", a, b)
	}
}

func TestGetenvBoolSet(t *testing.T) {
	t.Parallel()
	a := true
	os.Setenv("TEST_BOOL", strconv.FormatBool(a))

	if b := GetenvBool("TEST_BOOL"); b != a {
		t.Errorf("expected TEST_BOOL to be %v got %v", a, b)
	}
}

func TestGetenvStringBase64Unset(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e == nil {
			t.Errorf("expected GetenvStringBase64 to panic")
			return
		}
	}()

	GetenvStringBase64("UNSET")
}

func TestGetenvStringBase64WithDefault(t *testing.T) {
	t.Parallel()
	a := []byte("hello")

	if b := GetenvStringBase64WithDefault("UNSET", a); !bytes.Equal(b, a) {
		t.Errorf("expected UNSET to be %v got %v", string(a), string(b))
	}
}

func TestGetenvStringBase64Set(t *testing.T) {
	t.Parallel()
	a := []byte("hello")
	os.Setenv("TEST_STRING", base64.StdEncoding.EncodeToString(a))

	if b := GetenvStringBase64("TEST_STRING"); !bytes.Equal(b, a) {
		t.Errorf("expected TEST_STRING to be %v got %v", string(a), string(b))
	}
}
