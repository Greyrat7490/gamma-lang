package ast

import "gorec/token"

func (e *Lit)     constEval() string { return e.Val.Str }
func (e *Unary)   constEval() string { 
    v := e.Operand.constEval()

    if e.Operator.Type == token.Minus {
        return e.Operator.Str + v
    }
    
    if e.Operator.Type == token.Plus {
        return v
    }

    // TODO error
    
    return ""
}

// TODO
func (e *BadExpr) constEval() string { return "" }
func (e *FnCall)  constEval() string { return "" }
func (e *Ident)   constEval() string { return "" }
func (e *Binary)  constEval() string { return "" }
func (e *Paren)   constEval() string { return "" }
func (e *XSwitch) constEval() string { return "" }
func (e *XCase)   constEval() string { return "" }
