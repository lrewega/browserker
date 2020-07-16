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
			[]byte("/#?asdf=asdf"),
			injast.URI{
				Fields: []browserk.InjectionExpr{
					&injast.Ident{
						Name:     "",
						Location: browserk.InjectFile,
					},
					&injast.KeyValueExpr{
						SepChar: '?',
					},
					&injast.KeyValueExpr{
						Key: &injast.Ident{
							Name:     "asdf",
							Location: browserk.InjectFragment,
						},
						SepChar: '=',
					},
					&injast.KeyValueExpr{
						Key: &injast.Ident{
							Name:     "asdf",
							Location: browserk.InjectFragment,
						},
					},
				},
			},
		},
		{
			[]byte("/x?=asdf"),
			injast.URI{
				Fields: []browserk.InjectionExpr{
					&injast.Ident{
						Name:     "x",
						Location: browserk.InjectFile,
					},
					&injast.KeyValueExpr{
						Key:     &injast.Ident{Name: "", Location: browserk.InjectQueryName},
						SepChar: '=',
						Value: &injast.Ident{
							Name:     "asdf",
							Location: browserk.InjectQueryValue,
						},
					},
				},
			},
		},
		{
			[]byte("/x?a[]=&"),
			injast.URI{
				Fields: []browserk.InjectionExpr{
					&injast.Ident{
						Name:     "x",
						Location: browserk.InjectFile,
					},
					&injast.KeyValueExpr{
						Key: &injast.IndexExpr{
							X:     &injast.Ident{Name: "a", Location: browserk.InjectQueryName},
							Index: &injast.Ident{Name: "", Location: browserk.InjectQueryIndex},
						},
						SepChar: '=',
						Value: &injast.Ident{
							Name:     "",
							Location: browserk.InjectQueryValue,
						},
					},
				},
			},
		},
		{
			[]byte("/x#load"),
			injast.URI{
				Fields: []browserk.InjectionExpr{
					&injast.Ident{
						Name:     "x",
						Location: browserk.InjectFile,
					},
					&injast.KeyValueExpr{
						Key:   &injast.Ident{Name: "load", Location: browserk.InjectQueryName},
						Value: nil,
					},
				},
			},
		},
		{
			[]byte("/x?a[10]=1#/load"),
			injast.URI{
				Fields: []browserk.InjectionExpr{
					&injast.Ident{
						Name:     "x",
						Location: browserk.InjectFile,
					},
					&injast.KeyValueExpr{
						Key: &injast.IndexExpr{
							X:     &injast.Ident{Name: "a", Location: browserk.InjectQueryName},
							Index: &injast.Ident{Name: "10", Location: browserk.InjectQueryIndex},
						},
						SepChar: '=',
						Value: &injast.Ident{
							Name:     "1",
							Location: browserk.InjectQueryValue,
						},
					},
					&injast.KeyValueExpr{
						Key:     nil,
						Value:   nil,
						SepChar: '/',
					},
					&injast.KeyValueExpr{
						Key:   &injast.Ident{Name: "load", Location: browserk.InjectQueryName},
						Value: nil,
					},
				},
			},
		},
		{
			[]byte("/path1/jlk?x=[]1"),
			injast.URI{
				Fields: []browserk.InjectionExpr{
					&injast.Ident{
						Name:     "path1",
						Location: browserk.InjectPath,
					},
					&injast.Ident{
						Name:     "jlk",
						Location: browserk.InjectFile,
					},
					&injast.KeyValueExpr{
						Key: &injast.Ident{
							Name:     "x",
							Location: browserk.InjectQueryName,
						},
						SepChar: '=',
						Value: &injast.Ident{
							Name:     "[]1",
							Location: browserk.InjectQueryValue,
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
		if len(uri.Fields) != len(in.expected.Fields) {
			spew.Config.ContinueOnMethod = true
			spew.Dump(uri.Fields)
			t.Fatalf("expected Fields to be equal length")
		}

		for i := 0; i < len(uri.Fields); i++ {
			res := uri.Fields[i]
			exp := in.expected.Fields[i]
			testCompareExpr(t, in.in, exp, res)
		}
	}
}
