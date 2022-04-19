package prs

import (
    "fmt"
    "gorec/ast"
    "gorec/token"
    "os"
)

var isMainDefined bool = false

func Parse() {
    tokensCount := len(token.GetTokens())

    for idx := 0; idx < tokensCount; idx++ {
        var decl ast.OpDecl
        decl, idx = prsDecl(idx)
        ast.AddOp(decl)
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }
}
