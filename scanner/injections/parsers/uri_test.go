package parsers_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/parsers"
)

func TestQuery(t *testing.T) {
	var inputs = []struct {
		in       []byte
		expected injast.URI
	}{
		{
			[]byte("/path1/jlk?x=1"),
			injast.URI{
				Paths: []*injast.Ident{
					{
						NamePos: 1,
						Name:    "path1",
					},
				},
				File: &injast.Ident{
					NamePos: 7,
					Name:    "jlk",
				},
				Query: &injast.Query{
					Params: []*injast.KeyValueExpr{
						{
							Key:     &injast.Ident{NamePos: 11, Name: "x"},
							Sep:     12,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 13, Name: "1"},
						},
					},
				},
				Fragment: &injast.Fragment{
					Paths:  nil,
					Params: nil,
				},
			},
		},
		{
			[]byte("/path1/path2/jlk?x=1"),
			injast.URI{
				Paths: []*injast.Ident{
					{
						NamePos: 1,
						Name:    "path1",
					},
					{
						NamePos: 7,
						Name:    "path2",
					},
				},
				File: &injast.Ident{
					NamePos: 13,
					Name:    "jlk",
				},
				Query: &injast.Query{
					Params: []*injast.KeyValueExpr{
						{
							Key:     &injast.Ident{NamePos: 17, Name: "x"},
							Sep:     18,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 19, Name: "1"},
						},
					},
				},
				Fragment: &injast.Fragment{
					Paths:  nil,
					Params: nil,
				},
			},
		},
		{
			[]byte("/x?a[10]=1&a[11]=2#/load"),
			injast.URI{
				Paths: nil,
				File:  &injast.Ident{NamePos: 1, Name: "x"},
				Query: &injast.Query{
					Params: []*injast.KeyValueExpr{
						{
							Key: &injast.IndexExpr{
								X:      &injast.Ident{NamePos: 3, Name: "a"},
								Lbrack: 4,
								Index:  &injast.Ident{NamePos: 5, Name: "10"},
								Rbrack: 6,
							},
							Sep:     8,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 9, Name: "1"},
						},
						{
							Key: &injast.IndexExpr{
								X:      &injast.Ident{NamePos: 11, Name: "a"},
								Lbrack: 12,
								Index:  &injast.Ident{NamePos: 13, Name: "11"},
								Rbrack: 15,
							},
							Sep:     16,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 17, Name: "2"},
						},
					},
				},
				Fragment: &injast.Fragment{
					Paths: []*injast.Ident{
						{NamePos: 20, Name: "load"},
					},
					Params: nil,
				},
			},
		},
		{
			[]byte("/x?a[]=&"),
			injast.URI{
				Paths: nil,
				File:  &injast.Ident{NamePos: 1, Name: "x"},
				Query: &injast.Query{
					Params: []*injast.KeyValueExpr{
						{
							Key: &injast.IndexExpr{
								X:      &injast.Ident{NamePos: 3, Name: "a"},
								Lbrack: 4,
								Index:  &injast.Ident{NamePos: 5, Name: ""},
								Rbrack: 5,
							},
							Sep:     6,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 6, Name: ""},
						},
					},
				},
				Fragment: &injast.Fragment{
					Paths:  nil,
					Params: nil,
				},
			},
		},
		{
			[]byte("/x?=asdf"),
			injast.URI{
				Paths: nil,
				File:  &injast.Ident{NamePos: 1, Name: "x"},
				Query: &injast.Query{
					Params: []*injast.KeyValueExpr{
						{
							Key:     &injast.Ident{NamePos: 3, Name: ""},
							Sep:     3,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 4, Name: "asdf"},
						},
					},
				},
				Fragment: &injast.Fragment{
					Paths:  nil,
					Params: nil,
				},
			},
		},
		// fragment tests
		{
			[]byte("/#/path/file"), // fragment 'as a path'
			injast.URI{
				Paths: nil,
				File:  &injast.Ident{NamePos: 1, Name: ""},
				Query: &injast.Query{},
				Fragment: &injast.Fragment{
					Paths: []*injast.Ident{
						{
							NamePos: 3,
							Name:    "path",
						},
						{
							NamePos: 8,
							Name:    "file",
						},
					},
					Params: nil,
				},
			},
		},
		{
			[]byte("/#asdf"), // fragment 'as a path'
			injast.URI{
				Paths: nil,
				File:  nil,
				Query: &injast.Query{},
				Fragment: &injast.Fragment{
					Paths: []*injast.Ident{
						{
							NamePos: 2,
							Name:    "asdf",
						},
					},
					Params: nil,
				},
			},
		},
		{
			[]byte("/#asdf=asdf"), // fragment 'as a query param'
			injast.URI{
				Paths: nil,
				File:  nil,
				Query: &injast.Query{},
				Fragment: &injast.Fragment{
					Params: []*injast.KeyValueExpr{
						{
							Key:     &injast.Ident{NamePos: 2, Name: "asdf"},
							Sep:     6,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 7, Name: "asdf"},
						},
					},
				},
			},
		},
		{
			[]byte("/#?asdf=asdf"), // fragment 'as a query param'
			injast.URI{
				Paths: nil,
				File:  nil,
				Query: &injast.Query{},
				Fragment: &injast.Fragment{
					Params: []*injast.KeyValueExpr{
						{
							Key:     &injast.Ident{NamePos: 3, Name: "asdf"},
							Sep:     7,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 8, Name: "asdf"},
						},
					},
				},
			},
		},
		{
			[]byte("/?#asdf=asdf"), // fragment 'as a query param' (empty query)
			injast.URI{
				Paths: nil,
				File:  nil,
				Query: &injast.Query{},
				Fragment: &injast.Fragment{
					Params: []*injast.KeyValueExpr{
						{
							Key:     &injast.Ident{NamePos: 3, Name: "asdf"},
							Sep:     7,
							SepChar: '=',
							Value:   &injast.Ident{NamePos: 8, Name: "asdf"},
						},
					},
				},
			},
		},
	}

	for _, in := range inputs {
		p := &parsers.URIParser{}
		uri, err := p.Parse(string(in.in))
		if err != nil {
			t.Fatal(err)
		}

		// validate paths
		if len(uri.Paths) != len(in.expected.Paths) {
			spew.Dump(uri.Paths)
			t.Fatalf("expected paths to be equal length")
		}

		for i := 0; i < len(uri.Paths); i++ {
			exp := in.expected.Paths[i]
			res := uri.Paths[i]
			testCompareExpr(t, in.in, exp, res)
		}

		// validata params
		if len(uri.Query.Params) != len(in.expected.Query.Params) {
			t.Fatalf("expected query params (%d) to be equal length of (%d)", len(in.expected.Query.Params), len(uri.Query.Params))
		}

		for i := 0; i < len(uri.Query.Params); i++ {
			res := uri.Query.Params[i]
			exp := in.expected.Query.Params[i]
			testCompareKeyValue(t, in.in, exp, res)
		}

		// validate fragments
		if len(uri.Fragment.Paths) != len(in.expected.Fragment.Paths) {
			t.Fatalf("expected fragment paths to be equal length (%s)", string(in.in))
		}

		for i := 0; i < len(uri.Fragment.Paths); i++ {
			exp := in.expected.Fragment.Paths[i]
			res := uri.Fragment.Paths[i]
			testCompareExpr(t, in.in, exp, res)
		}
	}
}

