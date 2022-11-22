package prs

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/token"
    "gamma/import"
    "gamma/buildin"
)

var isMainDefined bool = false

func Parse(path string) (ast ast.Ast) {
    fmt.Println("[INFO] parsing...")

    buildin.Declare()

    tokens := imprt.ImportMain(path)

    for tokens.Peek().Type != token.EOF {
        tokens.SetLastImport()
        ast.Decls = append(ast.Decls, prsDecl(&tokens))
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }

    return
}
