package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsDecl() ast.OpDecl {
    
    switch t := token.Next(); t.Type {
    case token.Dec_var:
        d := prsDecVar()
        return &d

    case token.Def_var:
        d := prsDefVar()
        return &d

    case token.Def_fn:
        d := prsDefFn()
        return &d

    case token.Name:
        if token.Peek().Type == token.ParenL {
            fmt.Fprintln(os.Stderr, "[ERROR] function calls are not allowed in global scope")
            fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
            os.Exit(1)
        } else if token.Peek().Type == token.Assign {
            fmt.Fprintln(os.Stderr, "[ERROR] assigning variables is not allowed in global scope (but defining)")
            fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
            os.Exit(1)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not used\n", token.Peek().Str)
            fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
            os.Exit(1)
        }
        return &ast.BadDecl{}

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", t.Str)
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)

        return &ast.BadDecl{}
    }
}

func prsDefFn() ast.OpDefFn {
    name := token.Next()

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got \"%s\"(%s)\n", name.Str, name.Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if name.Str == "main" {
        isMainDefined = true
    }

    op := ast.OpDefFn{ FnName: name }
    token.Next()
    op.Args = prsDecArgs()
    token.Next()
    op.Block = prsBlock()

    return op
}

func prsDecVar() ast.OpDecVar {
    name := token.Next()
    vartype := token.Next()

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got \"%s\"(%s)\n", name.Str, name.Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
    if vartype.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Typename but got \"%s\"(%s)\n", vartype.Str, vartype.Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + vartype.At())
        os.Exit(1)
    }

    t := types.ToType(vartype.Str)
    if t == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", vartype.Str)
        fmt.Fprintln(os.Stderr, "\t" + vartype.At())
        os.Exit(1)
    }

    return ast.OpDecVar{ Varname: name, Vartype: t }
}

func prsDefVar() ast.OpDefVar {
    name := token.Last2()

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got \"%s\"(%s)\n", name.Type.Readable(), name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    token.Next()
    return ast.OpDefVar{ Varname: name, Value: prsExpr() }
}

func prsDecArgs() []ast.OpDecVar {
    decs := []ast.OpDecVar{}

    if token.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got \"%s\"(%s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    if token.Peek().Type != token.ParenR {
        decs = append(decs, prsDecVar())

        for token.Peek().Type == token.Comma {
            token.Next()
            decs = append(decs, prsDecVar())
        }
    }

    if token.Next().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got \"%s\"(%s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return decs
}
