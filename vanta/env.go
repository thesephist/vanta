package vanta

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

func New() Environment {
	return newEnvironment(nil)
}

func (env *Environment) Eval(v val) val {
	return eval(v, env)
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
