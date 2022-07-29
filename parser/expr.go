package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/ast/identObj"
    "gorec/token"
    "gorec/types"
    "gorec/types/array"
)

type precedence int
const (
    LOGICAL_PRECEDENCE precedence = iota // &&, ||
    COMPARE_PRECEDENCE            = iota // ==, !=, <, <=, >, >=
    XSWITCH_PRECEDENCE            = iota // $ ... { ... }
    ADD_SUB_PRECEDENCE            = iota // +, -
    MUL_DIV_PRECEDENCE            = iota // *, /, %
    EXP_PRECEDENCE                = iota // **(TODO)
    PAREN_PRECEDENCE              = iota // ()
)

func prsExpr() ast.Expr {
    var expr ast.Expr
    switch token.Cur().Type {
    case token.Number, token.Str, token.Boolean:
        expr = prsLitExpr()

    case token.BrackL:
        return prsArrayLit()

    case token.Name:
        if token.Peek().Type == token.ParenL {
            return prsCallFn()  // only tmp because binary ops are not supported with func calls yet
        } else {
            expr = prsIdentExpr()
        }

    case token.XSwitch:
        expr = prsXSwitch()

    case token.UndScr:
        expr = prsIdentExpr()

    case token.ParenL:
        expr = prsParenExpr()

    case token.Plus, token.Minus, token.Mul, token.Amp:
        expr = prsUnaryExpr()

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression (got type %v)\n", token.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)

        return &ast.BadExpr{}
    }

    for token.Peek().Type == token.BrackL {
        token.Next()
        expr = prsIndexExpr(expr)
    }

    if isBinaryExpr() {
        expr = prsBinary(expr, 0)
    }

    return expr
}

func isUnaryExpr() bool {
    return  token.Cur().Type == token.Plus || token.Cur().Type == token.Minus ||
            token.Cur().Type == token.Mul  || token.Cur().Type == token.Amp
}

func isParenExpr() bool {
    return  token.Cur().Type == token.ParenL
}

func isBinaryExpr() bool {
    if token.Peek().Pos.Line > token.Cur().Pos.Line {
        return false
    }

    return  token.Peek().Type == token.Plus || token.Peek().Type == token.Minus ||
            token.Peek().Type == token.Mul  || token.Peek().Type == token.Div   ||
            token.Peek().Type == token.Mod  ||
            token.Peek().Type == token.And  || token.Peek().Type == token.Or    ||
            isComparison()
}

func isComparison() bool {
    return  token.Peek().Type == token.Eql || token.Peek().Type == token.Neq ||
            token.Peek().Type == token.Grt || token.Peek().Type == token.Lss ||
            token.Peek().Type == token.Geq || token.Peek().Type == token.Leq
}

func getPrecedence() precedence {
    switch {
    case token.Peek().Type == token.And || token.Peek().Type == token.Or:
        return LOGICAL_PRECEDENCE
    case isComparison():
        return COMPARE_PRECEDENCE
    case token.Peek().Type == token.Plus || token.Peek().Type == token.Minus:
        return ADD_SUB_PRECEDENCE
    case token.Peek().Type == token.Mul || token.Peek().Type == token.Div || token.Peek().Type == token.Mod:
        return MUL_DIV_PRECEDENCE
    case isParenExpr():
        return PAREN_PRECEDENCE
    default:
        return precedence(0)
    }
}

func prsIdentExpr() *ast.Ident {
    ident := token.Cur()

    // if wildcard ("_")
    if ident.Type == token.UndScr {
        return &ast.Ident{ Name: "_", Pos: ident.Pos, Obj: nil }
    }

    if ident.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", ident)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if obj := identObj.Get(ident.Str); obj != nil {
        return &ast.Ident{ Name: ident.Str, Pos: ident.Pos, Obj: obj }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", ident.Str)
    fmt.Fprintln(os.Stderr, "\t" + ident.At())
    os.Exit(1)
    return nil
}

func prsLitExpr() *ast.Lit {
    val := token.Cur()
    t := types.TypeOfVal(val.Str)

    return &ast.Lit{ Val: val, Type: t }
}

func prsArrayLit() *ast.ArrayLit {
    lit := ast.ArrayLit{ Pos: token.Cur().Pos }

    lit.Type = prsArrType()

    pos := token.Next()
    if pos.Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    lit.BraceLPos = pos.Pos

    token.Next()
    lit.Values = prsArrayLitExprs(lit.Type.GetLens())

    lit.BraceRPos = token.Cur().Pos

    var values []token.Token
    for _,v := range lit.Values {
        values = append(values, v.ConstEval())
    }
    lit.Idx = array.Add(lit.Type, values)

    return &lit
}

