package vanta

// eval evaluates a Klisp form in a given Environment env, recursively calling
// itself where appropriate.
//
// At the moment, eval does not support proper tail
// calls. For small programs, Go's default stack is large enough, and the
// performance downsides of a tail call trampoline are not necessary.
func eval(v Val, env *Environment) Val {
	switch v.tag {
	case tcons:
		if v.car().tag == tsymbol {
			switch string(v.car().str) {
			case "quote":
				return v.cdr().car()
			case "def":
				name := v.cdr().car()
				val := eval(v.cdr().cdr().car(), env)
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
				return fn(func(args Val) Val {
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
				return macro(func(args Val) Val {
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
			panic("attempted to call a non-callable value at " + v.String())
		}
	case tsymbol:
		return env.get(string(v.str))
	default:
		return v
	}
}
