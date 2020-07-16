package parsers_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
)

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
	} else if r, isKeyValue := res.(*injast.KeyValueExpr); isKeyValue {
		e, _ := exp.(*injast.KeyValueExpr)
		testCompareKeyValue(t, in, e, r)
	} else if r, isObject := res.(*injast.ObjectExpr); isObject {
		e, _ := exp.(*injast.ObjectExpr)
		testCompareObject(t, in, e, r)
	}
}

func testCompareObject(t *testing.T, in []byte, exp, res *injast.ObjectExpr) {
	if exp == nil && res == nil {
		return
	}
	if len(res.Fields) != len(exp.Fields) {
		t.Fatalf("(in: %s) val res: %s != exp: %s\n", string(in), res.String(), exp.String())
	}

	for i := 0; i < len(exp.Fields); i++ {
		testCompareExpr(t, in, exp.Fields[i], res.Fields[i])
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
