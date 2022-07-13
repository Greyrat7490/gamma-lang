package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsStmt(ignoreUnusedExpr bool) ast.Stmt {
    switch t := token.Next(); t.Type {
    case token.BraceL:
        b := prsBlock()
        return &b

    case token.If:
        ifStmt := prsIfStmt()

        if token.Cur().Str == "{" {
            switchStmt := prsSwitch(ifStmt.Pos, ifStmt.Cond)
            return &switchStmt
        }

        return &ifStmt

    case token.Through:
        t := prsThrough()
        return &t

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

    case token.Number, token.Str, token.Boolean, token.ParenL:
        expr := prsExpr()
        if !ignoreUnusedExpr && expr.GetType() != nil{
            fmt.Fprintln(os.Stderr, "[ERROR] unused expr")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }
        return &ast.ExprStmt{ Expr: expr }

    case token.Name, token.UndScr, token.Plus, token.Minus, token.Mul, token.Amp:
        // define var (type is given)
        if token.Peek().Type == token.Typename {
            d := prsDefVar()
            return &ast.DeclStmt{ Decl: &d }
        }
        // define var (infer the type with the value)
        if token.Peek().Type == token.DefVar {
            d := prsDefVarInfer()
            return &ast.DeclStmt{ Decl: &d }
        }

        expr := prsExpr()

        if token.Peek().Type == token.Assign {
            a := prsAssignVar(expr)
            return &a
        } else {
            if !ignoreUnusedExpr && expr.GetType() != nil {
                fmt.Fprintln(os.Stderr, "[ERROR] unused expr")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }
            return &ast.ExprStmt{ Expr: expr }
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

    case token.Fn:
        fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}
    }
}

func prsBlock() ast.Block {
    block := ast.Block{ BraceLPos: token.Cur().Pos }

    for token.Peek().Type != token.BraceR {
        block.Stmts = append(block.Stmts, prsStmt(false))
    }

    block.BraceRPos = token.Next().Pos

    return block
}

func prsAssignVar(dest ast.Expr) ast.Assign {
    pos := token.Next().Pos
    token.Next()
    val := prsExpr()

    return ast.Assign{ Pos: pos, Dest: dest, Value: val }
}

