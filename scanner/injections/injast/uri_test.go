package injast_test

import (
	"bytes"
	"testing"

	"gitlab.com/browserker/scanner/injections/parsers"
)

func TestCopy(t *testing.T) {
	p := &parsers.URIParser{}
	u, _ := p.Parse("/path1/path2?x=1#/something")
	newURI := u.Copy()
	if bytes.Compare(newURI.Original, u.Original) != 0 {
		t.Fatalf("%s not copied %s\n", u.Original, newURI.Original)
	}

	newURI.Original[1] = 'x'
	if bytes.Compare(newURI.Original, u.Original) == 0 {
		t.Fatalf("%s was updated %s\n", u.Original, newURI.Original)
	}

	newURI.File.Name = "path3"
	if newURI.File.Name == u.File.Name {
		t.Fatalf("updating new URI modified original")
	}
}

func TestURIParts(t *testing.T) {
	p := &parsers.URIParser{}
	u, _ := p.Parse("/path1/path2?x=1#/something")
	if u.PathOnly() != "/path1/" {
		t.Fatalf("path %s should equal /path1/", u.PathOnly())
	}
	if u.FileOnly() != "path2" {
		t.Fatalf("file %s should equal path2", u.FileOnly())
	}
}

// TODO: add fuzzing
func TestURIString(t *testing.T) {
	p := &parsers.URIParser{}
	inputs := []string{
		"/?#/asdf/1?x=y&y=z&a[]=1",
		"/a?#/asdf/",
		"/asdf?x=1&y=2",
		"/#asdf=asdf",
		"/?x[0]=1&x[1]=2",
		"/path1/path2?x=1#/something",
		"/path1/path2?x=1#?something",
	}
	for _, input := range inputs {
		u, _ := p.Parse(input)
		if u.String() != input {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input, u.String())
		}
	}
}
