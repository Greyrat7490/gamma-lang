package prs

import (
    "gorec/ast"
    "gorec/token"
)

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

// https://en.wikipedia.org/wiki/Operator-precedence_parser
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
        b.OperandR, idx = prsLitExpr(idx+2)

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
