package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/std"
)

var isMainDefined bool = false

func Parse() {
    fmt.Println("[INFO] parsing...")

    std.Declare()
    
    for token.Peek().Type != token.EOF {
        ast.AddNode(prsDecl())
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }
}
