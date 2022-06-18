package main

import (
    "os"
    "fmt"
    "io/ioutil"
    "gorec/std"
    "gorec/ast"
    "gorec/token"
    "gorec/parser"
    "gorec/asm/x86_64/nasm"
)

func compile() {
    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    defer asm.Close()

    nasm.Header(asm)

    std.Define(asm)

    ast.Compile(asm)

    nasm.Footer(asm)
}


func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] you need to provide a source file to compile")
        os.Exit(1)
    }

    src, err := ioutil.ReadFile(os.Args[1])
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
        os.Exit(1)
    }

    token.Tokenize(src)
    prs.Parse()
    // TODO: optimization step
    ast.TypeCheck()
    ast.ShowAst()
    compile()
    nasm.GenExe()
}
