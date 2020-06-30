package parsers_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/parsers"
)

func TestBody(t *testing.T) {
	var inputs = []struct {
		in         []byte
		expected   injast.Body
		FieldCount int
	}{
		{
			[]byte("x=1[]&y=[2"),
			injast.Body{
				Fields: []browserk.InjectionExpr{
					&injast.KeyValueExpr{
						Key:     &injast.Ident{NamePos: 0, Name: "x"},
						Sep:     1,
						SepChar: '=',
						Value:   &injast.Ident{NamePos: 2, Name: "1[]"},
					},
					&injast.KeyValueExpr{
						Key:     &injast.Ident{NamePos: 6, Name: "y"},
						Sep:     7,
						SepChar: '=',
						Value:   &injast.Ident{NamePos: 8, Name: "[2"},
					},
				},
			},
			2,
		},
		{
			[]byte("x[]=1&y=2"),
			injast.Body{
				Fields: []browserk.InjectionExpr{
					&injast.KeyValueExpr{
						Key: &injast.IndexExpr{
							X:        &injast.Ident{NamePos: 0, Name: "x"},
							Lbrack:   1,
							Index:    &injast.Ident{NamePos: 2, Name: ""},
							Rbrack:   2,
							Location: browserk.InjectBodyIndex,
						},
						Sep:     3,
						SepChar: '=',
						Value:   &injast.Ident{NamePos: 4, Name: "1"},
					},
					&injast.KeyValueExpr{
						Key:     &injast.Ident{NamePos: 6, Name: "y"},
						Sep:     7,
						SepChar: '=',
						Value:   &injast.Ident{NamePos: 8, Name: "2"},
					},
				},
			},
			2,
		},
		{
			[]byte("x=1&y=2"),
			injast.Body{
				Fields: []browserk.InjectionExpr{
					&injast.KeyValueExpr{
						Key:     &injast.Ident{NamePos: 0, Name: "x"},
						Sep:     1,
						SepChar: '=',
						Value:   &injast.Ident{NamePos: 2, Name: "1"},
					},
					&injast.KeyValueExpr{
						Key:     &injast.Ident{NamePos: 4, Name: "y"},
						Sep:     5,
						SepChar: '=',
						Value:   &injast.Ident{NamePos: 6, Name: "2"},
					},
				},
			},
			2,
		},
	}

	for _, in := range inputs {
		p := &parsers.BodyParser{}
		body, err := p.Parse(in.in)
		if err != nil {
			t.Fatal(err)
		}

		if len(body.Fields) != in.FieldCount {
			t.Fatalf("expeected %d got %d", in.FieldCount, len(body.Fields))
		}

		for i, field := range body.Fields {
			testCompareExpr(t, in.in, in.expected.Fields[i], field)
		}
	}
}
