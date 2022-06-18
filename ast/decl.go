package ast

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/func"
    "gorec/types"
    "gorec/token"
)


type OpDecl interface {
    Op
    Compile(file *os.File)
    typeCheck()
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


func (o *OpDecVar) Compile(file *os.File) {
    vars.Declare(o.Varname, o.Vartype)
}

func (o *OpDefVar) Compile(file *os.File) {
    v := vars.GetVar(o.Varname.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", o.Varname.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.Varname.At())
        os.Exit(1)
    }

    switch e := o.Value.(type) {
    case *LitExpr:
        v.DefVal(file, e.Val)

    case *IdentExpr:
        if ptr, ok := vars.GetVar(o.Varname.Str).GetType().(types.PtrType); ok {
            otherType := vars.GetVar(e.Ident.Str).GetType()
            if ptr.BaseType != otherType {
                fmt.Fprintf(os.Stderr, "[ERROR] %s points to %v not should be %v\n", o.Varname.Str, otherType, ptr.BaseType)
                fmt.Fprintln(os.Stderr, "\t" + o.Varname.At())
                os.Exit(1)
            }

            vars.DefPtrWithVar(file, o.Varname, e.Ident)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only define global pointer with another global var")
            fmt.Fprintln(os.Stderr, "\t" + o.Varname.At())
            os.Exit(1)
        }

    default:
        o.Value.Compile(file)
        v.DefExpr(file)
    }
}

func (o *OpDefFn) Compile(file *os.File) {
    vars.CreateScope()

    fn.Define(file, o.FnName)
    fn.ReserveSpace(file, argsSize(o.Args), o.Block.maxFrameSize())
    regIdx := 0
    for _, a := range o.Args {
        fn.AddArg(a.Vartype)
        a.Compile(file)
        fn.DefArg(file, regIdx, a.Vartype)

        if a.Vartype.GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    o.Block.Compile(file)

    vars.RemoveScope()
    fn.End(file);
}

func (o *BadDecl) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
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
