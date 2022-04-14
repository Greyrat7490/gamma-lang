package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/token"
    "gorec/func"
)

type OpExpr interface {
    Op
    expr()  // to differenciate OpExpr from OpStmt and OpDecl
}

type OpFnCall struct {
    FnName token.Token
    Values []string
}

func (o OpFnCall) expr() {}

func (o OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "
    return fmt.Sprintf("%sOP_CALL_FN:\n%s%s %v\n", s, s2, o.FnName.Str, o.Values)
}

func (o OpFnCall) Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
    fn.CallFunc(asm, o.FnName)
}
