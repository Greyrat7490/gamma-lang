package prs

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/token"
    "gamma/types"
    "gamma/ast/identObj"
)

func prsStmt(tokens *token.Tokens, ignoreUnusedExpr bool) ast.Stmt {
    switch t := tokens.Next(); t.Type {
    case token.BraceL:
        b := prsBlock(tokens)
        return &b

    case token.If:
        ifStmt := prsIfStmt(tokens)

        if tokens.Cur().Str == "{" {
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

    case token.Number, token.Str, token.Boolean, token.ParenL:
        expr := prsExpr(tokens)
        if !ignoreUnusedExpr && expr.GetType() != nil {
            fmt.Fprintln(os.Stderr, "[ERROR] unused expr")
            fmt.Fprintln(os.Stderr, "\t" + expr.At())
            os.Exit(1)
        }
        return &ast.ExprStmt{ Expr: expr }

    case token.Name:
        if isDec(tokens) || isDefInfer(tokens) {
            return &ast.DeclStmt{ Decl: prsDefine(tokens) }
        }
        fallthrough

    case token.UndScr, token.XSwitch, token.Plus, token.Minus, token.Mul, token.Amp:
        expr := prsExpr(tokens)

        if tokens.Peek().Type == token.Assign {
            a := prsAssignVar(tokens, expr)
            return &a
        } else {
            if !ignoreUnusedExpr && expr.GetType() != nil {
                fmt.Fprintln(os.Stderr, "[ERROR] unused expr")
                fmt.Fprintln(os.Stderr, "\t" + expr.At())
                os.Exit(1)
            }
            return &ast.ExprStmt{ Expr: expr }
        }

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

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}
    }
}

