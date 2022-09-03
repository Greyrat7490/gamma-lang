package main

import (
    "os"
    "os/exec"
    "fmt"
    "flag"
    "gamma/check"
    "gamma/import"
    "gamma/parser"
    "gamma/gen"
    "gamma/gen/asm/x86_64/nasm"
)

var run bool
var showAst bool
var importDir string

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
    flag.StringVar(&importDir, "I", "./std", "set import dir")

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

    imprt.SetImportDirs(path, importDir)

    Ast := prs.Parse(path)
    if showAst { Ast.ShowAst() }

    check.TypeCheck(Ast)

    // TODO: optimization step
    gen.GenAsm(Ast)

    nasm.GenExe()
    if run { runExe() }
}
