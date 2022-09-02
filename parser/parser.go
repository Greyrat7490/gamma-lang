package prs

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/std"
    "gamma/token"
)

var isMainDefined bool = false

func Parse(path string) (ast ast.Ast) {
    fmt.Println("[INFO] parsing...")

    std.Declare()

    tokens := token.Tokenize(path)

    for tokens.Peek().Type != token.EOF {
        ast.Decls = append(ast.Decls, prsDecl(&tokens))
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }

    return
}