func testCompareExpr(t *testing.T, in []byte, exp, res browserk.InjectionExpr) {
	if exp == nil && res == nil {
		return
	}

	if r, isIndex := res.(*injast.IndexExpr); isIndex {
		e, _ := exp.(*injast.IndexExpr)
		testCompareIndex(t, in, e, r)
	} else if r, isIdent := res.(*injast.Ident); isIdent {
		e, _ := exp.(*injast.Ident)
		testCompareIdent(t, in, e, r)
	}
}

func testCompareIdent(t *testing.T, in []byte, exp, res *injast.Ident) {
	if exp == nil && res == nil {
		return
	}
	if res.String() != exp.String() {
		t.Fatalf("(in: %s) val res: %s != exp: %s\n", string(in), res.String(), exp.String())
	}
	if res.Pos() != exp.Pos() {
		t.Fatalf("(in: %s) pos res: %d != exp: %d\n", string(in), res.Pos(), exp.Pos())
	}
}

func testCompareIndex(t *testing.T, in []byte, exp, res *injast.IndexExpr) {
	testCompareExpr(t, in, exp.Index, res.Index)
	testCompareExpr(t, in, exp.X, res.X)
}

func testCompareKeyValue(t *testing.T, in []byte, exp, res *injast.KeyValueExpr) {
	testCompareExpr(t, in, exp.Key, res.Key)
	if res.Sep != exp.Sep {
		t.Fatalf("(in: %s) res.Sep %q did not match expected: %q\n", string(in), res.Sep, exp.Sep)
	}
	if res.SepChar != exp.SepChar {
		t.Fatalf("(in: %s) res.SepChar %q did not match expected: %q\n", string(in), res.SepChar, exp.SepChar)
	}
	testCompareExpr(t, in, exp.Value, res.Value)
}
