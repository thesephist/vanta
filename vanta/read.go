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

func parse(r *reader) val {
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

func Read(s string) val {
	r := newReader(strings.TrimSpace(s))

	r.forward() // consume leading comments

	tail := list(parse(&r), null())
	prog := list(symbol([]byte("do")), tail)
	r.forward()
	for r.index < len(r.str) {
		switch r.peek() {
		case ')':
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
