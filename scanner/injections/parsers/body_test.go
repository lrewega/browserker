package parsers_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
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
			t.Fatalf("expected %d got %d", in.FieldCount, len(body.Fields))
		}

		for i, field := range body.Fields {
			testCompareExpr(t, in.in, in.expected.Fields[i], field)
		}
	}
}

func TestBodyJSON(t *testing.T) {
	var inputs = []struct {
		in         []byte
		expected   injast.Body
		FieldCount int
	}{
		{
			[]byte(`{"x": "one \"and\" two", "y": 2, "arr": [1,20,3.1], "n": null}`),
			injast.Body{
				Fields: []browserk.InjectionExpr{
					&injast.ObjectExpr{
						LPos:     0,
						Location: browserk.InjectJSON,
						EncChar:  '{',
						Fields: []browserk.InjectionExpr{
							&injast.KeyValueExpr{
								Key: &injast.Ident{
									NamePos:  2,
									Name:     "x",
									Mod:      "",
									Modded:   false,
									EncChar:  '"',
									Location: browserk.InjectJSONName,
								},
								Sep:     4,
								SepChar: ':',
								Value: &injast.Ident{
									NamePos:  5,
									Name:     "one \"and\" two",
									Mod:      "",
									Modded:   false,
									EncChar:  '"',
									Location: browserk.InjectJSONValue,
								},
							},
							&injast.KeyValueExpr{
								Key: &injast.Ident{
									NamePos:  26,
									Name:     "y",
									Mod:      "",
									Modded:   false,
									EncChar:  '"',
									Location: browserk.InjectJSONName,
								},
								Sep:     28,
								SepChar: ':',
								Value: &injast.Ident{
									NamePos:  5,
									Name:     "2",
									Mod:      "",
									Modded:   false,
									Location: browserk.InjectJSONValue,
								},
							},
							&injast.KeyValueExpr{
								Key: &injast.Ident{
									NamePos:  36,
									Name:     "e",
									Mod:      "",
									Modded:   false,
									EncChar:  '"',
									Location: browserk.InjectJSONName,
								},
								Sep:     38,
								SepChar: ':',
								Value: &injast.ObjectExpr{
									LPos:     0,
									Location: browserk.InjectJSON,
									EncChar:  '[',
									Fields: []browserk.InjectionExpr{
										&injast.Ident{
											NamePos:  42,
											Name:     "1",
											Location: browserk.InjectJSONValue,
										},
										&injast.Ident{
											NamePos:  42,
											Name:     "20",
											Location: browserk.InjectJSONValue,
										},
										&injast.Ident{
											NamePos:  42,
											Name:     "3.1",
											Location: browserk.InjectJSONValue,
										},
									},
								},
							},
							&injast.KeyValueExpr{
								Key: &injast.Ident{
									NamePos:  36,
									Name:     "e",
									Mod:      "",
									Modded:   false,
									EncChar:  '"',
									Location: browserk.InjectJSONName,
								},
								Sep:     38,
								SepChar: ':',
								Value: &injast.Ident{
									NamePos: 11,
									Name:    "null",
								},
							},
						},
					},
				},
			},
			1,
		},
	}

	for _, in := range inputs {
		p := &parsers.BodyParser{}
		body, err := p.Parse(in.in)
		if err != nil {
			t.Fatal(err)
		}
		spew.Config.ContinueOnMethod = true
		///spew.Dump(body.Fields)
		if len(body.Fields) != in.FieldCount {
			t.Fatalf("expected %d got %d", in.FieldCount, len(body.Fields))
		}
		t.Logf("%s\n", body.Fields[0].String())
		for i, field := range body.Fields {
			testCompareExpr(t, in.in, in.expected.Fields[i], field)
		}
	}
}
