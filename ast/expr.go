package ast

import (
    "fmt"
    "gorec/arithmetic"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "os"
    "strings"
)

type OpExpr interface {
    Op
    Compile(asm *os.File, dest token.Token)
    GetValue() token.Token
}
type OpFnCall struct {
    FnName token.Token
    Values []string
}
type LitExpr struct {
    Val token.Token
    Type types.Type
}
type IdentExpr struct {
    Ident token.Token
}
type UnaryExpr struct {
    Operator token.Token
    Operand OpExpr
}
type BinaryExpr struct {
    OperandL OpExpr
    Operator token.Token
    OperandR OpExpr
}

func (o *LitExpr)    GetValue() token.Token { return o.Val }
func (o *IdentExpr)  GetValue() token.Token { return o.Ident }
func (o *OpFnCall)   GetValue() token.Token { return token.Token{} }
func (o *BinaryExpr) GetValue() token.Token {
    // deepest left side literal value (in work)
    // should be left literal value of the first mul/div expr
    if _, ok := o.OperandL.(*LitExpr); ok {
        if _, ok := o.OperandR.(*LitExpr); !ok {
            return o.OperandR.GetValue()
        }
    }

    return o.OperandL.GetValue()
}
func (o *UnaryExpr)  GetValue() token.Token {
    if l, ok := o.Operand.(*LitExpr); ok {
        l.Val.Str = o.Operator.Str + l.Val.Str
        return l.Val
    } else {
        return o.Operand.GetValue()
    }
}


func (o *OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "
    return fmt.Sprintf("%sOP_CALL_FN:\n%s%s %v\n", s, s2, o.FnName.Str, o.Values)
}
func (o *LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%s)\n", o.Val.Str, o.Type.Readable())
}
func (o *IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}
func (o *UnaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_UNARY:\n%s%s(%s)\n", s, s2, o.Operator.Str, o.Operator.Type.Readable()) +
        o.Operand.Readable(indent+1)
}
func (o *BinaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return s + "OP_BINARY:\n" +
        o.OperandL.Readable(indent+1) +
        s2 + o.Operator.Str + o.Operator.Type.Readable() + "\n" +
        o.OperandR.Readable(indent+1)
}


func (o *LitExpr)    Compile(asm *os.File, dest token.Token) {}
func (o *IdentExpr)  Compile(asm *os.File, dest token.Token) {}
func (o *UnaryExpr)  Compile(asm *os.File, dest token.Token) { /* TODO negating vars */ }
func (o *BinaryExpr) Compile(asm *os.File, dest token.Token) {
    o.OperandL.Compile(asm, dest)
    o.OperandR.Compile(asm, dest)

    if l, ok := o.OperandL.(*LitExpr); ok {
        if r, ok := o.OperandR.(*LitExpr); ok {
            arith.BinaryOp(asm, o.Operator.Type, r.Val, dest)
        } else {
            arith.BinaryOp(asm, o.Operator.Type, l.Val, dest)
        }
    } else if r, ok := o.OperandR.(*LitExpr); ok {
        arith.BinaryOp(asm, o.Operator.Type, r.Val, dest)
    }
}
func (o *OpFnCall) Compile(asm *os.File, dest token.Token) {
    fn.DefineArgs(asm, o.FnName, o.Values)
    fn.CallFunc(asm, o.FnName)
}
