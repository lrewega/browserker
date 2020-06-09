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
		{
			[]byte("#/,f_b/"),
			[]expected{
				{0, injast.HASH, ""},
				{1, injast.SLASH, ""},
				{2, injast.IDENT, ",f_b"},
				{6, injast.SLASH, ""},
				{7, injast.EOF, ""},
			},
		},
		{
			[]byte("/x?a[0]=1&a[1]=2"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.IDENT, "x"},
				{2, injast.QUESTION, ""},
				{3, injast.IDENT, "a"},
				{4, injast.LBRACK, ""},
				{5, injast.IDENT, "0"},
				{6, injast.RBRACK, ""},
				{7, injast.ASSIGN, ""},
				{8, injast.IDENT, "1"},
				{9, injast.AND, ""},
				{10, injast.IDENT, "a"},
				{11, injast.LBRACK, ""},
				{12, injast.IDENT, "1"},
				{13, injast.RBRACK, ""},
				{14, injast.ASSIGN, ""},
				{15, injast.IDENT, "2"},
				{16, injast.EOF, ""},
			},
		},
		{
			[]byte("/x?a[0]=1&a[1]=2#/load"),
			[]expected{
				{0, injast.SLASH, ""},
				{1, injast.IDENT, "x"},
				{2, injast.QUESTION, ""},
				{3, injast.IDENT, "a"},
				{4, injast.LBRACK, ""},
				{5, injast.IDENT, "0"},
				{6, injast.RBRACK, ""},
				{7, injast.ASSIGN, ""},
				{8, injast.IDENT, "1"},
				{9, injast.AND, ""},
				{10, injast.IDENT, "a"},
				{11, injast.LBRACK, ""},
				{12, injast.IDENT, "1"},
				{13, injast.RBRACK, ""},
				{14, injast.ASSIGN, ""},
				{15, injast.IDENT, "2"},
				{16, injast.HASH, ""},
				{17, injast.SLASH, ""},
				{18, injast.IDENT, "load"},
				{22, injast.EOF, ""},
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
				t.Fatalf("[iter: %d] in: %s Lit:[%s] did not match expected: %s\n", i, string(in.in), lit, in.expected[i].Lit)
			}
			i++
		}
	}
}

func TestScanBody(t *testing.T) {
	var inputs = []struct {
		in       []byte
		expected []expected
	}{
		{
			[]byte("x=1&y=2"),
			[]expected{
				{0, injast.IDENT, "x"},
				{1, injast.ASSIGN, ""},
				{2, injast.IDENT, "1"},
				{3, injast.AND, ""},
				{4, injast.IDENT, "y"},
				{5, injast.ASSIGN, ""},
				{6, injast.IDENT, "2"},
				{7, injast.EOF, ""},
			},
		},
	}

	s := scanner.New()
	for _, in := range inputs {
		s.Init(in.in, scanner.Body)
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
				t.Fatalf("[iter: %d] in: %s Lit:[%s] did not match expected: %s\n", i, string(in.in), lit, in.expected[i].Lit)
			}
			i++
		}
	}
}
