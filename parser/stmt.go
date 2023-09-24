package prs

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/ast"
    "gamma/ast/identObj"
)

func prsStmt(tokens *token.Tokens) ast.Stmt {
    switch t := tokens.Next(); t.Type {
    case token.BraceL:
        b := prsBlock(tokens)
        return &b

    case token.If:
        ifStmt := prsIfStmt(tokens)

        if tokens.Cur().Type == token.BraceL {
            switchStmt := prsSwitch(tokens, ifStmt.Pos, ifStmt.Cond)
            return &switchStmt
        }

        return &ifStmt

    case token.Through:
        t := prsThrough(tokens)
        return &t

    case token.While:
        w := prsWhileStmt(tokens)
        return &w

    case token.For:
        f := prsForStmt(tokens)
        return &f

    case token.Break:
        b := prsBreak(tokens)
        return &b

    case token.Continue:
        c := prsContinue(tokens)
        return &c

    case token.Ret:
        r := prsRet(tokens)
        return &r

    case token.Number, token.Str, token.Char, token.Boolean, token.ParenL:
        return &ast.ExprStmt{ Expr: prsExpr(tokens) }

    case token.UndScr:
        return &ast.DeclStmt{ Decl: prsDefine(tokens) }

    case token.Name:
        if isDec(tokens) || isDefInfer(tokens) {
            return &ast.DeclStmt{ Decl: prsDefine(tokens) }
        }
        fallthrough

    case token.XSwitch, token.Plus, token.Minus, token.Mul, token.Amp, token.Self:
        e := prsExpr(tokens)

        if tokens.Peek().Type == token.Assign {
            a := prsAssignVar(tokens, e)
            return &a
        }
        return &ast.ExprStmt{ Expr: e }

    case token.Typename:
        return &ast.ExprStmt{ Expr: prsExpr(tokens) }

    case token.Elif:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (elif without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Else:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (else without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Assign:
        fmt.Fprintf(os.Stderr, "[ERROR] no destination for assignment\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Fn:
        fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Import:
        fmt.Fprintln(os.Stderr, "[ERROR] importing is only allowed at the beginning of a file")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}
    }
}

func prsBlock(tokens *token.Tokens) ast.Block {
    identObj.StartScope()

    block := ast.Block{ BraceLPos: tokens.Cur().Pos }

    for tokens.Peek().Type != token.BraceR {
        block.Stmts = append(block.Stmts, prsStmt(tokens))
    }

    block.BraceRPos = tokens.Next().Pos

    identObj.EndScope()
    return block
}

func prsAssignVar(tokens *token.Tokens, dest ast.Expr) ast.Assign {
    pos := tokens.Next().Pos
    tokens.Next()
    val := prsExpr(tokens)

    return ast.Assign{ Pos: pos, Dest: dest, Value: val }
}

func prsIfStmt(tokens *token.Tokens) ast.If {
    identObj.StartScope()
    defer identObj.EndScope()

    pos := tokens.Cur().Pos

    // cond-switch without condBase
    if tokens.Next().Type == token.BraceL {
        return ast.If{ Pos: pos }
    }

    cond := prsExpr(tokens)

    // unwrap condition
    if tokens.Peek().Type == token.Colon {
        cond = prsUnwrap(tokens, cond)
    }

    // cond-switch with condBase
    if tokens.Cur().Type == token.BraceL {
        return ast.If{ Pos: pos, Cond: cond }
    }

    // normal if
    tokens.Next()
    block := prsBlock(tokens)

    ifStmt := ast.If{ Pos: pos, Cond: cond, Block: block }

    if tokens.Peek().Type == token.Else {
        tokens.Next()
        elseStmt := prsElse(tokens)
        ifStmt.Else = &elseStmt
    } else if tokens.Peek().Type == token.Elif {
        tokens.Next()
        elifStmt := prsElif(tokens)
        ifStmt.Elif = &elifStmt
    }

    return ifStmt
}

func prsElif(tokens *token.Tokens) ast.Elif {
    return ast.Elif(prsIfStmt(tokens))
}

func prsElse(tokens *token.Tokens) ast.Else {
    identObj.StartScope()
    defer identObj.EndScope()

    pos := tokens.Cur().Pos
    tokens.Next()
    block := prsBlock(tokens)

    return ast.Else{ ElsePos: pos, Block: block }
}

func prsWhileStmt(tokens *token.Tokens) ast.While {
    identObj.StartScope()
    defer identObj.EndScope()

    var op ast.While = ast.While{ WhilePos: tokens.Cur().Pos, Def: nil }

    tokens.Next()
    if isDec(tokens) {
        dec := prsDecVar(tokens)
        op.Def = &ast.DefVar{ V: dec.V, Type: dec.Type }

        if tokens.Next().Type != token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        tokens.Next()
        expr := prsExpr(tokens)

        if tokens.Next().Type == token.Comma {
            tokens.Next()
            op.Cond = prsExpr(tokens)
            op.Def.Value = expr
            tokens.Next()
        } else {
            op.Def.Value = &ast.IntLit{ Repr: 0, Val: token.Token{ Str: "0" }, Type: dec.Type }
            op.Cond = expr
        }
    } else {
        op.Cond = prsExpr(tokens)
        tokens.Next()
    }

    op.Block = prsBlock(tokens)

    return op
}

func prsForStmt(tokens *token.Tokens) ast.For {
    identObj.StartScope()
    defer identObj.EndScope()

    var op ast.For = ast.For{ ForPos: tokens.Cur().Pos, Limit: nil }

    tokens.Next()
    dec := prsDecVar(tokens)
    op.Def = ast.DefVar{ V: dec.V, Type: dec.Type }

    if tokens.Next().Type == token.Comma {
        tokens.Next()
        op.Limit = prsExpr(tokens)

        if tokens.Next().Type == token.Comma {
            tokens.Next()
            op.Def.Value = prsExpr(tokens)

            if tokens.Next().Type == token.Comma {
                tokens.Next()
                op.Step = prsExpr(tokens)
                tokens.Next()
            }
        }
    }

    if op.Def.Value == nil {
        op.Def.Value = &ast.IntLit{
            Repr: 0,
            Val: token.Token{ Str: "0", Type: token.Number },
            Type: dec.Type,
        }
    }

    if op.Step == nil {
        lit := &ast.IntLit{
            Repr: 1,
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: dec.Type,
        }

        op.Step = &ast.Binary{
            Operator: token.Token{ Type: token.Plus },
            OperandL: &ast.Ident{ Obj: op.Def.V, Name: op.Def.V.GetName(), Pos: op.Def.V.GetPos() },
            OperandR: lit,
            Type: dec.Type,
        }
    }

    op.Block = prsBlock(tokens)

    return op
}

func prsBreak(tokens *token.Tokens) ast.Break {
    return ast.Break{ Pos: tokens.Cur().Pos }
}

func prsContinue(tokens *token.Tokens) ast.Continue {
    return ast.Continue{ Pos: tokens.Cur().Pos }
}

func prsRet(tokens *token.Tokens) ast.Ret {
    r := ast.Ret{ Pos: tokens.Cur().Pos, F: identObj.GetCurFunc() }

    if tokens.Peek().Pos.Line == r.Pos.Line && tokens.Peek().Type != token.BraceR {
        tokens.Next()
        r.RetExpr = prsExpr(tokens)
    }

    return r
}

func getPlaceholder(cond ast.Expr) (expr *ast.Expr) {
    if cond == nil {
        return nil
    }

    for {
        if b, ok := cond.(*ast.Binary); !ok {
            fmt.Fprintln(os.Stderr, "[ERROR] expected condition to be a BinaryExpr")
            fmt.Fprintln(os.Stderr, "\t" + cond.At())
            os.Exit(1)
            return nil
        } else {
            if b.Operator.Type == token.Eql || b.Operator.Type == token.Neq ||
               b.Operator.Type == token.Grt || b.Operator.Type == token.Lss ||
               b.Operator.Type == token.Geq || b.Operator.Type == token.Leq {
                return &b.OperandR
            }

            cond = b.OperandR
        }
    }
}

// replace placeholder in condBase with expr1
// if expr2 is set they get combined with logical or
func completeCond(placeholder *ast.Expr, condBase ast.Expr, expr1 ast.Expr, expr2 ast.Expr) ast.Expr {
    if ident, ok := expr1.(*ast.Ident); ok {
        if ident.Obj == nil && ident.Name == "_" {
            return nil
        }
    }

    if condBase == nil {
        if expr2 != nil {
            return &ast.Binary{
                OperandL: expr2,
                OperandR: expr1,
                Operator: token.Token{ Str: "||", Type: token.Or },
                Type: types.BoolType{},
            }
        }

        return expr1
    }

    if b, ok := condBase.(*ast.Binary); ok {
        if expr2 != nil {
            *placeholder = expr1
            condCopy := *b
            return &ast.Binary{
                OperandL: expr2,
                OperandR: &condCopy,
                Operator: token.Token{ Str: "||", Type: token.Or },
                Type: types.BoolType{},
            }
        } else {
            *placeholder = expr1
            condCopy := *b
            return &condCopy
        }
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] expected condition to be a BinaryExpr")
        fmt.Fprintln(os.Stderr, "\t" + condBase.At())
        os.Exit(1)
        return nil
    }
}

func copyUnwrap(unwrapExpr ast.Expr, newElemName token.Token) *ast.Unwrap {
    if unwrap, ok := unwrapExpr.(*ast.Unwrap); ok {
        unwrapCopy := *unwrap
        unwrapCopy.ElemName = newElemName
        return &unwrapCopy
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] (internal) expected condition to be an Unwrap")
        fmt.Fprintln(os.Stderr, "\t" + unwrapExpr.At())
        os.Exit(1)
        return nil
    }
}