func prsIfStmt() ast.If {
    pos := token.Cur().Pos

    // cond-switch without condBase
    if token.Next().Str == "{" {
        return ast.If{ Pos: pos }
    }

    cond := prsExpr()

    // cond-switch with condBase
    if token.Cur().Str == "{" {
        return ast.If{ Pos: pos, Cond: cond }
    }

    // normal if
    token.Next()
    block := prsBlock()

    ifStmt := ast.If{ Pos: pos, Cond: cond, Block: block }

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

func prsElif() ast.Elif {
    return ast.Elif(prsIfStmt())
}

func prsElse() ast.Else {
    pos := token.Cur().Pos
    token.Next()
    block := prsBlock()

    return ast.Else{ ElsePos: pos, Block: block }
}

func prsWhileStmt() ast.While {
    var op ast.While = ast.While{ WhilePos: token.Cur().Pos, Def: nil }

    if token.Peek().Type == token.Name && token.Peek2().Type == token.Typename {
        token.Next()
        dec := prsDecVar()
        op.Def = &ast.DefVar{ Name: dec.Name, Type: dec.Type }

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
            op.Def.Value = expr
            token.Next()
        } else {
            op.Def.Value = &ast.Lit{ Val: token.Token{ Type: token.Number, Str: "0" }, Type: types.I32Type{} }
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

func prsForStmt() ast.For {
    var op ast.For = ast.For{
        ForPos: token.Cur().Pos,
        Limit: nil,
        Def: ast.DefVar{
            Value: &ast.Lit{
                Val: token.Token{ Str: "0", Type: token.Number },
                Type: types.I32Type{},
            },
        },
    }

    token.Next()
    dec := prsDecVar()
    op.Def.Name = dec.Name
    op.Def.Type = dec.Type

    op.Step = &ast.Binary{
        Operator: token.Token{ Type: token.Plus },
        OperandL: &ast.Ident{ Ident: op.Def.Name },
        OperandR: &ast.Lit{
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: types.I32Type{},
        },
    }

    if token.Next().Type == token.Comma {
        token.Next()
        op.Limit = prsExpr()

        if token.Next().Type == token.Comma {
            token.Next()
            op.Def.Value = prsExpr()

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

func prsBreak() ast.Break {
    var op ast.Break = ast.Break{ Pos: token.Cur().Pos }
    return op
}

func prsContinue() ast.Continue {
    var op ast.Continue = ast.Continue{ Pos: token.Cur().Pos }
    return op
}


func getPlaceholder(cond ast.Expr) (expr *ast.Expr) {
    if cond == nil {
        return nil
    }

    for {
        if b, ok := cond.(*ast.Binary); !ok {
            fmt.Fprintln(os.Stderr, "[ERROR] expected condition to be a BinaryExpr")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
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
        if ident.Ident.Type == token.UndScr {
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
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return nil
    }
}

func prsCases(condBase ast.Expr) (cases []ast.Case) {
    placeholder := getPlaceholder(condBase)
    var conds ast.Expr = nil
    cur := -1

    expectColon := false
    lastStmtLine := 0

    for token.Peek().Type != token.BraceR {
        stmt := prsStmt(true)

        // comma-separated condition ----------
        for token.Peek().Type == token.Comma {
            token.Next()

            if cond, ok := stmt.(*ast.ExprStmt); ok {
                conds = completeCond(placeholder, condBase, cond.Expr, conds)

                token.Next()
                stmt = &ast.ExprStmt{ Expr: prsExpr() }
                expectColon = true
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \",\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }
        }

        if expectColon && token.Peek().Type != token.Colon {
            fmt.Fprintf(os.Stderr, "[ERROR] expected end of case condition(\":\") but got %v\n", token.Peek())
            fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
            os.Exit(1)
        }

        // case end without ";" before --------
        if token.Peek().Type == token.Colon {
            if token.Cur().Pos.Line == lastStmtLine {
                fmt.Fprintln(os.Stderr, "[ERROR] cases should always start in a new line or after a \";\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

            token.Next()
            if nextColon := token.FindNext(token.Colon); token.Cur().Pos.Line == nextColon.Line {
                nextSemiCol := token.FindNext(token.SemiCol)

                if nextSemiCol.Line == -1 || (nextSemiCol.Line == nextColon.Line && nextSemiCol.Col > nextColon.Col) {
                    fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
                    fmt.Fprintln(os.Stderr, "\t" + nextColon.At())
                    os.Exit(1)
                }
            }

            if token.Last().Pos.Line < token.Cur().Pos.Line {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

            if cond, ok := stmt.(*ast.ExprStmt); ok {
                cond.Expr = completeCond(placeholder, condBase, cond.Expr, conds)
                cases = append(cases, ast.Case{ Cond: cond.Expr, ColonPos: token.Cur().Pos })
                cur++
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

            conds = nil
            expectColon = false
        // case stmts --------
        } else {
            if cur == -1 {
                fmt.Fprintln(os.Stderr, "[ERROR] missing case at the beginning of the cond-switch(or missing \":\")")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

            lastStmtLine = token.Cur().Pos.Line
            cases[cur].Stmts = append(cases[cur].Stmts, stmt)

            // case end with before ";" -------
            if token.Peek().Type == token.SemiCol {
                pos := token.Next().Pos

                if token.Peek().Type == token.Colon {
                    fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
                    fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                    os.Exit(1)
                }

                stmt := prsStmt(true)

                if token.Next().Type != token.Colon {
                    fmt.Fprintln(os.Stderr, "[ERROR] \";\" should be at the end of the case")
                    fmt.Fprintln(os.Stderr, "\t" + pos.At())
                    os.Exit(1)
                }

                if cond, ok := stmt.(*ast.ExprStmt); ok {
                    cond.Expr = completeCond(placeholder, condBase, cond.Expr, conds)
                    cases = append(cases, ast.Case{ Cond: cond.Expr, ColonPos: token.Cur().Pos })
                    cur++
                } else {
                    fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \":\"")
                    fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                    os.Exit(1)
                }

                conds = nil
                expectColon = false
            }
        }
    }

    if token.Next().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" at the end of the cond-switch " +
            "but got \"%s\"(%v)\n", token.Cur().Str, token.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return
}

func prsSwitch(pos token.Pos, condBase ast.Expr) ast.Switch {
    switchStmt := ast.Switch{ BraceLPos: pos }

    if token.Peek().Type == token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] empty cond-switch")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    switchStmt.Cases = prsCases(condBase)

    switchStmt.BraceRPos = token.Cur().Pos

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

func prsThrough() ast.Through {
    return ast.Through{ Pos: token.Cur().Pos }
}
