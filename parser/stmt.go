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
        var ifStmt ast.IfStmt
        ifStmt = prsIfStmt()

        if token.Peek().Type == token.Else {
            token.Next()
            e := prsIfElse(ifStmt)
            return &e
        }

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

    case token.Name:
        if token.Peek().Type == token.ParenL {
            c := prsCallFn()
            return &ast.OpExprStmt{ Expr: &c }
        } else if token.Peek().Type == token.Assign {
            a := prsAssignVar()
            return &a
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not used\n", token.Cur().Str)
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
            return &ast.BadStmt{}
        }

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

func prsAssignVar() ast.OpAssignVar {
    name := token.Cur()
    token.Next()
    token.Next()
    val := prsExpr()

    return ast.OpAssignVar{ Varname: name, Value: val }
}

func prsIfStmt() ast.IfStmt {
    pos := token.Cur().Pos
    token.Next()
    cond := prsExpr()
    token.Next()
    block := prsBlock()

    return ast.IfStmt{ IfPos: pos, Cond: cond, Block: block }
}

func prsIfElse(If ast.IfStmt) ast.IfElseStmt {
    pos := token.Cur().Pos
    token.Next()
    block := prsBlock()

    return ast.IfElseStmt{ If: If, ElsePos: pos, Block: block }
}

func prsWhileStmt() ast.WhileStmt {
    var op ast.WhileStmt = ast.WhileStmt{ WhilePos: token.Cur().Pos, InitVal: nil }

    if isDec() {
        op.Dec = prsDecVar()

        if token.Next().Type != token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        token.Next()
        expr := prsExpr()

        if token.Peek().Type == token.Comma {
            token.Next()
            token.Next()
            op.Cond = prsExpr()
            op.InitVal = expr
        } else {
            op.InitVal = &ast.LitExpr{ Val: token.Token{ Type: token.Number, Str: "0" }, Type: types.I32 }
            op.Cond = expr
        }
    } else {
        token.Next()
        op.Cond = prsExpr()
    }

    token.Next()
    op.Block = prsBlock()

    return op
}

func prsForStmt() ast.ForStmt {
    var op ast.ForStmt = ast.ForStmt{
        ForPos: token.Cur().Pos,
        Limit: nil,
        Start: &ast.LitExpr{
            Val: token.Token{ Str: "0", Type: token.Number },
            Type: types.I32,
        },
    }

    op.Dec = prsDecVar()

    op.Step = &ast.BinaryExpr{
        Operator: token.Token{ Type: token.Plus },
        OperandL: &ast.IdentExpr{ Ident: op.Dec.Varname },
        OperandR: &ast.LitExpr{
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: types.I32,
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
