package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/func"
    "gorec/token"
    "gorec/types"
)

type OpExpr interface {
    Op
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

func (o *LitExpr)   GetValue() token.Token { return o.Val }
func (o *IdentExpr) GetValue() token.Token { return o.Ident }
func (o *UnaryExpr) GetValue() token.Token { return o.Operand.GetValue() }
func (o *OpFnCall)  GetValue() token.Token { return o.FnName }


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

func (o *LitExpr)   Compile(asm *os.File) {}
func (o *IdentExpr) Compile(asm *os.File) {}
func (o *UnaryExpr) Compile(asm *os.File) {
    if o.Operator.Type == token.Minus {
        switch e := o.Operand.(type) {
        case *LitExpr:
            e.Val.Str = o.Operator.Str + e.Val.Str
        case *IdentExpr:
            // TODO: negating variables
        default:
            os.Exit(1)
        }
    }
}
func (o *OpFnCall) Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
    fn.CallFunc(asm, o.FnName)
}
