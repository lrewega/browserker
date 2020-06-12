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
}
