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
				var cdr Val
				for rest.tag == tcons {
					if cdr = rest.cdr(); cdr.isNull() {
						return eval(rest.car(), env)
					}

					eval(rest.car(), env)
					rest = cdr
				}
			case "if":
				cond := v.cdr().car()
				var body Val
				if eval(cond, env).tag == tbooltrue {
					body = v.cdr().cdr().car()
				} else {
					body = v.cdr().cdr().cdr().car()
				}
				return eval(body, env)
			case "fn":
				paramsTpl := v.cdr().car()
				body := v.cdr().cdr().car()
				return fn(func(args Val) Val {
					params := paramsTpl
					envc := newEnvironment(env)
					var param, arg Val
					for !params.isNull() && !args.isNull() {
						param = params.car()
						arg = args.car()
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
					args = list(args, null)
					params := paramsTpl
					envc := newEnvironment(env)
					var param, arg Val
					for !params.isNull() && !args.isNull() {
						param = params.car()
						arg = args.car()
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
			return eval(fn.fn(argcs), env)
		} else {
			panic("attempted to call a non-callable value at " + v.String())
		}
	case tsymbol:
		return env.get(string(v.str))
	default:
		return v
	}
}
