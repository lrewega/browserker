package browserk_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
)

func TestInjectorLocation(t *testing.T) {
	l := browserk.InjectAll
	if !l.Has(browserk.InjectPath) {
		t.Fatalf("should have path")
	}

	l = browserk.InjectCommon
	if l.Has(browserk.InjectMethod) {
		t.Fatalf("should not have method")
	}
}

func TestInjectorLocationHasIn(t *testing.T) {
	l := []browserk.InjectionLocation{browserk.InjectValues}
	if browserk.InjectPath.HasIn(l) {
		t.Fatalf("should not have path")
	}

	if !browserk.InjectQueryValue.HasIn(l) {
		t.Fatalf("should have query value")
	}
}
