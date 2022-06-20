# gore-lang

> Fun Project to write my own language, to learn Go and to see how a good language would look like for me. So it will probably end up like a modern C ... so Go

A statically and strongly typed programming language similar to Go, but with more focus on memory. It is more like a mix of C, Go and Rust.
It is named Gore, because I took everything out of C, Go and Rust I like to write this bloody mess.
And because it is written in Go...

* compiled
* statically and strongly typed
* lightweigth
* important build-in functions
* designed around hardware-near programming
* crossplatform
* fast
* easy and consistend syntax

## Supported:
* [x] Linux
* [ ] MacOS
* [ ] windows
* [x] x86_64
* [ ] ARM

## TODO:
* [x] generate assembly file
  * [x] nasm
  * [ ] fasm (preferable!)
* [x] generate executable
* [x] variables
* [x] syscalls
* [x] tests
* [x] arithmetics
  * [x] unary ops
  * [x] binary ops
    * [x] parse by precedence
  * [x] parentheses
* [ ] controll structures
  * [x] if
  * [x] else
  * [ ] elif
  * [x] while
  * [x] for
* [x] pointer
  * [x] define/assign
  * [x] deref
  * [x] get addr (via "&")
  * [x] arithmetic
* [x] turing complete -> actual programming language
  * [x] proof with Rule 110 programm
* [x] type checking
* [ ] examples
* [ ] self-hosted
* [ ] cross-platform

## Get Started

compile a source file
```console
$ go run gorec <source_file>
```
run tests
```console
$ go test ./test -v
```

later examples in examples folder...
