package scanner_test

import (
	"testing"

	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/scanner"
)

type expected struct {
	Pos injast.Pos
	Tok injast.Token
	Lit string
}

func TestScan(t *testing.T) {

	var inputs = []struct {
		in       []byte
		expected []expected
	}{
		{
			[]byte("/path1/path2/jlk?x=1"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.IDENT, "path1"},
				{6, injast.SLASH, ""},
				{7, injast.IDENT, "path2"},
				{12, injast.SLASH, ""},
				{13, injast.IDENT, "jlk"},
				{16, injast.QUESTION, ""},
				{17, injast.IDENT, "x"},
				{18, injast.ASSIGN, ""},
				{19, injast.IDENT, "1"},
				{20, injast.EOF, ""},
			},
		},
		{
			[]byte("/123"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.IDENT, "123"},
				{4, injast.EOF, ""},
			},
		},
		{
			[]byte("/"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.EOF, ""},
			},
		},
		{
			[]byte("/f_b/"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.IDENT, "f_b"},
				{4, injast.SLASH, ""},
				{5, injast.EOF, ""},
			},
		},
		{
			[]byte("/,f_b/"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.IDENT, ",f_b"},
				{5, injast.SLASH, ""},
				{6, injast.EOF, ""},
			},
		},
	}
	s := scanner.New()
	for _, in := range inputs {
		s.Init(in.in, scanner.Path)
		i := 0
		for {
			pos, tok, lit := s.Scan()
			if tok == injast.EOF {
				break
			}
			if pos != in.expected[i].Pos {
				t.Fatalf("[iter: %d] in: %s pos:%d did not match expected: %d\n", i, string(in.in), pos, in.expected[i].Pos)
			}
			if tok != in.expected[i].Tok {
				t.Fatalf("[iter: %d] in: %s Tok:%d did not match expected: %d\n", i, string(in.in), tok, in.expected[i].Tok)
			}
			if lit != in.expected[i].Lit {
				t.Fatalf("[iter: %d] in: %s Lit:%s did not match expected: %s\n", i, string(in.in), lit, in.expected[i].Lit)
			}
			i++
		}
	}
}
