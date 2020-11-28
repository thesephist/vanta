# Vanta 🦈

*Vanta* is a minimal [Klisp](https://dotink.co/posts/klisp/) interpreter implemented in about 750 lines of pure Go. The canonical implementation of Klisp is written in [Ink](https://dotink.co) and lives at [github.com/thesephist/klisp](https://github.com/thesephist/klisp). This implementation is intended to be 100% compatible, but isn't meant for serious use (though to be honest, neither is the other one). For example, it doesn't handle errors very well (the interpreter panics). I might improve this later, but don't count on it.

Vanta has a small Go API that you can use to embed the interpreter in other Go programs. The Vanta CLI provides an interactive repl with the `-i` flag, or when run without any arguments.

The name _Vanta_ is a slight variation on _Manta_, which is the name for Joyent's ZFS-based cloud object storage service. I thought it sounded good, so flipped a letter.

Here's an example repl session in Vanta.

```clj
$ ./klisp -i ../klisp/lib/klisp.klisp
Klisp interpreter v0.1-vanta.
> (+ 1 2 3)
6
> (* (+ 1 2 3) (/ 100 5))
120
> (join (list 1 2 3) (list 'a' 'b' 'c'))
(1 2 3 a b c)
> (defn sq (x) (* x x))
(function)
> (sq 10)
100
> (sq (sq 4))
256
> (sum (nat 10))
55
> (prod (nat 10))
3628800
> (each (nat 10) println)
1
2
3
4
5
6
7
8
9
10
()
> (map (nat 10) sq)
(1 4 9 16 25 36 49 64 81 100)
>
```

This repository contains _just the interpreter_. To do anything useful, you'll probably want the [core library](https://github.com/thesephist/klisp) from the main Klisp repo as well. You can start a repl after importing libraries by passing those files to Vanta after the `-i` CLI option.

## Why?

Mostly for fun. I enjoy writing small Klisp programs to solve problems, but sometimes the [Ink-based interpreter](https://github.com/thesephist/klisp) is not as fast as I'd like. I also wanted to try implementing a Lisp in Go, and Klisp was already there for me, so I thought I'd make a much faster Klisp interpreter.

Even with the current naive implementation, Vanta is around **40-45x** faster than the Ink-based interpreter in my limited testing on the standard library tests. This probably says more about how slow Ink is than how fast Go or Vanta is.

## Go API

The full godoc is available at [godoc.org](https://godoc.org/thesephist/vanta)

// TODO

