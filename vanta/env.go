package vanta

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// New returns a new top-level environment in which Klisp programs can be
// evaluated. Because Klisp doesn't have the traditional notion of a "VM",
// initialize a new Environment to evaluate a form.
func New() Environment {
	return newEnvironment(nil)
}

// Eval evaluates Klisp forms passed as Vals in the Environment Env.
func (env *Environment) Eval(v Val) Val {
	return eval(v, env, true)
}

// Environment represents a lexical scope in which Klisp forms are evaluated.
// It contains a lexical scope that contains bound variable names.
type Environment struct {
	scope  map[string]Val
	parent *Environment
}

// globalScope contains all built-in functions and values in Klisp
var globalScope = map[string]Val{
	"true":  boolTrue,
	"false": boolFalse,
	"car": nativeFn(func(args Val) Val {
		return args.car().car()
	}),
	"cdr": nativeFn(func(args Val) Val {
		return args.car().cdr()
	}),
	"cons": nativeFn(func(args Val) Val {
		return list(args.car(), args.cdr().car())
	}),
	"len": nativeFn(func(args Val) Val {
		switch args.car().tag {
		case tstr:
			return number(float64(len(args.car().str)))
		case tsymbol:
			return number(float64(len(args.car().symb)))
		default:
			return number(0)
		}
	}),
	"gets": nativeFn(func(args Val) Val {
		s := args.car()
		start := int(args.cdr().car().number)
		end := int(args.cdr().cdr().car().number)

		if start < 0 {
			start = 0
		}
		if end > len(s.str) {
			end = len(s.str)
		}

		return str(s.str[start:end])
	}),
	"sets!": nativeFn(func(args Val) Val {
		s := args.car()
		start := int(args.cdr().car().number)
		substr := args.cdr().cdr().car().str
		end := start + len(substr)

		if start < 0 {
			start = 0
		}
		if end > len(s.str) {
			end = len(s.str)
		}

		copy(s.str[start:end], substr[0:end-start])
		return s
	}),
	"point": nativeFn(func(args Val) Val {
		str := args.car().str
		return number(float64(str[0]))
	}),
	"char": nativeFn(func(args Val) Val {
		return str([]byte{byte(args.car().number)})
	}),
	"sin": nativeFn(func(args Val) Val {
		return number(math.Sin(args.car().number))
	}),
	"cos": nativeFn(func(args Val) Val {
		return number(math.Cos(args.car().number))
	}),
	"floor": nativeFn(func(args Val) Val {
		return number(math.Trunc(args.car().number))
	}),
	"rand": nativeFn(func(_ Val) Val {
		return number(rand.Float64())
	}),
	"time": nativeFn(func(_ Val) Val {
		unixSeconds := float64(time.Now().UnixNano()) / 1e9
		return number(unixSeconds)
	}),
	"=": nativeFn(func(args Val) Val {
		return boolean(args.car().Equals(args.cdr().car()))
	}),
	"<": nativeFn(func(args Val) Val {
		switch args.car().tag {
		case tnumber:
			return boolean(args.car().number < args.cdr().car().number)
		case tstr:
			return boolean(bytes.Compare(args.car().str, args.cdr().car().str) == 1)
		default:
			return boolFalse
		}
	}),
	">": nativeFn(func(args Val) Val {
		switch args.car().tag {
		case tnumber:
			return boolean(args.car().number > args.cdr().car().number)
		case tstr:
			return boolean(bytes.Compare(args.car().str, args.cdr().car().str) == -1)
		default:
			return boolFalse
		}
	}),
	"+": nativeFn(func(args Val) Val {
		rest := args.cdr()
		if args.car().tag == tnumber {
			acc := args.car().number
			for !rest.isNull() {
				acc += rest.car().number
				rest = rest.cdr()
			}
			return number(acc)
		}
		acc := args.car().str
		for !rest.isNull() {
			acc = append(acc, rest.car().str...)
			rest = rest.cdr()
		}
		return str(acc)
	}),
	"-": nativeFn(func(args Val) Val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc -= rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"*": nativeFn(func(args Val) Val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc *= rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"/": nativeFn(func(args Val) Val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc /= rest.car().number
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"#": nativeFn(func(args Val) Val {
		acc := args.car().number
		rest := args.cdr()
		for !rest.isNull() {
			acc = math.Pow(acc, rest.car().number)
			rest = rest.cdr()
		}
		return number(acc)
	}),
	"%": nativeFn(func(args Val) Val {
		acc := int64(args.car().number)
		rest := args.cdr()
		for !rest.isNull() {
			acc = acc % int64(rest.car().number)
			rest = rest.cdr()
		}
		return number(float64(acc))
	}),
	"&": nativeFn(func(args Val) Val {
		rest := args.cdr()
		if args.car().tag == tnumber {
			acc := int64(args.car().number)
			for !rest.isNull() {
				acc = acc & int64(rest.car().number)
				rest = rest.cdr()
			}
			return number(float64(acc))
		}
		acc := args.car().asBool()
		for !rest.isNull() {
			acc = acc && rest.car().asBool()
			rest = rest.cdr()
		}
		return boolean(acc)
	}),
	"|": nativeFn(func(args Val) Val {
		rest := args.cdr()
		if args.car().tag == tnumber {
			acc := int64(args.car().number)
			for !rest.isNull() {
				acc = acc | int64(rest.car().number)
				rest = rest.cdr()
			}
			return number(float64(acc))
		}
		acc := args.car().asBool()
		for !rest.isNull() {
			acc = acc || rest.car().asBool()
			rest = rest.cdr()
		}
		return boolean(acc)
	}),
	"^": nativeFn(func(args Val) Val {
		rest := args.cdr()
		if args.car().tag == tnumber {
			acc := int64(args.car().number)
			for !rest.isNull() {
				acc = acc ^ int64(rest.car().number)
				rest = rest.cdr()
			}
			return number(float64(acc))
		}
		acc := args.car().asBool()
		for !rest.isNull() {
			acc = acc != rest.car().asBool()
			rest = rest.cdr()
		}
		return boolean(acc)
	}),
	"type": nativeFn(func(args Val) Val {
		switch args.car().tag {
		case tnull:
			return str([]byte("()"))
		case tbooltrue, tboolfalse:
			return str([]byte("boolean"))
		case tnumber:
			return str([]byte("number"))
		case tstr:
			return str([]byte("string"))
		case tsymbol:
			return str([]byte("symbol"))
		case tcons:
			return str([]byte("list"))
		case tfn, tmacro:
			return str([]byte("function"))
		default:
			panic("Unknown Val type:" + strconv.Itoa(int(args.car().tag)))
		}
	}),
	"string->number": nativeFn(func(args Val) Val {
		operand := args.car()
		if operand.tag == tstr {
			n, err := strconv.ParseFloat(string(operand.str), 64)
			if err != nil {
				return number(0)
			} else {
				return number(n)
			}
		} else {
			return number(0)
		}
	}),
	"number->string": nativeFn(func(args Val) Val {
		v := args.car()
		if i := int64(v.number); v.number == float64(i) {
			return str([]byte(strconv.FormatInt(i, 10)))
		}
		return str([]byte(strconv.FormatFloat(v.number, 'f', 8, 64)))
	}),
	"string->symbol": nativeFn(func(args Val) Val {
		operand := args.car()
		if operand.tag == tstr {
			return symbol(string(operand.str))
		} else {
			panic("string->symbol on non-string: " + Print(operand))
		}
	}),
	"symbol->string": nativeFn(func(args Val) Val {
		operand := args.car()
		if operand.tag == tsymbol {
			return str([]byte(operand.symb))
		} else {
			panic("symbol->string on non-symbol: " + Print(operand))
		}
	}),
	"print": nativeFn(func(args Val) Val {
		rest := args
		for {
			cur := rest.car()
			if cur.tag == tstr || cur.tag == tsymbol {
				os.Stdout.Write(cur.str)
			} else {
				fmt.Printf(Print(cur))
			}
			if !rest.cdr().isNull() {
				fmt.Printf(" ")
			} else {
				break
			}

			rest = rest.cdr()
		}
		return null
	}),
}

// Create a (local) environment, optionally with a parent environment/scope provided.
func newEnvironment(parent *Environment) Environment {
	if parent == nil {
		return Environment{
			scope: globalScope,
		}
	}

	return Environment{
		scope:  map[string]Val{},
		parent: parent,
	}
}

func (env *Environment) get(name string) Val {
	if v, prs := env.scope[name]; prs {
		return v
	}

	if env.parent == nil {
		return null
	}

	return env.parent.get(name)
}

func (env *Environment) put(name string, v Val) {
	env.scope[name] = v
}
