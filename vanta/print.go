package vanta

import (
	"strconv"
	"strings"
)

func Print(v val) string {
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
		strings.ReplaceAll(s, "\\", "\\\\")
		strings.ReplaceAll(s, "'", "\\'")
		return s
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
	default:
		// tfn, tmacro
		return "(function)"
	}
}
