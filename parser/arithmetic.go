package prs

import (
    "fmt"
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
        // 2
        // 1
        // 2
        // -> 2 regs needed
        fmt.Println(precedence)

        var b ast.BinaryExpr
        b.Operator = tokens[idx+1]
        b.OperandR, idx = prsLitExpr(idx+2)

        for isBinaryExpr(idx) && getPrecedence(tokens[idx+1].Type) > precedence {
            //
            if b.Operator.Type == token.Minus {
                b.Operator.Type = token.Plus
                b.Operator.Str = "+"

                t := token.Token{ Type: token.Minus, Str: "-", Pos: lhs.GetValue().Pos }
                b.OperandR = &ast.UnaryExpr{ Operator: t, Operand: b.OperandR }
            }

            b.OperandR, idx = prsBinary(idx, b.OperandR, precedence + 1)
        }

        b.OperandL = lhs

        lhs = &b
    }

    return lhs, idx
}
