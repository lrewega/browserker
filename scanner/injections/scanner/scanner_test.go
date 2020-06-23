package scanner_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/scanner"
	"gitlab.com/browserker/scanner/injections/token"
)

type expected struct {
	Pos browserk.InjectionPos
	Tok token.Token
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
				{0, token.SLASH, ""},
				{1, token.IDENT, "path1"},
				{6, token.SLASH, ""},
				{7, token.IDENT, "path2"},
				{12, token.SLASH, ""},
				{13, token.IDENT, "jlk"},
				{16, token.QUESTION, ""},
				{17, token.IDENT, "x"},
				{18, token.ASSIGN, ""},
				{19, token.IDENT, "1"},
				{20, token.EOF, ""},
			},
		},
		{
			[]byte("/123"),
			[]expected{
				{0, token.SLASH, ""},
				{1, token.IDENT, "123"},
				{4, token.EOF, ""},
			},
		},
		{
			[]byte("/"),
			[]expected{
				{0, token.SLASH, ""},
				{1, token.EOF, ""},
			},
		},
		{
			[]byte("/f_b/"),
			[]expected{
				{0, token.SLASH, ""},
				{1, token.IDENT, "f_b"},
				{4, token.SLASH, ""},
				{5, token.EOF, ""},
			},
		},
		{
			[]byte("/,f_b/"),
			[]expected{
				{0, token.SLASH, ""},
				{1, token.IDENT, ",f_b"},
				{5, token.SLASH, ""},
				{6, token.EOF, ""},
			},
		},
		{
			[]byte("#/,f_b/"),
			[]expected{
				{0, token.HASH, ""},
				{1, token.SLASH, ""},
				{2, token.IDENT, ",f_b"},
				{6, token.SLASH, ""},
				{7, token.EOF, ""},
			},
		},
		{
			[]byte("/x?a[0]=1&a[1]=2"),
			[]expected{
				{0, token.SLASH, ""},
				{1, token.IDENT, "x"},
				{2, token.QUESTION, ""},
				{3, token.IDENT, "a"},
				{4, token.LBRACK, ""},
				{5, token.IDENT, "0"},
				{6, token.RBRACK, ""},
				{7, token.ASSIGN, ""},
				{8, token.IDENT, "1"},
				{9, token.AND, ""},
				{10, token.IDENT, "a"},
				{11, token.LBRACK, ""},
				{12, token.IDENT, "1"},
				{13, token.RBRACK, ""},
				{14, token.ASSIGN, ""},
				{15, token.IDENT, "2"},
				{16, token.EOF, ""},
			},
		},
		{
			[]byte("/x?a[0]=1&a[1]=2#/load"),
			[]expected{
				{0, token.SLASH, ""},
				{1, token.IDENT, "x"},
				{2, token.QUESTION, ""},
				{3, token.IDENT, "a"},
				{4, token.LBRACK, ""},
				{5, token.IDENT, "0"},
				{6, token.RBRACK, ""},
				{7, token.ASSIGN, ""},
				{8, token.IDENT, "1"},
				{9, token.AND, ""},
				{10, token.IDENT, "a"},
				{11, token.LBRACK, ""},
				{12, token.IDENT, "1"},
				{13, token.RBRACK, ""},
				{14, token.ASSIGN, ""},
				{15, token.IDENT, "2"},
				{16, token.HASH, ""},
				{17, token.SLASH, ""},
				{18, token.IDENT, "load"},
				{22, token.EOF, ""},
			},
		},
	}

	s := scanner.New()
	for _, in := range inputs {
		s.Init(in.in, scanner.URI)
		i := 0
		for {
			pos, tok, lit := s.Scan()
			if tok == token.EOF {
				break
			}
			if pos != in.expected[i].Pos {
				t.Fatalf("[iter: %d] in: %s pos:%d did not match expected: %d\n", i, string(in.in), pos, in.expected[i].Pos)
			}
			if tok != in.expected[i].Tok {
				t.Fatalf("[iter: %d] in: %s Tok:%s did not match expected: %s\n", i, string(in.in), tok, in.expected[i].Tok)
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
				{0, token.IDENT, "x"},
				{1, token.ASSIGN, ""},
				{2, token.IDENT, "1"},
				{3, token.AND, ""},
				{4, token.IDENT, "y"},
				{5, token.ASSIGN, ""},
				{6, token.IDENT, "2"},
				{7, token.EOF, ""},
			},
		},
		{
			[]byte("<xml><hi>hello</hi></xml>"),
			[]expected{
				{0, token.LSS, ""},
				{1, token.IDENT, "xml"},
				{4, token.GTR, ""},
				{5, token.LSS, ""},
				{6, token.IDENT, "hi"},
				{8, token.GTR, ""},
				{9, token.IDENT, "hello"},
				{14, token.LSS, ""},
				{15, token.SLASH, ""},
				{16, token.IDENT, "hi"},
				{18, token.GTR, ""},
				{19, token.LSS, ""},
				{20, token.SLASH, ""},
				{21, token.IDENT, "xml"},
				{24, token.GTR, ""},
				{25, token.EOF, ""},
			},
		},
	}

	s := scanner.New()
	for _, in := range inputs {
		s.Init(in.in, scanner.Body)
		i := 0
		for {
			pos, tok, lit := s.Scan()
			if tok == token.EOF {
				break
			}
			if pos != in.expected[i].Pos {
				t.Fatalf("[iter: %d] in: %s pos:%d did not match expected: %d\n", i, string(in.in), pos, in.expected[i].Pos)
			}
			if tok != in.expected[i].Tok {
				t.Fatalf("[iter: %d] in: %s Tok:%s did not match expected: %s\n", i, string(in.in), tok, in.expected[i].Tok)
			}
			if lit != in.expected[i].Lit {
				t.Fatalf("[iter: %d] in: %s Lit:[%s] did not match expected: [%s]\n", i, string(in.in), lit, in.expected[i].Lit)
			}
			i++
		}
	}
}
