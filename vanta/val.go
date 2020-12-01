package vanta

import (
	"bytes"
)

const (
	tnull = iota
	tbooltrue
	tboolfalse
	tnumber
	tstr
	tsymbol
	tcons
	tfn
	tmacro
)

type cons struct {
	car Val
	cdr Val
}

func (c cons) String() string {
	return Print(Val{tag: tcons, cell: &c})
}

// Val opaquely represents a valid Klisp value. It can be one of either: null
// (), boolean (true, false), number, string, symbol, list (a cons cell), a
// function, or a macro. Call val.String() to get a string representation, and
// val.Equals(val) to compare equality. Read, Eval, and Print operate on Vals.
type Val struct {
	tag int
	// number
	number float64
	// string value, mutable byte slice
	str []byte
	// symbol value, immutable string
	symb string
	// list
	cell *cons
	// fn, macro
	fn func(Val) Val
}

func (v Val) clone() Val {
	dststr := make([]byte, len(v.str), len(v.str))
	copy(dststr, v.str)

	cell := v.cell
	if cell != nil {
		cell = &cons{
			car: v.car().clone(),
			cdr: v.cdr().clone(),
		}
	}

	return Val{
		tag:    v.tag,
		number: v.number,
		str:    dststr,
		symb:   v.symb,
		cell:   cell,
		fn:     v.fn,
	}
}

// String returns an S-expression representation of the given Val.
func (v Val) String() string {
	return Print(v)
}

// Equals compares two Vals *by value*. This equality check is a deep equality
// for everything except functions and macros.
func (v Val) Equals(w Val) bool {
	if v.tag != w.tag {
		return false
	}

	switch v.tag {
	case tnull, tbooltrue, tboolfalse:
		return true
	case tnumber:
		return v.number == w.number
	case tstr:
		return bytes.Equal(v.str, w.str)
	case tsymbol:
		return v.symb == w.symb
	case tcons:
		return v.car().Equals(w.car()) && v.cdr().Equals(w.cdr())
	default:
		// tfn, tmacro
		return false
	}
}

func (v Val) car() Val {
	if v.tag == tcons {
		return v.cell.car
	}

	panic("tried to take car of " + v.String())
}

func (v Val) cdr() Val {
	if v.tag == tcons {
		return v.cell.cdr
	}

	panic("tried to take car of " + v.String())
}

func (v Val) isNull() bool {
	return v.tag == tnull
}

func (v Val) asBool() bool {
	return v.tag == tbooltrue
}

var null = Val{tag: tnull}

func boolean(v bool) Val {
	if v {
		return Val{tag: tbooltrue}
	}

	return Val{tag: tboolfalse}
}

var boolTrue = boolean(true)
var boolFalse = boolean(false)

func number(n float64) Val {
	return Val{tag: tnumber, number: n}
}

func str(s []byte) Val {
	return Val{tag: tstr, str: s}
}

func symbol(s string) Val {
	return Val{tag: tsymbol, symb: s}
}

func list(a, b Val) Val {
	return Val{tag: tcons, cell: &cons{
		car: a,
		cdr: b,
	}}
}

func fn(f func(Val) Val) Val {
	return Val{tag: tfn, fn: f}
}

func macro(f func(Val) Val) Val {
	return Val{tag: tmacro, fn: f}
}
