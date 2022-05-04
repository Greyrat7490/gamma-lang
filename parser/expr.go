package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

type precedence int
const (
    COMPARE_PRECEDENCE  precedence = 0
    ADD_SUB_PRECEDENCE  precedence = 1
    MUL_DIV_PRECEDENCE  precedence = 2
    EXP_ROOT_PRECEDENCE precedence = 3
    PAREN_PRECEDENCE    precedence = 4
)

func prsExpr(idx int) (ast.OpExpr, int) {
    var expr ast.OpExpr
    switch {
    case isLitExpr(idx):
        expr, idx = prsLitExpr(idx)

    case isIdentExpr(idx):
        expr, idx = prsIdentExpr(idx)

    case isParenExpr(idx):
        expr, idx = prsParenExpr(idx)

    case isUnaryExpr(idx):
        expr, idx = prsUnaryExpr(idx)

    // TODO: OpFnCall

    default:
        tokens := token.GetTokens()

        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression (got type %s)\n", tokens[idx].Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)

        return &ast.BadExpr{}, -1
    }

    if isBinaryExpr(idx) {
        expr, idx = prsBinary(idx, expr, 0)
    }

    return expr, idx
}

func isLitExpr(idx int) bool {
    tokens := token.GetTokens()
    return tokens[idx].Type == token.Number || tokens[idx].Type == token.Str || tokens[idx].Type == token.Boolean
}
func isIdentExpr(idx int) bool {
    tokens := token.GetTokens()
    return tokens[idx].Type == token.Name
}

func isUnaryExpr(idx int) bool {
    tokens := token.GetTokens()
    return tokens[idx].Type == token.Plus || tokens[idx].Type == token.Minus
}

func isBinaryExpr(idx int) bool {
    tokens := token.GetTokens()

    if idx + 2 >= len(tokens) {
        return false
    }

    return  tokens[idx+1].Type == token.Plus || tokens[idx+1].Type == token.Minus ||
            tokens[idx+1].Type == token.Mul  || tokens[idx+1].Type == token.Div ||
            isCompareExpr(idx)
}

func isParenExpr(idx int) bool {
    tokens := token.GetTokens()
    return tokens[idx].Type == token.ParenL
}

func isCompareExpr(idx int) bool {
    tokens := token.GetTokens()

    if idx + 2 >= len(tokens) {
        return false
    }

    return  tokens[idx+1].Type == token.Eql || tokens[idx+1].Type == token.Neq ||
            tokens[idx+1].Type == token.Grt || tokens[idx+1].Type == token.Lss ||
            tokens[idx+1].Type == token.Geq || tokens[idx+1].Type == token.Leq
}

func getPrecedence(idx int) precedence {
    tokens := token.GetTokens()

    if isCompareExpr(idx) {
        return COMPARE_PRECEDENCE
    } else if tokens[idx+1].Type == token.Plus || tokens[idx+1].Type == token.Minus {
        return ADD_SUB_PRECEDENCE
    } else if tokens[idx+1].Type == token.Mul || tokens[idx+1].Type == token.Div {
        return MUL_DIV_PRECEDENCE
    } else if isParenExpr(idx) {
        return PAREN_PRECEDENCE
    } else {
        return precedence(0)
    }
}

func prsIdentExpr(idx int) (*ast.IdentExpr, int) {
    tokens := token.GetTokens()
    return &ast.IdentExpr{ Ident: tokens[idx] }, idx
}

func prsLitExpr(idx int) (*ast.LitExpr, int) {
    tokens := token.GetTokens()
    return &ast.LitExpr{ Val: tokens[idx], Type: types.TypeOfVal(tokens[idx].Str) }, idx
}

func prsValue(idx int) (ast.OpExpr, int) {
    tokens := token.GetTokens()
    if tokens[idx].Type == token.Name {
        return prsIdentExpr(idx)
    } else {
        return prsLitExpr(idx)
    }
}

func prsParenExpr(idx int) (*ast.ParenExpr, int) {
    tokens := token.GetTokens()
    expr := ast.ParenExpr{ ParenLPos: tokens[idx].Pos }

    expr.Expr, idx = prsExpr(idx+1)
    idx++

    if tokens[idx].Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected ) but got \"%s\"(%s)\n", tokens[idx].Str, tokens[idx].Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    expr.ParenRPos = tokens[idx].Pos

    return &expr, idx
}

func prsUnaryExpr(idx int) (*ast.UnaryExpr, int) {
    tokens := token.GetTokens()
    expr := ast.UnaryExpr{ Operator: tokens[idx] }

    expr.Operand, idx = prsValue(idx+1)

    return &expr, idx
}

func prsBinary(idx int, expr ast.OpExpr, min_precedence precedence) (ast.OpExpr, int) {
    tokens := token.GetTokens()

    for isBinaryExpr(idx) && getPrecedence(idx) >= min_precedence {
        var b ast.BinaryExpr

        precedenceL := getPrecedence(idx)
        precedenceR := getPrecedence(idx+2)

        b.Operator = tokens[idx+1]
        b.OperandL = expr

        if isParenExpr(idx+2) {
            b.OperandR, idx = prsParenExpr(idx+2)
        } else if isUnaryExpr(idx+2) {
            b.OperandR, idx = prsUnaryExpr(idx+2)
        } else {
            b.OperandR, idx = prsValue(idx+2)
        }

        if isBinaryExpr(idx) {
            b.OperandR, idx = prsBinary(idx, b.OperandR, precedenceL+1)
        }

        // left to right as correct order of operations
        if precedenceR > precedenceL {
            swap(&b)
        }

        expr = &b
    }

    return expr, idx
}

func swap(expr *ast.BinaryExpr) {
    if expr.Operator.Type == token.Minus {
        expr.Operator.Type = token.Plus
        expr.Operator.Str = "+"

        t := token.Token{ Type: token.Minus, Str: "-" }
        expr.OperandR = &ast.UnaryExpr{ Operator: t, Operand: expr.OperandR }
    }

    tmp := expr.OperandR
    expr.OperandR = expr.OperandL
    expr.OperandL = tmp
}
