package prs

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/std"
    "gamma/token"
)

var isMainDefined bool = false

func Parse() (ast ast.Ast) {
    fmt.Println("[INFO] parsing...")

    std.Declare()

    for token.Peek().Type != token.EOF {
        ast.Decls = append(ast.Decls, prsDecl())
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }

    return
}
