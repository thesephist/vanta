package vanta

import (
	"strconv"
	"strings"
)

// Print transforms a Klisp value v back into a string, and hence is the
// inverse function of Read.
func Print(v Val) string {
	switch v.tag {
	case tnull:
		return "()"
	case tbooltrue:
		return "true"
	case tboolfalse:
		return "false"
	case tnumber:
		if i := int64(v.number); v.number == float64(i) {
			return strconv.FormatInt(i, 10)
		}
		return strconv.FormatFloat(v.number, 'f', 8, 64)
	case tstr:
		s := string(v.str)
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "'", "\\'")
		return "'" + s + "'"
	case tsymbol:
		return v.symb
	case tcons:
		term := v
		acc := []string{}
		for {
			if term.tag == tcons {
				if term.cell.cdr.tag == tcons {
					acc = append(acc, Print(term.cell.car))
					term = term.cell.cdr
				} else if term.cell.cdr.tag == tnull {
					acc = append(acc, Print(term.cell.car))
					break
				} else {
					acc = append(acc, Print(term.cell.car), ".")
					term = term.cell.cdr
				}
			} else {
				acc = append(acc, Print(term))
				break
			}
		}
		return "(" + strings.Join(acc, " ") + ")"
	default:
		// tfn, tmacro
		return "(function)"
	}
}
