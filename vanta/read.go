package vanta

import (
	"strconv"
	"strings"
)

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

// parse contains most of the S-expression parsing logic in Read, and calls
// itself recursively to parse one (1) S-expression.
func parse(r *reader) Val {
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

// Read parses a string input into a (do...) S-expression containing all parsed
// S-expressions in the input, and therefore is (loosely) the inverse of Print.
func Read(s string) Val {
	r := newReader(strings.TrimSpace(s))

	r.forward() // consume leading comments

	tail := list(parse(&r), null())
	prog := list(symbol([]byte("do")), tail)
	r.forward()
	for r.index < len(r.str) {
		switch r.peek() {
		case ')':
			// Without this guard, an extra right paren causes an infinite loop
			// in the reader as parse() will immediately return null.
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
