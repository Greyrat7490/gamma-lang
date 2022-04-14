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
    expr()  // to differenciate OpExpr from OpStmt and OpDecl
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

func (o LitExpr)   expr() {}
func (o IdentExpr) expr() {}
func (o OpFnCall)  expr() {}


func (o OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "
    return fmt.Sprintf("%sOP_CALL_FN:\n%s%s %v\n", s, s2, o.FnName.Str, o.Values)
}

func (o LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%s)", o.Val.Str, o.Type.Readable())
}
func (o IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)"
}


func (o LitExpr)   Compile(asm *os.File) {}
func (o IdentExpr) Compile(asm *os.File) {}
func (o OpFnCall)  Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
    fn.CallFunc(asm, o.FnName)
}
