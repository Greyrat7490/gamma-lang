package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/types"
    "gorec/token"
    "gorec/vars"
    "gorec/func"
)

type OpDecl interface {
    Op
    decl()  // to differenciate OpDecl from OpStmt and OpExpr
}

type OpDecVar struct {
    Varname token.Token
    Vartype types.Type
}

type OpDefVar struct {
    Varname token.Token
    Value OpExpr
}

type OpDefFn struct {
    FnName token.Token
    Args []fn.Arg
    Block OpBlock
}


func (o *OpDecVar) decl() {}
func (o *OpDefVar) decl() {}
func (o *OpDefFn)  decl() {}


func (o *OpDecVar) Compile(asm *os.File) {
    vars.Declare(o.Varname, o.Vartype)
}

func (o *OpDefVar) Compile(asm *os.File) {
    o.Value.Compile(asm)
    value := o.Value.GetValue()

    vars.Define(asm, o.Varname, value)
}

func (o *OpDefFn) Compile(asm *os.File) {
    fn.Define(asm, o.FnName)
    fn.DeclareArgs(o.Args)

    o.Block.Compile(asm)

    fn.End(asm);
}


func (o *OpDecVar) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_DEC_VAR:\n%s%s(%s) %s(Typename)\n", s, s2,
        o.Varname.Str, o.Varname.Type.Readable(),
        o.Vartype.Readable())
}

func (o *OpDefVar) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_DEF_VAR:\n%s%s(%s)\n", s, s2,
        o.Varname.Str, o.Varname.Type.Readable()) + o.Value.Readable(indent+1)
}

func (o *OpDefFn) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    res := fmt.Sprintf("%sOP_DEF_FN:\n%s%s(%s) %v\n", s, s2,
        o.FnName.Str, o.FnName.Type.Readable(), o.Args) +
        o.Block.Readable(indent+2)

    return res
}
