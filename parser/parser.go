package prs

import (
    "fmt"
    "os"
    "gorec/token"
)

var isMainDefined bool = false

func Parse() {
    tokens := token.GetTokens()

    for i := 0; i < len(tokens); i++ {
        switch tokens[i].Type {
        case token.Dec_var:
            i = prsDecVar(i)
        case token.Def_var:
            i = prsDefVar(i)
        case token.Def_fn:
            i = prsDefFn(i)
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
