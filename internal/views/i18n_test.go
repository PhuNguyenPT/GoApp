package views

import (
	"reflect"
	"testing"
)

func TestT_NoEmptyStrings(t *testing.T) {
	for _, lang := range []string{LangEN, LangVI} {
		tr := T(lang)
		v := reflect.ValueOf(tr)
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i).Name
			val := v.Field(i).String()
			if val == "" {
				t.Errorf("T(%q).%s is empty", lang, field)
			}
		}
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	tr := T("fr")
	if tr.Home != T(LangEN).Home {
		t.Errorf("unknown lang should fall back to English")
	}
}

func TestT_LanguagesDiffer(t *testing.T) {
	en := T(LangEN)
	vi := T(LangVI)
	if reflect.DeepEqual(en, vi) {
		t.Error("EN and VI translations should not be identical")
	}
}
