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
    Compile(asm *os.File)
    decl()  // to differenciate OpDecl from OpStmt and OpExpr
}

type BadDecl struct {}

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
    Args []OpDecVar
    Block OpBlock
}


func (o *OpDecVar) decl() {}
func (o *OpDefVar) decl() {}
func (o *OpDefFn)  decl() {}
func (o *BadDecl)  decl() {}


func (o *OpDecVar) Compile(asm *os.File) {
    vars.Declare(o.Varname, o.Vartype)
}

func (o *OpDefVar) Compile(asm *os.File) {
    if l, ok := o.Value.(*LitExpr); ok {
        vars.DefWithVal(asm, o.Varname, l.Val)
    } else if _, ok := o.Value.(*IdentExpr); ok {
        fmt.Fprintf(os.Stderr, "[ERROR] you cannot define a global var with another var(yet)")
        fmt.Fprintln(os.Stderr, "\t" + o.Varname.At())
        os.Exit(1)
    } else {
        o.Value.Compile(asm)
        vars.DefWithExpr(asm, o.Varname, "rax")
    }
}

func (o *OpDefFn) Compile(asm *os.File) {
    vars.CreateScope()

    fn.Define(asm, o.FnName)
    fn.ReserveSpace(asm, argsSize(o.Args), o.Block.maxFrameSize())
    for i, a := range o.Args {
        fn.AddArg(a.Vartype)
        a.Compile(asm)
        fn.DefArg(asm, i, a.Varname, a.Vartype)
    }

    o.Block.Compile(asm)

    vars.RemoveScope()
    fn.End(asm);
}

func (o *BadDecl) Compile(asm *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
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
    res := strings.Repeat("   ", indent) + "OP_DEF_FN:\n"

    s := ""
    for _,a := range o.Args {
        s += fmt.Sprintf("%s(Name) %s(Typename), ", a.Varname.Str, a.Vartype.Readable())
    }
    if len(s) > 0 { s = s[:len(s)-2] }

    res += fmt.Sprintf("%s%s(%s) [%s]\n", strings.Repeat("   ", indent+1), o.FnName.Str, o.FnName.Type.Readable(), s) +
        o.Block.Readable(indent+2)

    return res
}
func (o *BadDecl) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
    return ""
}


func argsSize(args []OpDecVar) (res int) {
    for _,a := range args {
        res += a.Vartype.Size()
    }

    return res
}

func (o *OpBlock) maxFrameSize() int {
    res := 0
    maxInnerSize := 0

    for _,s := range o.Stmts {
        switch stmt := s.(type) {
        case *OpDeclStmt:
            if dec, ok := stmt.Decl.(*OpDecVar); ok {
                res += dec.Vartype.Size()
            }
        case *OpBlock:
            size := stmt.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *ForStmt:
            size := stmt.Dec.Vartype.Size()
            size += stmt.Block.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *WhileStmt:
            size := stmt.Block.maxFrameSize()
            if stmt.InitVal != nil {
                size += stmt.Dec.Vartype.Size()
            }
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *IfStmt:
            size := stmt.Block.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *IfElseStmt:
            size := stmt.Block.maxFrameSize()
            size2 := stmt.If.Block.maxFrameSize()
            if size < size2 {
                size = size2
            }

            if size > maxInnerSize {
                maxInnerSize = size
            }
        }
    }

    res += maxInnerSize

    return res
}
