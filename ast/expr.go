package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/arithmetic"
)

type OpExpr interface {
    Op
    Compile(asm *os.File, dest token.Token)
    GetValue() token.Token
}

type BadExpr struct{}

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

// TODO: use always rax as dest
func (o *LitExpr)    Compile(asm *os.File, dest token.Token) {}
func (o *IdentExpr)  Compile(asm *os.File, dest token.Token) {}
func (o *UnaryExpr)  Compile(asm *os.File, dest token.Token) { /* TODO negating vars */ }
func (o *BinaryExpr) Compile(asm *os.File, dest token.Token) {
    o.OperandL.Compile(asm, dest)
    o.OperandR.Compile(asm, dest)

    arith.BinaryOp(asm, o.Operator.Type, o.OperandR.GetValue(), dest)
}

func (o *OpFnCall) Compile(asm *os.File, dest token.Token) {
    fn.DefineArgs(asm, o.FnName, o.Values)
    fn.CallFunc(asm, o.FnName)
}

func (o *BadExpr) Compile(asm *os.File, dest token.Token) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}


func (o *LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%s)\n", o.Val.Str, o.Type.Readable())
}

func (o *IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}

func (o *OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "
    return fmt.Sprintf("%sOP_CALL_FN:\n%s%s %v\n", s, s2, o.FnName.Str, o.Values)
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

func (o *BadExpr) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return ""
}


func (o *LitExpr)    GetValue() token.Token { return o.Val }
func (o *IdentExpr)  GetValue() token.Token { return o.Ident }
func (o *OpFnCall)   GetValue() token.Token { return token.Token{} }
func (o *BinaryExpr) GetValue() token.Token {
    // first literal of Operation with highest precedence
    // deepest left op (see prsBinary)
    return o.OperandL.GetValue()
}

func (o *UnaryExpr) GetValue() token.Token {
    if l, ok := o.Operand.(*LitExpr); ok {
        t := l.Val
        t.Str = o.Operator.Str + l.Val.Str
        return t
    } else {
        return o.Operand.GetValue()
    }
}

func (o *BadExpr) GetValue() token.Token {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return token.Token{}
}