func prsArrayLitExprs(lenghts []uint64) (exprs []ast.Expr) {
    // TODO test len of parsed []expr
    switch token.Cur().Type {
        case token.BraceL:
            if len(lenghts) == 1 {
                // TODO better error
                fmt.Fprintln(os.Stderr, "[ERROR] unexpected \"{\" maybe a missing \"}\" or one \"{\" to much")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

            token.Next()
            tmp := prsArrayLitExprs(lenghts[1:])
            for _,e := range tmp {
                exprs = append(exprs, e)
            }

        case token.BraceR:
            return

        case token.XSwitch, token.UndScr:
            fmt.Fprintln(os.Stderr, "[ERROR] XSwitch(\"$\") and Wildcard(\"_\") are not supported in ArrayLits (yet)")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)

        default:
            if len(lenghts) == 1 {
                exprs = append(exprs, prsExpr())
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", token.Cur())
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }
    }

    if token.Next().Type == token.Comma {
        token.Next()
        tmp := prsArrayLitExprs(lenghts)
        for _,e := range tmp {
            exprs = append(exprs, e)
        }
    }

    if token.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return
}

func prsIndexExpr(expr ast.Expr) *ast.Indexed {
    res := ast.Indexed{ ArrExpr: expr, BrackLPos: token.Cur().Pos }

    token.Next()
    res.Indices = append(res.Indices, prsExpr())

    posR := token.Next()
    if posR.Type != token.BrackR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"]\" but got %v\n", posR)
        os.Exit(1)
    }

    for token.Peek().Type == token.BrackL {
        token.Next()
        token.Next()
        res.Indices = append(res.Indices, prsExpr())

        posR := token.Next()
        if posR.Type != token.BrackR {
            fmt.Fprintf(os.Stderr, "[ERROR] expected \"]\" but got %v\n", posR)
            os.Exit(1)
        }
    }

    return &res
}

func prsValue() ast.Expr {
    if token.Cur().Type == token.Name {
        return prsIdentExpr()
    } else {
        return prsLitExpr()
    }
}

func prsParenExpr() *ast.Paren {
    expr := ast.Paren{ ParenLPos: token.Cur().Pos }

    token.Next()
    expr.Expr = prsExpr()

    expr.ParenRPos = token.Next().Pos

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return &expr
}

func prsUnaryExpr() *ast.Unary {
    expr := ast.Unary{ Operator: token.Cur() }

    switch expr.Operator.Type {
    case token.Mul:
        if token.Next().Type == token.ParenL {
            expr.Operand = prsParenExpr()
        } else {
            expr.Operand = prsIdentExpr()
        }
    case token.Amp:
        token.Next()
        expr.Operand = prsIdentExpr()
    default:
        token.Next()
        expr.Operand = prsValue()
    }

    return &expr
}

