package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsExpr(idx int) (ast.OpExpr, int) {
    tokens := token.GetTokens()
    value := tokens[idx]

    switch value.Type {
    case token.Name:
        return prsIdentExpr(idx)

    case token.Plus, token.Minus:
        var expr ast.OpExpr
        expr, idx = prsUnaryExpr(idx)
        return prsBinary(idx, expr, 0)

    case token.Number, token.Str:
        var expr ast.OpExpr
        expr, idx = prsLitExpr(idx)
        return prsBinary(idx, expr, 0)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression (got type %s)\n", value.Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadExpr{}, -1
    }
}

func isUnaryExpr(idx int) bool {
    tokens := token.GetTokens()
    return tokens[idx].Type == token.Plus || tokens[idx].Type == token.Minus
}

func isBinaryExpr(idx int) bool {
    tokens := token.GetTokens()

    if idx + 1 >= len(tokens) {
        return false
    }

    return tokens[idx+1].Type == token.Plus || tokens[idx+1].Type == token.Minus ||
                tokens[idx+1].Type == token.Mul || tokens[idx+1].Type == token.Div
}

func getPrecedence(t token.TokenType) int {
    if t == token.Plus || t == token.Minus {
        return 1
    } else {
        return 2
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

func prsUnaryExpr(idx int) (*ast.UnaryExpr, int) {
    tokens := token.GetTokens()
    expr := ast.UnaryExpr{ Operator: tokens[idx] }

    expr.Operand, idx = prsValue(idx+1)

    return &expr, idx
}

func prsBinary(idx int, lhs ast.OpExpr, min_precedence int) (ast.OpExpr, int) {
    tokens := token.GetTokens()

    for isBinaryExpr(idx) && getPrecedence(tokens[idx+1].Type) >= min_precedence {
        precedence := getPrecedence(tokens[idx+1].Type)

        // TODO detect if 2 or more registers are needed
        // if an OperandR is a BinaryExpr
        // i.e. 2 * 2 + 3 * 3

        var b ast.BinaryExpr
        b.Operator = tokens[idx+1]
        b.OperandL = lhs

        if isUnaryExpr(idx+2) {
            b.OperandR, idx = prsUnaryExpr(idx+2)
        } else {
            b.OperandR, idx = prsValue(idx+2)
        }

        // TODO test later with parentheses expr
        for isBinaryExpr(idx) && getPrecedence(tokens[idx+1].Type) > precedence {
            if b.Operator.Type == token.Minus {
                b.Operator.Type = token.Plus
                b.Operator.Str = "+"

                t := token.Token{ Type: token.Minus, Str: "-", Pos: tokens[idx+1].Pos }
                b.OperandR = &ast.UnaryExpr{ Operator: t, Operand: b.OperandR }
            }

            // left to right as correct order of operations
            // OperandL has lower precedence so put it right (swap with OperandR)
            b.OperandL, idx = prsBinary(idx, b.OperandR, precedence + 1)
            b.OperandR = lhs
        }

        lhs = &b
    }

    return lhs, idx
}
