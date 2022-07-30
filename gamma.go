package main

import (
    "os"
    "fmt"
    "flag"
    "os/exec"
    "gamma/std"
    "gamma/ast"
    "gamma/token"
    "gamma/parser"
    "gamma/asm/x86_64/nasm"
)

var run bool
var showAst bool

func compile() {
    fmt.Println("[INFO] compiling...")

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

func runExe() {
    fmt.Printf("[EXEC] ./output\n\n")

    out, err := exec.Command("./output").CombinedOutput()
    fmt.Print(string(out))
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
    }
}

func init() {
    flag.BoolVar(&run, "r", false, "run the compiled executable")
    flag.BoolVar(&showAst, "ast", false, "show the AST")

    flag.Usage = func() {
        fmt.Println("gamma usage:")
        flag.PrintDefaults()
    }

    flag.Parse()
}

func main() {
    path := flag.Arg(0)
    if path == "" {
        fmt.Fprintln(os.Stderr, "[ERROR] you need to provide a source file to compile")
        os.Exit(1)
    }

    token.Tokenize(path)
    prs.Parse()
    // TODO: optimization step
    compile()
    nasm.GenExe()

    if showAst { ast.ShowAst() }
    if run { runExe() }
}