func prsCaseExpr(condBase ast.Expr, placeholder *ast.Expr, lastCaseEnd token.Pos) (caseExpr ast.XCase) {
    if token.Cur().Type == token.Colon {
        if token.Last().Pos.Line == token.Cur().Pos.Line {
            fmt.Fprintln(os.Stderr, "[ERROR] missing case body(expr) for this case")
            fmt.Fprintln(os.Stderr, "\t" + lastCaseEnd.At())
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        }
        os.Exit(1)
    }
    if token.Cur().Type == token.Comma {
        fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \",\"")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    if token.Last().Pos.Line == token.Cur().Pos.Line && token.Last().Type != token.SemiCol && token.Last().Type != token.BraceL {
        fmt.Fprintln(os.Stderr, "[ERROR] cases should always start in a new line or after a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    // parse case cond(s) ----------------
    expr := prsExpr()
    var conds ast.Expr = nil
    for token.Next().Type == token.Comma {
        conds = completeCond(placeholder, condBase, expr, conds)

        if token.Peek().Type == token.Colon || token.Peek().Type == token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: no expr after \",\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        token.Next()
        expr = prsExpr()
    }

    caseExpr.ColonPos = token.Cur().Pos
    caseExpr.Cond = completeCond(placeholder, condBase, expr, conds)

    if token.Cur().Type != token.Colon {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \":\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    if nextColon := token.FindNext(token.Colon); token.Cur().Pos.Line == nextColon.Line {
        nextSemiCol := token.FindNext(token.SemiCol)

        if nextSemiCol.Line == -1 || (nextSemiCol.Line == nextColon.Line && nextSemiCol.Col > nextColon.Col) {
            fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
            fmt.Fprintln(os.Stderr, "\t" + nextColon.At())
            os.Exit(1)
        }
    }


    // parse case body -------------------
    if token.Peek().Type == token.SemiCol {
        fmt.Fprintln(os.Stderr, "[ERROR] missing case body(expr) for this case")
        fmt.Fprintln(os.Stderr, "\t" + token.Last().At())
        os.Exit(1)
    }

    token.Next()
    caseExpr.Expr = prsExpr()

    if token.Peek().Type == token.SemiCol { token.Next() }

    return
}

func prsXSwitch() *ast.XSwitch {
    switchExpr := ast.XSwitch{ Pos: token.Cur().Pos }
    var condBase ast.Expr = nil
    var placeholder *ast.Expr = nil

    // set condBase -----------------------
    if token.Next().Type != token.BraceL {
        condBase = prsExpr()
        placeholder = getPlaceholder(condBase)
    }

    // parse cases ------------------------
    if token.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" at the beginning of the xswitch " +
            "but got \"%s\"(%v)\n", token.Cur().Str, token.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    switchExpr.BraceLPos = token.Cur().Pos

    lastCaseEnd := token.Pos{}
    for token.Next().Type != token.BraceR {
        expr := prsCaseExpr(condBase, placeholder, lastCaseEnd)
        lastCaseEnd = expr.ColonPos
        switchExpr.Cases = append(switchExpr.Cases, expr)
    }

    switchExpr.BraceRPos = token.Cur().Pos


    // catch some syntax errors -----------
    if len(switchExpr.Cases) == 0 {
        fmt.Fprintln(os.Stderr, "[ERROR] empty xswitch")
        fmt.Fprintln(os.Stderr, "\t" + switchExpr.BraceLPos.At())
        os.Exit(1)
    }
    for i,c := range switchExpr.Cases {
        if c.Cond == nil && i != len(switchExpr.Cases)-1 {
            i = len(switchExpr.Cases)-1 - i
            if i == 1 {
                fmt.Fprintln(os.Stderr, "[ERROR] one case after the default case (unreachable code)")
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %d cases after the default case (unreachable code)\n", i)
            }
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }
    }
    if switchExpr.Cases[len(switchExpr.Cases)-1].Cond != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] every xswitch requires a default case")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return &switchExpr
}

func prsBinary(expr ast.Expr, min_precedence precedence) ast.Expr {
    for isBinaryExpr() && getPrecedence() >= min_precedence {
        var b ast.Binary
        b.OperandL = expr

        precedenceL := getPrecedence()
        b.Operator = token.Next()

        token.Next()
        precedenceR := getPrecedence()

        // switch/xswitch
        if token.Cur().Type == token.BraceL {
            return &b
        }

        switch {
        case isParenExpr():
            b.OperandR = prsParenExpr()
        case isUnaryExpr():
            b.OperandR = prsUnaryExpr()
        default:
            b.OperandR = prsValue()
        }

        if isBinaryExpr() {
            b.OperandR = prsBinary(b.OperandR, precedenceL+1)
        }

        // left to right as correct order of operations
        if precedenceR > precedenceL {
            swap(&b)
        }

        expr = &b
    }

    return expr
}

func swap(expr *ast.Binary) {
    switch expr.Operator.Type {
    case token.Minus:
        expr.Operator.Type = token.Plus
        expr.Operator.Str = "+"

        t := token.Token{ Type: token.Minus, Str: "-" }
        expr.OperandR = &ast.Unary{ Operator: t, Operand: expr.OperandR }

    // TODO: proper fix
    // only tmp
    case token.Div:
        return

    case token.Geq:
        expr.Operator.Type = token.Leq
        expr.Operator.Str = "<="

    case token.Leq:
        expr.Operator.Type = token.Geq
        expr.Operator.Str = ">="

    case token.Grt:
        expr.Operator.Type = token.Lss
        expr.Operator.Str = "<"

    case token.Lss:
        expr.Operator.Type = token.Grt
        expr.Operator.Str = ">"
    }

    tmp := expr.OperandR
    expr.OperandR = expr.OperandL
    expr.OperandL = tmp
}


func prsCallFn() *ast.FnCall {
    ident := prsIdentExpr()
    posL := token.Next().Pos
    vals := prsPassArgs()
    posR := token.Cur().Pos

    return &ast.FnCall{ Ident: *ident, Values: vals, ParenLPos: posL, ParenRPos: posR }
}

func prsPassArgs() []ast.Expr {
    if token.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    var values []ast.Expr

    if token.Next().Type == token.ParenR {
        return values
    }

    values = append(values, prsExpr())
    for token.Next().Type == token.Comma {
        token.Next()
        values = append(values, prsExpr())
    }

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return values
}