func prsBlock(tokens *token.Tokens) ast.Block {
    block := ast.Block{ BraceLPos: tokens.Cur().Pos }

    for tokens.Peek().Type != token.BraceR {
        block.Stmts = append(block.Stmts, prsStmt(tokens, false))
    }

    block.BraceRPos = tokens.Next().Pos

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
    if tokens.Next().Str == "{" {
        return ast.If{ Pos: pos }
    }

    cond := prsExpr(tokens)

    // cond-switch with condBase
    if tokens.Cur().Str == "{" {
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
            op.Def.Value = &ast.Lit{ Val: token.Token{ Type: token.Number, Str: "0" }, Type: types.I32Type{} }
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

    var op ast.For = ast.For{
        ForPos: tokens.Cur().Pos,
        Limit: nil,
        Def: ast.DefVar{
            Value: &ast.Lit{
                Val: token.Token{ Str: "0", Type: token.Number },
                Type: types.I32Type{},
            },
        },
    }

    tokens.Next()
    dec := prsDecVar(tokens)
    op.Def.V = dec.V
    op.Def.Type = dec.Type

    op.Step = &ast.Binary{
        Operator: token.Token{ Type: token.Plus },
        OperandL: &ast.Ident{ Obj: op.Def.V, Name: op.Def.V.GetName(), Pos: op.Def.V.GetPos() },
        OperandR: &ast.Lit{
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: types.I32Type{},
        },
    }

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

    if r.F.GetRetType() != nil {
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
            return &ast.Binary{ OperandL: expr2, OperandR: expr1, Operator: token.Token{ Str: "||", Type: token.Or } }
        }

        return expr1
    }

    if b, ok := condBase.(*ast.Binary); ok {
        if expr2 != nil {
            *placeholder = expr1
            condCopy := *b
            return &ast.Binary{ OperandL: expr2, OperandR: &condCopy, Operator: token.Token{ Str: "||", Type: token.Or } }
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

func prsCases(tokens *token.Tokens, condBase ast.Expr) (cases []ast.Case) {
    placeholder := getPlaceholder(condBase)
    var conds ast.Expr = nil
    cur := -1

    expectColon := false
    lastStmtLine := 0

    for tokens.Peek().Type != token.BraceR {
        stmt := prsStmt(tokens, true)

        // comma-separated condition ----------
        for tokens.Peek().Type == token.Comma {
            tokens.Next()

            if cond, ok := stmt.(*ast.ExprStmt); ok {
                conds = completeCond(placeholder, condBase, cond.Expr, conds)

                tokens.Next()
                stmt = &ast.ExprStmt{ Expr: prsExpr(tokens) }
                expectColon = true
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \",\"")
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                os.Exit(1)
            }
        }

        if expectColon && tokens.Peek().Type != token.Colon {
            fmt.Fprintf(os.Stderr, "[ERROR] expected end of case condition(\":\") but got %v\n", tokens.Peek())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())
            os.Exit(1)
        }

        // case end without ";" before --------
        if tokens.Peek().Type == token.Colon {
            if tokens.Cur().Pos.Line == lastStmtLine {
                fmt.Fprintln(os.Stderr, "[ERROR] cases should always start in a new line or after a \";\"")
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                os.Exit(1)
            }

            tokens.Next()
            if nextColon := tokens.FindNext(token.Colon); tokens.Cur().Pos.Line == nextColon.Line {
                nextSemiCol := tokens.FindNext(token.SemiCol)

                if nextSemiCol.Line == -1 || (nextSemiCol.Line == nextColon.Line && nextSemiCol.Col > nextColon.Col) {
                    fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
                    fmt.Fprintln(os.Stderr, "\t" + nextColon.At())
                    os.Exit(1)
                }
            }

            if tokens.Last().Pos.Line < tokens.Cur().Pos.Line {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                os.Exit(1)
            }

            if cond, ok := stmt.(*ast.ExprStmt); ok {
                cond.Expr = completeCond(placeholder, condBase, cond.Expr, conds)
                cases = append(cases, ast.Case{ Cond: cond.Expr, ColonPos: tokens.Cur().Pos })
                cur++
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                os.Exit(1)
            }

            conds = nil
            expectColon = false
        // case stmts --------
        } else {
            if cur == -1 {
                fmt.Fprintln(os.Stderr, "[ERROR] missing case at the beginning of the cond-switch(or missing \":\")")
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                os.Exit(1)
            }

            lastStmtLine = tokens.Cur().Pos.Line
            cases[cur].Stmts = append(cases[cur].Stmts, stmt)

            // case end with before ";" -------
            if tokens.Peek().Type == token.SemiCol {
                pos := tokens.Next().Pos

                if tokens.Peek().Type == token.Colon {
                    fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
                    fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                    os.Exit(1)
                }

                stmt := prsStmt(tokens, true)

                if tokens.Next().Type != token.Colon {
                    fmt.Fprintln(os.Stderr, "[ERROR] \";\" should be at the end of the case")
                    fmt.Fprintln(os.Stderr, "\t" + pos.At())
                    os.Exit(1)
                }

                if cond, ok := stmt.(*ast.ExprStmt); ok {
                    cond.Expr = completeCond(placeholder, condBase, cond.Expr, conds)
                    cases = append(cases, ast.Case{ Cond: cond.Expr, ColonPos: tokens.Cur().Pos })
                    cur++
                } else {
                    fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \":\"")
                    fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                    os.Exit(1)
                }

                conds = nil
                expectColon = false
            }
        }
    }

    if tokens.Next().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" at the end of the cond-switch " +
            "but got \"%s\"(%v)\n", tokens.Cur().Str, tokens.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return
}

func prsSwitch(tokens *token.Tokens, pos token.Pos, condBase ast.Expr) ast.Switch {
    switchStmt := ast.Switch{ BraceLPos: pos }

    if tokens.Peek().Type == token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] empty cond-switch")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    identObj.StartScope()
    switchStmt.Cases = prsCases(tokens, condBase)
    identObj.EndScope()

    switchStmt.BraceRPos = tokens.Cur().Pos

    for i,c := range switchStmt.Cases {
        if len(c.Stmts) == 0 {
            fmt.Fprintln(os.Stderr, "[ERROR] no stmts provided for this case")
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }

        // is default case last
        if c.Cond == nil && i != len(switchStmt.Cases)-1 {
            i = len(switchStmt.Cases)-1 - i
            if i == 1 {
                fmt.Fprintln(os.Stderr, "[ERROR] one case after the default case (unreachable code)")
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %d cases after the default case (unreachable code)\n", i)
            }
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }
    }

    return switchStmt
}

func prsThrough(tokens *token.Tokens) ast.Through {
    return ast.Through{ Pos: tokens.Cur().Pos }
}
