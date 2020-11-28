# Vanta 🦈

**Vanta** is a minimal [Klisp](https://dotink.co/posts/klisp/) interpreter implemented in about 750 lines of pure Go. The canonical implementation of Klisp is written in [Ink](https://dotink.co) and lives at [github.com/thesephist/klisp](https://github.com/thesephist/klisp). This implementation is intended to be 100% compatible, but isn't meant for serious use (though to be honest, neither is the other one). For example, it doesn't handle errors very well (the interpreter panics). I might improve this later, but don't count on it. I was mostly just experimenting here.

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

This repository contains _just the interpreter_. To do anything useful, you'll probably want the [core library](https://github.com/thesephist/klisp/blob/main/lib/klisp.klisp) from the main Klisp repo as well. You can start a repl after importing libraries by passing those files to Vanta after the `-i` CLI option.

## Why?

Mostly for fun. I enjoy writing small Klisp programs to solve problems, but sometimes the [Ink-based interpreter](https://github.com/thesephist/klisp) is not as fast as I'd like. I also wanted to try implementing a Lisp in Go, and Klisp was already there for me, so I thought I'd make a much faster Klisp interpreter.

Even with the current naive implementation, Vanta is around **40-45x** faster than the Ink-based interpreter in my limited testing on the standard library tests. But this probably says more about how slow Ink is than how fast Go or Vanta is.

## Go API

The full godoc is available at [pkg.go.dev/github.com/thesephist/vanta](https://pkg.go.dev/github.com/thesephist/vanta/vanta). Here's a summary.

Vanta's API is modeled after the classic read/eval/print functions of Lisps. These functions operate on two types: `Val`, which opaquely represents Klisp types like linked list, number, null, and symbol; and `Environment`, which represents an execution context.

**Read** takes a string and returns a `Val` that represents the fully parsed Klisp code.

**Print** takes a `Val` and returns a string of S-expressions that represents the passed value.

To evaluate code, you should first create an environment with `vanta.New()`, which returns an `Environment`. Call the **Eval** method on the environment with a `Val` to evaluate the Val as code.

You can find a minimal example of these APIs in use in `main.go`.

At the moment, any unsuccessful attempt to parse or evaluate Klisp will panic the interpreter.

## Development

This repo uses a Makefile, and is intended to be used when the `vanta` repo is installed side-by-side with the `klisp` repository to share tests and library code. `make run` opens a repl, `make build` builds release binaries, and `make clean` removes any generated files.

Running `make run-all` (the default behavior) or `make` by itself will look for the Klisp standard library tests installed locally on your machine next to this repo, and run them.

