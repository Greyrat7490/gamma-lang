<div align="center">
 <img width="20%" src="https://user-images.githubusercontent.com/35865858/182031998-8febc538-375a-4663-9a71-d61e90907e39.svg">
  <h1>Gamma Programming Language</h1>
</div>

> Fun Project to write my own language, to learn Go and to see how a good language would look like for me. So it will probably end up like a modern C

A statically and strongly typed programming language similar to Go, but with more focus on memory. It is more like a mix of C, Go and Rust.

* fast
* easy
* compiled
* statically and strongly typed
* lightweigth
* important build-in functions
* designed around hardware-near programming
* crossplatform

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
* [x] variables
* [x] functions
* [x] io
  * [x] print
* [x] arithmetics
  * [x] unary ops
  * [x] binary ops
    * [x] parse by precedence
  * [x] parentheses
* [x] controll structures
  * [x] if, else, elif
  * [x] while, for
  * [x] switch
  * [x] xswitch (expr switch)
* [x] pointer
  * [x] define/assign
  * [x] deref
  * [x] get addr (via "&")
  * [x] arithmetic
* [x] consts
  * [x] define/use
  * [x] compile time eval
* [x] arrays
  * [x] define/use
  * [x] multi-dimensionale
  * [x] compile time eval
* [ ] structs
* [x] turing complete -> actual programming language
  * [x] proof with Rule 110 programm
* [x] type checking
* [x] tests
* [ ] examples
* [ ] self-hosted
* [ ] cross-platform

## Get Started

compile a source file
```console
$ go run gamma <source_file>
```
run tests
```console
$ go test ./test -v
```
gamma usage
```console
$ go run gamma --help
gamma usage:
  -ast
    	show the AST
  -r	run the compiled executable
```