func prsUnwrapCaseCond(tokens *token.Tokens, condBase ast.Expr) (cond ast.Expr, colonPos token.Pos) {
    elemName := tokens.Cur()
    if elemName.Type != token.UndScr {
        cond = prsUnwrapElem(tokens, copyUnwrap(condBase, elemName))
    }
    colonPos = tokens.Next().Pos

    return
}

func prsUnwrapCase(tokens *token.Tokens, condBase ast.Expr) ast.Case {
    cond, colonPos := prsUnwrapCaseCond(tokens, condBase)

    stmt := prsStmt(tokens)

    if colonPos.Line == tokens.Peek().Pos.Line && tokens.Peek().Type != token.SemiCol && tokens.Peek().Type != token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.SemiCol { tokens.Next() }

    return ast.Case{ Cond: cond, ColonPos: colonPos, Stmt: stmt }
}

func prsCase(tokens *token.Tokens, condBase ast.Expr, placeholder *ast.Expr) ast.Case {
    cond, colonPos := prsCaseCond(tokens, condBase, placeholder)

    stmt := prsStmt(tokens)

    if colonPos.Line == tokens.Peek().Pos.Line && tokens.Peek().Type != token.SemiCol && tokens.Peek().Type != token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.SemiCol { tokens.Next() }

    return ast.Case{ Cond: cond, ColonPos: colonPos, Stmt: stmt }
}

func prsCases(tokens *token.Tokens, condBase ast.Expr) (cases []ast.Case) {
    if _,ok := condBase.(*ast.Unwrap); ok { 
        for tokens.Next().Type != token.BraceR {
            cases = append(cases, prsUnwrapCase(tokens, condBase))
        }
    } else {
        placeholder := getPlaceholder(condBase)
        for tokens.Next().Type != token.BraceR {
            cases = append(cases, prsCase(tokens, condBase, placeholder))
        }
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" at the end of the switch (got \"%s\")\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return
}

func prsSwitch(tokens *token.Tokens, pos token.Pos, condBase ast.Expr) ast.Switch {
    switchStmt := ast.Switch{ BraceLPos: pos }

    if tokens.Peek().Type == token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] empty switch")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    identObj.StartScope()
    switchStmt.Cases = prsCases(tokens, condBase)
    identObj.EndScope()
    switchStmt.BraceRPos = tokens.Cur().Pos

    return switchStmt
}

func prsThrough(tokens *token.Tokens) ast.Through {
    return ast.Through{ Pos: tokens.Cur().Pos }
}
