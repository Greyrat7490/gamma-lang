package main

import (
	"flag"
	"fmt"
	"gamma/asm/x86_64/nasm"
	"gamma/check"
	"gamma/gen"
	"gamma/parser"
	"gamma/token"
	"os"
	"os/exec"
)

var run bool
var showAst bool

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

    Ast := prs.Parse()
    if showAst { Ast.ShowAst() }

    check.TypeCheck(Ast)

    // TODO: optimization step
    gen.GenAsm(Ast)

    nasm.GenExe()
    if run { runExe() }
}
