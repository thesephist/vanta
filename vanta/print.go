package vanta

import (
	"fmt"
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
