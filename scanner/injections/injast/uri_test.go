package injast_test

import (
	"bytes"
	"net/url"
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
		unescaped, err := url.PathUnescape(u.String())
		if err != nil {
			t.Fatalf("failure unescaping: %s\n", err)
		}
		if unescaped != input {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input, u.String())
		}
	}
}

func TestReplaceFile(t *testing.T) {
	p := &parsers.URIParser{}
	var inputs = []struct {
		in       string
		newFile  string
		expected string
	}{
		{
			"/file/?x=1",
			"newfile",
			"/file/newfile?x=1",
		},
		{
			"/file?x=1",
			"",
			"/?x=1",
		},
		{
			"/file?x=1",
			"newfile",
			"/newfile?x=1",
		},
	}

	for _, input := range inputs {
		u, _ := p.Parse(input.in)
		u.ReplaceFile(input.newFile)
		if u.String() != input.expected {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input.in, u.String())
		}
	}
}

func TestReplacePath(t *testing.T) {
	p := &parsers.URIParser{}
	var inputs = []struct {
		in       string
		newPath  string
		index    int
		expected string
	}{
		{
			"/path/?x=1",
			"newpath",
			0,
			"/newpath/?x=1",
		},
		{
			"/path1/path2/?x=1",
			"newpath",
			1,
			"/path1/newpath/?x=1",
		},
		{
			"/path1/path2/?x=1",
			"newpath",
			8,
			"/path1/path2/?x=1",
		},
		{
			"/path1/path2/path3/file?x=1",
			"newpath",
			1,
			"/path1/newpath/path3/file?x=1",
		},
		{
			"/path1/path2/path3/file?x=1",
			"",
			1,
			"/path1/path3/file?x=1",
		},
	}

	for _, input := range inputs {
		u, _ := p.Parse(input.in)
		u.ReplacePath(input.newPath, input.index)
		if u.String() != input.expected {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input.in, u.String())
		}
	}
}

func TestReplaceParam(t *testing.T) {
	p := &parsers.URIParser{}
	var inputs = []struct {
		in       string
		original string
		newKey   string
		newVal   string
		expected string
	}{
		{
			"/file?x[0]=1",
			"x[0]",
			"y",
			"zoop",
			"/file?y[0]=zoop",
		},
		{
			"/file?x=1",
			"x",
			"x",
			"blah",
			"/file?x=blah",
		},
		{
			"/file?x=1",
			"notexist",
			"notexist",
			"notexist",
			"/file?x=1",
		},
		{
			"/file?x[0]=1",
			"x[0]",
			"x",
			"zoop",
			"/file?x[0]=zoop",
		},
	}

	for _, input := range inputs {
		u, _ := p.Parse(input.in)
		u.ReplaceParam(input.original, input.newKey, input.newVal)
		unescaped, err := url.PathUnescape(u.String())
		if err != nil {
			t.Fatalf("failure unescaping: %s\n", err)
		}
		if unescaped != input.expected {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input.expected, u.String())
		}
	}
}

func TestReplaceFragmentParam(t *testing.T) {
	p := &parsers.URIParser{}
	var inputs = []struct {
		in       string
		original string
		newKey   string
		newVal   string
		expected string
	}{
		{
			"/file?x[0]=1#x[0]=1",
			"x[0]",
			"y",
			"zoop",
			"/file?x[0]=1#y[0]=zoop",
		},
	}

	for _, input := range inputs {
		u, _ := p.Parse(input.in)
		u.ReplaceFragmentParam(input.original, input.newKey, input.newVal)
		unescaped, err := url.PathUnescape(u.String())
		if err != nil {
			t.Fatalf("failure unescaping: %s\n", err)
		}
		if unescaped != input.expected {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input.expected, u.String())
		}
	}
}

func TestReplaceParamByIndex(t *testing.T) {
	p := &parsers.URIParser{}
	var inputs = []struct {
		in       string
		original int
		newKey   string
		newVal   string
		expected string
	}{
		{
			"/file?x[0]=1",
			0,
			"y",
			"zoop",
			"/file?y[0]=zoop",
		},
		{
			"/file?x=1",
			0,
			"x",
			"blah",
			"/file?x=blah",
		},
		{
			"/file?x=1&y=2",
			1,
			"newParam",
			"newVal",
			"/file?x=1&newParam=newVal",
		},
	}

	for _, input := range inputs {
		u, _ := p.Parse(input.in)
		u.ReplaceParamByIndex(input.original, input.newKey, input.newVal)
		unescaped, err := url.PathUnescape(u.String())
		if err != nil {
			t.Fatalf("failure unescaping: %s\n", err)
		}
		if unescaped != input.expected {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input.expected, u.String())
		}
	}
}

func TestReplaceIndexedParam(t *testing.T) {
	p := &parsers.URIParser{}
	var inputs = []struct {
		in          string
		original    string
		newKey      string
		newIndexVal string
		newVal      string
		expected    string
	}{
		{
			"/file?x[0]=1",
			"x[0]",
			"y",
			"zoop",
			"zoop",
			"/file?y[zoop]=zoop",
		},
		{
			"/file?x[0]=1",
			"x[0]",
			"x",
			"zoop",
			"1",
			"/file?x[zoop]=1",
		},
		{
			"/file?x[0]=1&x[1]=2",
			"x[1]",
			"x",
			"zoop",
			"newval",
			"/file?x[0]=1&x[zoop]=newval",
		},
	}

	for _, input := range inputs {
		u, _ := p.Parse(input.in)
		u.ReplaceIndexedParam(input.original, input.newKey, input.newIndexVal, input.newVal)
		unescaped, err := url.PathUnescape(u.String())
		if err != nil {
			t.Fatalf("failure unescaping: %s\n", err)
		}
		if unescaped != input.expected {
			t.Fatalf("did not rebuild URI properly: %s != %s\n", input.expected, u.String())
		}
	}
}
