package iterator

import (
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
)

type Visitor interface {
	Visit(inj browserk.InjectionExpr) (v Visitor)
}

func Walk(v Visitor, inj browserk.InjectionExpr) {
	if v = v.Visit(inj); v == nil {
		return
	}

	switch t := inj.(type) {
	case *injast.Ident:
		// do nothing
	case *injast.IndexExpr:
		Walk(v, t.X)
		Walk(v, t.Index)
	case *injast.KeyValueExpr:
		Walk(v, t.Key)
		Walk(v, t.Value)
	case *injast.ObjectExpr:
		if t.Fields != nil && len(t.Fields) > 0 {
			for _, expr := range t.Fields {
				Walk(v, expr)
			}
		}
	}

	v.Visit(nil)
}

type injVisitor func(browserk.InjectionExpr) bool

func (i injVisitor) Visit(inj browserk.InjectionExpr) Visitor {
	if i(inj) {
		return i
	}
	return nil
}

func Inspect(inj browserk.InjectionExpr, i func(browserk.InjectionExpr) bool) {
	Walk(injVisitor(i), inj)
}

func Collect(inj browserk.InjectionExpr) []browserk.InjectionExpr {
	expressions := make([]browserk.InjectionExpr, 0)
	Inspect(inj, func(expr browserk.InjectionExpr) bool {
		if expr == nil {
			return false
		}
		expressions = append(expressions, expr)
		return true
	})
	return expressions
}
