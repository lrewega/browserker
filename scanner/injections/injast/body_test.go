package injast_test

import (
	"bytes"
	"testing"

	"gitlab.com/browserker/scanner/injections/parsers"
)

func TestCopyBody(t *testing.T) {
	p := &parsers.BodyParser{}
	b, _ := p.Parse([]byte("x=1&y=1"))
	newBody := b.Copy()
	if bytes.Compare(newBody.Original, b.Original) != 0 {
		t.Fatalf("%s not copied %s\n", b.Original, newBody.Original)
	}

	newBody.Original[1] = 'x'
	if bytes.Compare(newBody.Original, b.Original) == 0 {
		t.Fatalf("%s was updated %s\n", b.Original, newBody.Original)
	}

}

func TestBodySerialization(t *testing.T) {
	p := &parsers.BodyParser{}

	inputs := []string{
		"{\"asdf\": \"blah\"}",
		"x=1&y=2&f=asdf",
		"",
		"x[0]=y",
	}
	for _, v := range inputs {
		b, _ := p.Parse([]byte(v))
		if b.String() != v {
			t.Fatalf("exp: %s got %s\n", v, b.String())
		}
	}

}

func TestJSONBodyParts(t *testing.T) {
	p := &parsers.BodyParser{}
	b, _ := p.Parse([]byte("{\"asdf\": \"blah\"}"))
	if !b.IsJSON() {
		t.Fatalf("expected json")
	}
}
