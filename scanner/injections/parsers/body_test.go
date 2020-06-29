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
			[]byte("x=1&y=2"),
			injast.Body{
				Fields: []browserk.InjectionExpr{
					&injast.KeyValueExpr{
						Key:   &injast.Ident{NamePos: 0, Name: "x"},
						Value: &injast.Ident{NamePos: 2, Name: "1"},
					},
					&injast.KeyValueExpr{
						Key:   &injast.Ident{NamePos: 4, Name: "y"},
						Value: &injast.Ident{NamePos: 6, Name: "2"},
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
