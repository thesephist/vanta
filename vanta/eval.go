package vanta

// eval evaluates a Klisp form in a given Environment env, recursively calling
// itself where appropriate. eval properly optimizes tail calls using a return
// trampoline through a tthunk type Val.
func eval(v Val, env *Environment, eager bool) Val {
	switch v.tag {
	case tsymbol:
		return env.get(v.symb)
	case tcons:
		if v.car().tag == tsymbol {
			switch v.car().symb {
			case "quote":
				return v.cdr().car()
			case "def":
				name := v.cdr().car()
				val := eval(v.cdr().cdr().car(), env, true)
				env.put(name.symb, val)
				return val
			case "do":
				rest := v.cdr()
				var cdr Val
				for rest.tag == tcons {
					if cdr = rest.cdr(); cdr.isNull() {
						return eval(rest.car(), env, eager)
					}

					eval(rest.car(), env, true)
					rest = cdr
				}
			case "if":
				cond := v.cdr().car()
				var body Val
				if eval(cond, env, true).tag == tbooltrue {
					body = v.cdr().cdr().car()
				} else {
					body = v.cdr().cdr().cdr().car()
				}
				return eval(body, env, eager)
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
						envc.put(param.symb, arg)

						params = params.cdr()
						args = args.cdr()
					}

					return eval(body, &envc, false)
				}, v)
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
						envc.put(param.symb, arg)

						params = params.cdr()
						args = args.cdr()
					}

					return eval(body, &envc, false)
				}, v)
			}
		}

		argcs := v.cdr()
		fn := eval(v.car(), env, true)
		if fn.tag == tfn {
			head := argcs.clone()
			rest := head
			for !rest.isNull() {
				rest.cell.car = eval(rest.car(), env, true)
				rest = rest.cdr()
			}

			if eager {
				return fn.fn(head).unwrap()
			} else {
				return thunk(fn.fn, &head)
			}
		} else if fn.tag == tmacro {
			return eval(fn.fn(argcs).unwrap(), env, eager)
		} else {
			panic("attempted to call a non-callable value at " + v.String())
		}
	default:
		return v
	}
}
