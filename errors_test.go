package rison

import "testing"

type errorInLang interface {
	error
	ErrorInLang(lang string) string
	Langs() []string
}

type translatable interface {
	error
	Translate(lang string)
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

func TestParseError_Translate(t *testing.T) {
	_, err := Decode([]byte(`(`), Rison)
	e, _ := err.(translatable)
	e.Translate("ja")
	want := `"(" が閉じていません (場所: 文字列終端: "(" → EOS)`
	if e.Error() != want {
		t.Errorf(`(*ParseError).Error: want %s, got %s`, want, e.Error())
	}
}
