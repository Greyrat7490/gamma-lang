package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
)

var isMainDefined bool = false

func Parse() {
    fmt.Println("[INFO] parsing...")
    for token.Peek().Type != token.EOF {
        ast.AddOp(prsDecl())
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }
}
