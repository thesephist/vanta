package vanta

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func New() Environment {
	return newEnvironment(nil)
}

func (env *Environment) Eval(v val) val {
	return eval(v, env)
}

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
		return v.car().Equals(w.car()) && w.cdr().Equals(w.cdr())
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

type Environment struct {
	scope  map[string]val
	parent *Environment
}

var globalScope = map[string]val{
	"true":  boolean(true),
	"false": boolean(false),
	"car":   fn(func(args val) val { return args.car().car() }),
	"cdr":   fn(func(args val) val { return args.car().cdr() }),
	"cons": fn(func(args val) val {
		return list(args.car(), args.cdr().car())
	}),
	"len": fn(func(args val) val {
		switch args.car().tag {
		case tstr, tsymbol:
			return number(float64(len(args.car().str)))
		default:
			return number(0)
		}
	}),
	// TODO: get-slice, set-slice!, char, point, sin, cos, floor, rand, time
	"=": fn(func(args val) val {
		return boolean(args.car().Equals(args.cdr().car()))
	}),
	"<": fn(func(args val) val {
		switch args.car().tag {
		case tnumber:
			return boolean(args.car().number < args.cdr().car().number)
		case tstr:
			return boolean(bytes.Compare(args.car().str, args.cdr().car().str) == 1)
		default:
			return boolean(false)
		}
	}),
	">": fn(func(args val) val {
		switch args.car().tag {
		case tnumber:
			return boolean(args.car().number > args.cdr().car().number)
		case tstr:
			return boolean(bytes.Compare(args.car().str, args.cdr().car().str) == -1)
		default:
			return boolean(false)
		}
	}),
	"+": fn(func(args val) val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc += rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"-": fn(func(args val) val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc -= rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"*": fn(func(args val) val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc *= rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"/": fn(func(args val) val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc /= rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"#": fn(func(args val) val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc = math.Pow(acc, rest.car().number)
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"%": fn(func(args val) val {
		acc := int64(args.car().number)
		rest := args.cdr()
		for !rest.isNull() {
			acc = acc % int64(rest.car().number)
			rest = rest.cdr()
		}
		return number(float64(acc))
	}),
	// TODO: integer and byte string ops for &|^
	"&": fn(func(args val) val {
		acc := args.car().asBool()
		rest := args.cdr()
		for !rest.isNull() {
			acc = acc && rest.car().asBool()
			rest = rest.cdr()
		}
		return boolean(acc)
	}),
	"|": fn(func(args val) val {
		acc := args.car().asBool()
		rest := args.cdr()
		for !rest.isNull() {
			acc = acc || rest.car().asBool()
			rest = rest.cdr()
		}
		return boolean(acc)
	}),
	"^": fn(func(args val) val {
		acc := args.car().asBool()
		rest := args.cdr()
		for !rest.isNull() {
			acc = acc != rest.car().asBool()
			rest = rest.cdr()
		}
		return boolean(acc)
	}),
	// TODO: &, |, ^
	"type": fn(func(args val) val {
		switch args.car().tag {
		case tnull:
			return str([]byte("()"))
		case tbooltrue, tboolfalse:
			return str([]byte("boolean"))
		case tnumber:
			return str([]byte("number"))
		case tstr, tsymbol:
			return str([]byte("string"))
		case tcons:
			return str([]byte("list"))
		case tfn, tmacro:
			return str([]byte("function"))
		default:
			panic("Unknown val type:" + strconv.Itoa(args.car().tag))
		}
	}),
	// TODO: string->number, number->string
	"print": fn(func(args val) val {
		rest := args
		for !rest.isNull() {
			fmt.Printf(Print(rest.car()))
			rest = rest.cdr()
		}
		return null()
	}),
}

func newEnvironment(parent *Environment) Environment {
	if parent == nil {
		return Environment{
			scope: globalScope,
		}
	} else {
		return Environment{
			scope:  map[string]val{},
			parent: parent,
		}
	}
}

func (env *Environment) get(name string) val {
	if v, prs := env.scope[name]; prs {
		return v
	} else {
		if env.parent == nil {
			panic("Could not look up name " + name)
		} else {
			return env.parent.get(name)
		}
	}
}

func (env *Environment) put(name string, v val) {
	env.scope[name] = v
}

type reader struct {
	index int
	str   string
}

func newReader(s string) reader {
	return reader{
		index: 0,
		str:   s,
	}
}

func (r *reader) peek() byte {
	return r.str[r.index]
}

func (r *reader) next() byte {
	cur := r.peek()
	r.index += 1
	return cur
}

func (r *reader) nextSpan() []byte {
	acc := []byte{}
	for r.index < len(r.str) {
		switch cur := r.peek(); cur {
		case ' ', '\n', '\t', '(', ')':
			return acc
		default:
			r.next()
			acc = append(acc, cur)
		}
	}

	return acc
}

func (r *reader) forward() {
	for r.index < len(r.str) {
		switch cur := r.peek(); cur {
		case ' ', '\n', '\t':
			r.index += 1
		case ';':
			for r.index < len(r.str) && r.next() != '\n' {
			}
		default:
			return
		}
	}
}

func parse(r *reader) val {
	for r.index < len(r.str) {
		switch r.peek() {
		case ')':
			return null()
		case ',':
			r.next()
			r.forward()
			return list(
				symbol([]byte("quote")),
				list(parse(r), null()),
			)
		case '\'':
			r.next()
			acc := []byte{}
		parseStr:
			for {
				switch cur := r.next(); cur {
				case '\'':
					break parseStr
				case '\\':
					acc = append(acc, r.next())
				default:
					acc = append(acc, cur)
				}
			}
			return str(acc)
		case '(':
			r.next()
			r.forward()

			acc := null()
			tail := null()
		parseSexpr:
			for r.index < len(r.str) {
				switch cur := r.peek(); cur {
				case ')':
					r.next()
					break parseSexpr
				case '.':
					r.next()
					r.forward()

					cons := parse(r)
					r.forward()

					if acc.tag == tnull {
						acc = cons
					} else {
						tail.cell.cdr = cons
					}
					tail = cons
				default:
					cons := list(parse(r), null())
					r.forward()

					if acc.tag == tnull {
						acc = cons
					} else {
						tail.cell.cdr = cons
					}
					tail = cons
				}
			}
			return acc
		default:
			span := r.nextSpan()
			r.forward()
			n, err := strconv.ParseFloat(string(span), 64)
			if err != nil {
				return symbol(span)
			} else {
				return number(n)
			}
		}
	}

	return null()
}

func Read(s string) val {
	r := newReader(strings.TrimSpace(s))

	r.forward() // consume leading comments

	tail := list(parse(&r), null())
	prog := list(symbol([]byte("do")), tail)
	r.forward()
	for r.index < len(r.str) {
		switch r.peek() {
		case ')':
			return prog
		default:
			term := list(parse(&r), null())
			tail.cell.cdr = term
			r.forward()
			tail = term
		}
	}

	return prog
}

func eval(v val, env *Environment) val {
	// TODO: tail recursion
	// only tail recurse/trampoline if near recursion limit?
	// or... just unroll this into a loop
	switch v.tag {
	case tcons:
		if v.car().tag == tsymbol {
			switch string(v.car().str) {
			case "quote":
				return v.cdr().car()
			case "def":
				name := v.cdr().car()
				val := eval(v.cdr().cdr().car(), env)
				// TODO: Ugh, string(name) is at least a copy. Optimize later
				// IDEA: maybe for values originating from strings, we keep a string
				// value in the val type... idk
				env.put(string(name.str), val)
				return val
			case "do":
				rest := v.cdr()
				for rest.tag == tcons && rest.cdr().tag != tnull {
					eval(rest.car(), env)
					rest = rest.cdr()
				}
				return eval(rest.car(), env)
			case "if":
				cond := v.cdr().car()
				conseq := v.cdr().cdr().car()
				altern := v.cdr().cdr().cdr().car()
				if eval(cond, env).tag == tbooltrue {
					return eval(conseq, env)
				} else {
					return eval(altern, env)
				}
			case "fn":
				paramsTpl := v.cdr().car()
				body := v.cdr().cdr().car()
				return fn(func(args val) val {
					params := paramsTpl.clone()
					envc := newEnvironment(env)
					for !params.isNull() && !args.isNull() {
						param := params.car()
						arg := args.car()
						envc.put(string(param.str), arg)

						params = params.cdr()
						args = args.cdr()
					}

					return eval(body, &envc)
				})
			case "macro":
				paramsTpl := v.cdr().car()
				body := v.cdr().cdr().car()
				return macro(func(args val) val {
					args = list(args, null())
					params := paramsTpl.clone()
					envc := newEnvironment(env)
					for !params.isNull() && !args.isNull() {
						param := params.car()
						arg := args.car()
						envc.put(string(param.str), arg)

						params = params.cdr()
						args = args.cdr()
					}

					return eval(body, &envc)
				})
			}
		}

		argcs := v.cdr()
		fn := eval(v.car(), env)
		if fn.tag == tfn {
			head := argcs.clone()
			rest := head
			for !rest.isNull() {
				rest.cell.car = eval(rest.car(), env)
				rest = rest.cdr()
			}
			return fn.fn(head)
		} else if fn.tag == tmacro {
			t := fn.fn(argcs)
			return eval(t, env)
		} else {
			panic("attempted to call a non-function at " + v.String())
		}
	case tsymbol:
		return env.get(string(v.str))
	default:
		return v
	}
}

func Print(v val) string {
	switch v.tag {
	case tnull:
		return "()"
	case tbooltrue:
		return "true"
	case tboolfalse:
		return "false"
	case tnumber:
		return fmt.Sprintf("%f", v.number)
	case tstr:
		// TODO: use Klisp logic here instead
		return strconv.Quote(string(v.str))
	case tsymbol:
		return string(v.str)
	case tcons:
		term := v
		acc := []string{}
		for {
			if term.tag == tcons && term.cell.cdr.tag == tcons {
				acc = append(acc, Print(term.cell.car))
				term = term.cell.cdr
			} else if term.tag == tcons && term.cell.cdr.tag == tnull {
				acc = append(acc, Print(term.cell.car))
				break
			} else if term.tag == tcons {
				acc = append(acc, Print(term.cell.car), ".")
				term = term.cell.cdr
			} else {
				acc = append(acc, Print(term))
				break
			}
		}
		return "(" + strings.Join(acc, " ") + ")"
	case tfn, tmacro:
		return "(function)"
	}

	panic("unreachable.")
}
