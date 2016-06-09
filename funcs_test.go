package sq

import (
	"errors"
	"testing"
	"time"
)

func TestNilFuncs(t *testing.T) {

	s, err := (&parser{}).parse("string")
	if err != nil {
		t.Errorf("Expected nil, got %q", err)
	}
	if s != "string" {
		t.Errorf("Expected %q, got %q", "string", s)
	}

	v, err := (&loader{}).load(nil, "string")
	if err != nil {
		t.Errorf("Expected nil, got %q", err)
	}
	if v != "string" {
		t.Errorf("Expected %q, got %q", "string", v)
	}

}

func TestRegexp(t *testing.T) {

	f := parseFuncs["regexp"]

	tests := []struct {
		input, pattern, output string
		err                    error
	}{
		{"test", "s", "s", nil},
		{"the rain in spain falls", " in (\\w+)", "spain", nil},
		{"the rain in spain falls", " in (\\w+)", "spain", nil},
		{"the rain in spain falls", " rain(\\S?) ", "spain", ErrNoRegexpMatch},
		{"the rain in spain falls", " dogs ", "spain", ErrNoRegexpMatch},
		{"the rain in spain falls", "(", "spain", errors.New("error parsing regexp: missing closing ): `(`")},
	}

	for _, test := range tests {
		ouput, err := f(test.input, test.pattern)
		if err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("Expected %q, got %q", test.err, err)
			}
			continue
		}
		if ouput != test.output {
			t.Errorf("Expected %q, got %q", test.output, ouput)
		}
	}

}

func TestTime(t *testing.T) {

	f := loadFuncs["time"]

	tests := []struct {
		input, layout string
		output        time.Time
	}{
		{"2006", "2006", time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2016 05 23", "2006 01 02", time.Date(2016, 5, 23, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		ouput, err := f(nil, test.input, test.layout)
		if err != nil {
			t.Error(err)
		}
		if ouput.(time.Time) != test.output {
			t.Errorf("Expected %v, got %v", test.output, ouput)
		}
	}

}
