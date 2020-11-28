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
	car val
	cdr val
}

func (c cons) String() string {
	return Print(val{tag: tcons, cell: &c})
}

type val struct {
	tag int
	// number
	number float64
	// str, symbol
	str []byte
	// list
	cell *cons
	// fn, macro
	fn func(val) val
}

func (v val) clone() val {
	dststr := make([]byte, len(v.str))
	copy(dststr, v.str)

	cell := v.cell
	if cell != nil {
		cell = &cons{
			car: v.car().clone(),
			cdr: v.cdr().clone(),
		}
	}

	return val{
		tag:    v.tag,
		number: v.number,
		str:    dststr,
		cell:   cell,
		fn:     v.fn,
	}
}

func (v val) String() string {
	return Print(v)
}

func (v val) Equals(w val) bool {
	if v.tag != w.tag {
		return false
	}

	switch v.tag {
	case tnull, tbooltrue, tboolfalse:
		return true
	case tnumber:
		return v.number == w.number
	case tstr, tsymbol:
		return bytes.Equal(v.str, w.str)
	case tcons:
		return v.car().Equals(w.car()) && v.cdr().Equals(w.cdr())
	case tfn, tmacro:
		return false
	default:
		panic("unreachable.")
	}
}

func (v val) car() val {
	if v.tag == tcons {
		return v.cell.car
	} else {
		panic("tried to take car of " + v.String())
	}
}

func (v val) cdr() val {
	if v.tag == tcons {
		return v.cell.cdr
	} else {
		panic("tried to take car of " + v.String())
	}
}

func (v val) isNull() bool {
	return v.tag == tnull
}

func (v val) asBool() bool {
	return v.tag == tbooltrue
}

func null() val {
	return val{tag: tnull}
}

func boolean(v bool) val {
	if v {
		return val{tag: tbooltrue}
	} else {
		return val{tag: tboolfalse}
	}
}

func number(n float64) val {
	return val{tag: tnumber, number: n}
}

func str(s []byte) val {
	return val{tag: tstr, str: s}
}

func symbol(s []byte) val {
	return val{tag: tsymbol, str: s}
}

func list(a, b val) val {
	return val{tag: tcons, cell: &cons{
		car: a,
		cdr: b,
	}}
}

func fn(f func(val) val) val {
	return val{tag: tfn, fn: f}
}

func macro(f func(val) val) val {
	return val{tag: tmacro, fn: f}
}
