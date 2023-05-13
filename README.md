<div align="center">
 <img width="20%" src="https://user-images.githubusercontent.com/35865858/182031998-8febc538-375a-4663-9a71-d61e90907e39.svg">
  <h1>Gamma Programming Language</h1>
</div>

> Fun Project to write my own language, to learn Go and to see how a good language would look like for me. So it will probably end up like a modern C or Rust

A statically and strongly typed programming language similar to Go and Rust.

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

## Examples
### define var / const
```v
v := 69         // infer type from value
v u32 := 64     // explicit type

c :: -64        // just change "=" to ":" to make it a const
c i64 :: -420
```

### print
```v
// print, println, eprint, eprintln
println("hello world!")
println(fmt("print a number {} or {}", 69, "somthing else"))
eprintln("ERROR") // write to stderr
```

### vectors
```v
v := [$]i32{ len: 1, cap: 6 }
v[0] = 0
v = append::<i32>(v, 2)
v = append::<i32>(v, 3)
println(vtos(v)) // vectors are not yet supported for 'fmt'
```

### structs
```v
struct Test {
    a i32,
    b u64
}

t1 := Test{ a: 1, b: 6 }
t2 := Test{ 1, 6 }        // omit field names -> go by order (will be changed later)

println(fmt("Test{ a: {}, b: {} }", t1.a, t1.b))
```

### functions
```v
fn func(a i32, b i32) {
    println(fmt("{}", a + b))
}

fn func(a i32, b i32) -> i32 {
    ret a + b
}
```

### generic functions
```v
fn func<T>(v [$]T) -> [$]T {
    for i u64, v.len {
        v[i] = i
    }
    ret v
}

// calling generic functions
v = func::<i32>(v)
```

### interfaces
```v
// define an interface
interface Number {
    fn add(a u64, b u64) -> u64
    fn inc(self) -> u64
    fn dec(*self) -> u64
}

struct Test {
    a u64
}

// implement an interface
impl Test :: Number {
    fn add(a u64, b u64) -> u64 {
        ret a + b
    }
                // optinal: use explicitly Self type
                // optinal: use explicitly actual Struct type
    fn inc(self) -> u64 {
        ret self.a + 1
    }
    
    fn dec(*self) -> u64 {
       self.a = self.a + 1
       ret self.a
    }
}

// call interface functions
res := Test::add(30, 39)

// call like methods
t := Test{ 64 }
res := t.inc()
res2 := t.dec()    // causes a side effect due of *self

Test::inc(t)       // is possible too
Test::dec(&t)
```

### methods
```v
struct Test {
    a u64
}

// implment without an interface
impl Test {
    fn inc(self) -> u64 {
        ret self.a + 1
    }

    fn dec(*self) -> u64 {
       self.a = self.a + 1
       ret self.a
    }
}

t := Test{ 64 }
res := t.inc()
res2 := t.dec()
Test::inc(t)
Test::dec(&t)
```

### enums (like in rust)
```v
// in work
```

### const functions
```v
// only tmp sytnax for funcs (will be changed)
cfn func(a i32, b i32) -> i32 {
    // non const stmts will cause an error (like "print")
    if a > b * 2 {
        ret b
    } else {
        ret a + b
    }
}

cfn func2(a i32) -> i32 {
    // vars are allowed
    res := 0
    for i i32, a {
        res = res + i
    }
    ret res
}

a :: 30
b :: 39
c := func(a, b)    // func will be executed at compile time and "69" will be "hardcoded"
d := 69
e := func(a, d)    // if one or more args are not const the func gets executed at runtime (like a normal func)
```

### const function examples
```v
cfn fib(i u64) -> u64 {
    ret $ i == {
        0, 1: 1 as u64
           _: fib(i-1) + fib(i-2)
    }
}
```

```v
// check at compile time if your machine is little or big endian
cfn isBigEndian() -> bool {
    // even pointers (and dereferencing) are allowed (only addresses in the stack)
    word i16 := 0x0001
    ptr *bool := &word as u64 as *bool
    ret *ptr == false
}
```

### switches (syntax is still in work!!)
```v
if {
    x == 0: do_stuff()
    _: do_stuff()
}

if x == {
    0: do_stuff()
    1: do_stuff()
    _: do_stuff()
}

// cases in one-line-switches are seperated with ";" 
if x { true: do_stuff(); false: do_stuff() }
```

### xswitches (eXpression switch)
```v
// same as a normal switch but with a $ in front
res := $ if {
    x == 0: -1     // only an expression is allowed as case body (statments are allowed later too)
    _: 69
}
```

## Get Started

### compile a source file
```console
$ go run gamma <source_file>
```
### run tests
```console
$ go test ./test -v
```
### gamma usage
```console
$ go run gamma --help
gamma usage:
  -ast
    	show the AST
  -r	run the compiled executable
```
### run simple http server example
```console
$ go run gamma -r ./examples/http.gma

run on http://localhost:6969
...
```

### run rule110 example
```console
$ go run gamma -r ./test/rule110.gma

                o
               oo
              ooo
             oo o
            ooooo
           oo   o
          ooo  oo
         oo o ooo
        ooooooo o
       oo     ooo
      ooo    oo o
     oo o   ooooo
    ooooo  oo   o
   oo   o ooo  oo
  ooo  oooo o ooo
 oo o oo  ooooo o
oooooooo oo   ooo
...
```

## TODO:
* [x] generate assembly file
  * [x] nasm
  * [ ] fasm (preferable!)
* [x] variables
* [ ] functions
  * [x] define/call
  * [x] System V AMD64 ABI calling convention
  * [ ] lambda
  * [x] const function
* [ ] packages
  * [x] import
  * [x] import only once
  * [x] detected import cycles
  * [ ] pub keyword
  * [ ] access by package name
* [ ] stdlib
  * [x] sockets
  * [x] io
    * [x] print
    * [x] read/write files
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
* [x] structs
  * [x] define struct type
  * [x] define object
  * [x] access fields (read/write)
  * [x] compile time eval
* [x] turing complete -> actual programming language
  * [x] proof with Rule 110 programm
* [x] type checking
* [x] tests
* [x] examples
  * [x] simple http server
* [ ] self-hosted
* [ ] cross-platform
