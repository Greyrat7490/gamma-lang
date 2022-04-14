package prs

import (
    "fmt"
    "gorec/ast"
    "gorec/token"
    "os"
)

var isMainDefined bool = false

func Parse() {
    tokens := token.GetTokens()

    for i := 0; i < len(tokens); i++ {
        switch tokens[i].Type {
        case token.Dec_var:
            var op ast.OpDecVar
            op, i = prsDecVar(i)
            ast.AddOp(op)
        case token.Def_var:
            var op ast.OpDefVar
            op, i = prsDefVar(i)
            ast.AddOp(op)
        case token.Def_fn:
            var op ast.OpDefFn
            op, i = prsDefFn(i)
            ast.AddOp(op)
        case token.Name:
            if tokens[i+1].Type == token.ParenL {
                fmt.Fprintln(os.Stderr, "[ERROR] function calls are not allowed in global scope")
                fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
                os.Exit(1)
            }
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", tokens[i].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
            os.Exit(1)
        }
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }
}
