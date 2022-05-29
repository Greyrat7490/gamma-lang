package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsStmt(idx int) (ast.OpStmt, int) {
    tokens := token.GetTokens()

    switch tokens[idx].Type {
    case token.Dec_var:
        var decOp ast.OpDecVar
        decOp, idx = prsDecVar(idx)
        return &ast.OpDeclStmt{ Decl: &decOp }, idx

    case token.Def_var:
        var defOp ast.OpDefVar
        defOp, idx = prsDefVar(idx)
        return &ast.OpDeclStmt{ Decl: &defOp }, idx

    case token.If:
        var ifStmt ast.IfStmt
        ifStmt, idx = prsIfStmt(idx)

        if tokens[idx+1].Type == token.Else {
            var ifElse ast.IfElseStmt
            ifElse, idx = prsIfElse(ifStmt, idx+1)
            return &ifElse, idx
        }

        return &ifStmt, idx

    case token.While:
        var whileStmt ast.WhileStmt
        whileStmt, idx = prsWhileStmt(idx)
        return &whileStmt, idx

    case token.For:
        var forStmt ast.ForStmt
        forStmt, idx = prsForStmt(idx)
        return &forStmt, idx

    case token.Break:
        var breakStmt ast.BreakStmt
        breakStmt, idx = prsBreak(idx)
        return &breakStmt, idx

    case token.Continue:
        var continueStmt ast.ContinueStmt
        continueStmt, idx = prsContinue(idx)
        return &continueStmt, idx

    case token.Name:
        if tokens[idx+1].Type == token.ParenL {
            var callOp ast.OpFnCall
            callOp, idx = prsCallFn(idx)
            return &ast.OpExprStmt{ Expr: &callOp }, idx
        } else if tokens[idx+1].Type == token.Assign {
            var o ast.OpAssignVar
            o, idx = prsAssignVar(idx+1)
            return &o, idx
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not used\n", tokens[idx].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
            return &ast.BadStmt{}, -1
        }

    case token.Else:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (else without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1

    case token.Assign:
        fmt.Fprintf(os.Stderr, "[ERROR] no destination for assignment\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1

    case token.Def_fn:
        fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token \"%s\" (of type \"%s\")\n", tokens[idx].Str, tokens[idx].Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1
    }
}

func prsBlock(idx int) (ast.OpBlock, int) {
    tokens := token.GetTokens()
    block := ast.OpBlock{ BraceLPos: tokens[idx].Pos }
    idx++

    for ; idx < len(tokens); idx++ {
        if tokens[idx].Type == token.BraceR {
            block.BraceRPos = tokens[idx].Pos
            return block, idx
        }

        var stmt ast.OpStmt
        stmt, idx = prsStmt(idx)
        block.Stmts = append(block.Stmts, stmt)
    }

    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" was not closed (missing \"}\")\n", tokens[idx+1].Str)
    fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
    os.Exit(1)

    return ast.OpBlock{}, -1
}

func prsAssignVar(idx int) (ast.OpAssignVar, int) {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    v := tokens[idx-1]
    value, idx := prsExpr(idx+1)

    op := ast.OpAssignVar{ Varname: v, Value: value }

    return op, idx
}

func prsIfStmt(idx int) (ast.IfStmt, int) {
    tokens := token.GetTokens()

    var op ast.IfStmt = ast.IfStmt{ IfPos: tokens[idx].Pos }

    op.Cond, idx = prsExpr(idx+1)
    op.Block, idx = prsBlock(idx+1)

    return op, idx
}

func prsIfElse(If ast.IfStmt, idx int) (ast.IfElseStmt, int) {
    tokens := token.GetTokens()
    var op ast.IfElseStmt = ast.IfElseStmt{ If: If, ElsePos: tokens[idx].Pos }

    op.Block, idx = prsBlock(idx+1)

    return op, idx
}

func prsWhileStmt(idx int) (ast.WhileStmt, int) {
    tokens := token.GetTokens()

    var op ast.WhileStmt = ast.WhileStmt{ WhilePos: tokens[idx].Pos, InitVal: nil }

    if isDec(idx) {
        op.Dec, idx = prsDecVar(idx)

        if tokens[idx+1].Type != token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        }

        var expr ast.OpExpr
        expr, idx = prsExpr(idx+2)

        if tokens[idx+1].Type == token.Comma {
            op.Cond, idx = prsExpr(idx+2)
            op.InitVal = expr
        } else {
            op.InitVal = &ast.LitExpr{ Val: token.Token{ Type: token.Number, Str: "0" }, Type: types.I32 }
            op.Cond = expr
        }
    } else {
        op.Cond, idx = prsExpr(idx+1)
    }

    op.Block, idx = prsBlock(idx+1)

    return op, idx
}

func prsForStmt(idx int) (ast.ForStmt, int) {
    tokens := token.GetTokens()

    var op ast.ForStmt = ast.ForStmt{
        ForPos: tokens[idx].Pos,
        Limit: nil,
        Start: &ast.LitExpr{
            Val: token.Token{ Str: "0", Type: token.Number },
            Type: types.I32,
        },
    }

    op.Dec, idx = prsDecVar(idx)

    op.Step = &ast.BinaryExpr{
        Operator: token.Token{ Type: token.Plus },
        OperandL: &ast.IdentExpr{ Ident: op.Dec.Varname },
        OperandR: &ast.LitExpr{
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: types.I32,
        },
    }

    if tokens[idx+1].Type == token.Comma {
        op.Limit, idx = prsExpr(idx+2)

        if tokens[idx+1].Type == token.Comma {
            op.Start, idx = prsExpr(idx+2)

            if tokens[idx+1].Type == token.Comma {
                op.Step, idx = prsExpr(idx+2)
            }
        }
    }

    op.Block, idx = prsBlock(idx+1)

    return op, idx
}

func prsBreak(idx int) (ast.BreakStmt, int) {
    tokens := token.GetTokens()

    var op ast.BreakStmt = ast.BreakStmt{ Pos: tokens[idx].Pos }
    return op, idx
}

func prsContinue(idx int) (ast.ContinueStmt, int) {
    tokens := token.GetTokens()

    var op ast.ContinueStmt = ast.ContinueStmt{ Pos: tokens[idx].Pos }
    return op, idx
}
