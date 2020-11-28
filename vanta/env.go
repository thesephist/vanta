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
	"car": fn(func(args val) val {
		return args.car().car()
	}),
	"cdr": fn(func(args val) val {
		return args.car().cdr()
	}),
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
	// TODO: get-slice, set-slice!
	"point": fn(func(args val) val {
		str := args.car().str
		return number(float64(str[0]))
	}),
	"char": fn(func(args val) val {
		return str([]byte{byte(args.car().number)})
	}),
	"sin": fn(func(args val) val {
		return number(math.Sin(args.car().number))
	}),
	"cos": fn(func(args val) val {
		return number(math.Cos(args.car().number))
	}),
	"floor": fn(func(args val) val {
		return number(math.Trunc(args.car().number))
	}),
	"rand": fn(func(_ val) val {
		return number(rand.Float64())
	}),
	"time": fn(func(_ val) val {
		unixSeconds := float64(time.Now().UnixNano()) / 1e9
		return number(unixSeconds)
	}),
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
	"&": fn(func(args val) val {
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
	"|": fn(func(args val) val {
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
	"^": fn(func(args val) val {
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
	"string->number": fn(func(args val) val {
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
	"number->string": fn(func(args val) val {
		return str([]byte(strconv.FormatFloat(args.car().number, 'f', 8, 64)))
	}),
	"print": fn(func(args val) val {
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
			return null()
		} else {
			return env.parent.get(name)
		}
	}
}

func (env *Environment) put(name string, v val) {
	env.scope[name] = v
}
