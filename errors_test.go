package rison

import "testing"

type errorInLang interface {
	ErrorInLang(lang string) string
}

func TestParseError_Error(t *testing.T) {
	_, err := Decode([]byte(`(`), Rison)
	want := `unmatched "(" (at the end of string "(" -> EOS)`
	if err.Error() != want {
		t.Errorf(`(*ParseError).Error: want %s, got %s`, want, err.Error())
	}
	e, _ := err.(errorInLang)
	if err.Error() != e.ErrorInLang("en") {
		t.Errorf(`(*ParseError).Error: want %s, got %s`, e.ErrorInLang("en"), err.Error())
	}
	if e.ErrorInLang("") != e.ErrorInLang("en") {
		t.Errorf(`(*ParseError).ErrorInLang: want %s, got %s`, e.ErrorInLang("en"), e.ErrorInLang(""))
	}
}
