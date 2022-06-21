package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsStmt() ast.OpStmt {
    switch t := token.Next(); t.Type {
    case token.Dec_var:
        d := prsDecVar()
        return &ast.OpDeclStmt{ Decl: &d }

    case token.Def_var:
        d := prsDefVar()
        return &ast.OpDeclStmt{ Decl: &d }

    case token.If:
        ifStmt := prsIfStmt()
        return &ifStmt

    case token.While:
        w := prsWhileStmt()
        return &w

    case token.For:
        f := prsForStmt()
        return &f

    case token.Break:
        b := prsBreak()
        return &b

    case token.Continue:
        c := prsContinue()
        return &c

    case token.Mul:
        expr := prsUnaryExpr()

        if token.Peek().Type != token.Assign {
            fmt.Fprintf(os.Stderr, "[ERROR] expected an assignment after dereferencing a pointer variable but got \"%s\"\n", token.Peek().Str)
            fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
            os.Exit(1)
            return &ast.BadStmt{}
        }

        a := prsAssignVar(expr)
        return &a

    case token.Name:
        if token.Peek().Type == token.ParenL {
            c := prsCallFn()
            return &ast.OpExprStmt{ Expr: &c }
        } else if token.Peek().Type == token.Assign {
            name := prsIdentExpr()
            a := prsAssignVar(name)
            return &a
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not used\n", token.Cur().Str)
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
            return &ast.BadStmt{}
        }


    case token.Elif:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (elif without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Else:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (else without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Assign:
        fmt.Fprintf(os.Stderr, "[ERROR] no destination for assignment\n")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Def_fn:
        fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token \"%s\" (of type %s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}
    }
}

func prsBlock() ast.OpBlock {
    block := ast.OpBlock{ BraceLPos: token.Cur().Pos }

    for token.Peek().Type != token.BraceR {
        block.Stmts = append(block.Stmts, prsStmt())
    }

    block.BraceRPos = token.Next().Pos

    return block
}

func prsAssignVar(dest ast.OpExpr) ast.OpAssignVar {
    pos := token.Next().Pos
    token.Next()
    val := prsExpr()

    return ast.OpAssignVar{ Pos: pos, Dest: dest, Value: val }
}

func prsIfStmt() ast.IfStmt {
    pos := token.Cur().Pos
    token.Next()
    cond := prsExpr()
    token.Next()
    block := prsBlock()

    ifStmt := ast.IfStmt{ Pos: pos, Cond: cond, Block: block }

    if token.Peek().Type == token.Else {
        token.Next()
        elseStmt := prsElse()
        ifStmt.Else = &elseStmt
    } else if token.Peek().Type == token.Elif {
        token.Next()
        elifStmt := prsElif()
        ifStmt.Elif = &elifStmt
    }

    return ifStmt
}

func prsElif() ast.ElifStmt {
    return ast.ElifStmt(prsIfStmt())
}

func prsElse() ast.ElseStmt {
    pos := token.Cur().Pos
    token.Next()
    block := prsBlock()

    return ast.ElseStmt{ ElsePos: pos, Block: block }
}

func prsWhileStmt() ast.WhileStmt {
    var op ast.WhileStmt = ast.WhileStmt{ WhilePos: token.Cur().Pos, InitVal: nil }

    if token.Peek().Type == token.Name && token.Peek2().Type == token.Typename {
        op.Dec = prsDecVar()

        if token.Next().Type != token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        token.Next()
        expr := prsExpr()

        if token.Next().Type == token.Comma {
            token.Next()
            op.Cond = prsExpr()
            op.InitVal = expr
            token.Next()
        } else {
            op.InitVal = &ast.LitExpr{ Val: token.Token{ Type: token.Number, Str: "0" }, Type: types.I32Type{} }
            op.Cond = expr
        }
    } else {
        token.Next()
        op.Cond = prsExpr()
        token.Next()
    }

    op.Block = prsBlock()

    return op
}

func prsForStmt() ast.ForStmt {
    var op ast.ForStmt = ast.ForStmt{
        ForPos: token.Cur().Pos,
        Limit: nil,
        Start: &ast.LitExpr{
            Val: token.Token{ Str: "0", Type: token.Number },
            Type: types.I32Type{},
        },
    }

    op.Dec = prsDecVar()

    op.Step = &ast.BinaryExpr{
        Operator: token.Token{ Type: token.Plus },
        OperandL: &ast.IdentExpr{ Ident: op.Dec.Varname },
        OperandR: &ast.LitExpr{
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: types.I32Type{},
        },
    }

    if token.Next().Type == token.Comma {
        token.Next()
        op.Limit = prsExpr()

        if token.Next().Type == token.Comma {
            token.Next()
            op.Start = prsExpr()

            if token.Next().Type == token.Comma {
                token.Next()
                op.Step = prsExpr()
                token.Next()
            }
        }
    }

    op.Block = prsBlock()

    return op
}

func prsBreak() ast.BreakStmt {
    var op ast.BreakStmt = ast.BreakStmt{ Pos: token.Cur().Pos }
    return op
}

func prsContinue() ast.ContinueStmt {
    var op ast.ContinueStmt = ast.ContinueStmt{ Pos: token.Cur().Pos }
    return op
}
