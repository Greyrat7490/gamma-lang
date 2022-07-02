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
    decl()  // to differenciate OpDecl from OpStmt
}

type BadDecl struct {}

type OpDecVar struct {
    Name token.Token
    Type types.Type
}

type OpDefVar struct {
    Name token.Token
    Type types.Type
    ColPos token.Pos
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
    vars.Declare(o.Name, o.Type)
}

func (o *OpDefVar) Compile(file *os.File) {
    if o.Type == nil {
        o.Type = o.Value.GetType()
    }
    
    vars.Declare(o.Name, o.Type)

    o.typeCheck()

    v := vars.GetVar(o.Name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", o.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.Name.At())
        os.Exit(1)
    }

    switch e := o.Value.(type) {
    case *LitExpr:
        v.DefVal(file, e.Val)

    case *IdentExpr:
        if ptr, ok := vars.GetVar(o.Name.Str).GetType().(types.PtrType); ok {
            otherType := vars.GetVar(e.Ident.Str).GetType()
            if ptr.BaseType != otherType {
                fmt.Fprintf(os.Stderr, "[ERROR] %s points to %v not should be %v\n", o.Name.Str, otherType, ptr.BaseType)
                fmt.Fprintln(os.Stderr, "\t" + o.Name.At())
                os.Exit(1)
            }

            vars.DefPtrWithVar(file, o.Name, e.Ident)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only define global pointer with another global var")
            fmt.Fprintln(os.Stderr, "\t" + o.Name.At())
            os.Exit(1)
        }

    default:
        o.Value.Compile(file)
        v.DefExpr(file)
    }
}

func (o *OpDefFn) Compile(file *os.File) {
    fn.Declare(o.FnName)

    vars.CreateScope()

    regIdx := 0

    fn.Define(file, o.FnName, argsSize(o.Args), o.Block.maxFrameSize())

    for _,a := range o.Args {
        fn.AddArg(a.Type)
        vars.Declare(a.Name, a.Type)
        fn.DefArg(file, regIdx, a.Type)

        if a.Type.GetKind() == types.Str {
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
        res += a.Type.Size()
    }

    return res
}

func (o *OpBlock) maxFrameSize() int {
    res := 0
    maxInnerSize := 0

    for _,s := range o.Stmts {
        switch stmt := s.(type) {
        case *OpDeclStmt:
            if def, ok := stmt.Decl.(*OpDefVar); ok {
                res += def.Type.Size()
            }
        case *OpBlock:
            size := stmt.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *ForStmt:
            size := stmt.Def.Type.Size()
            size += stmt.Block.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *WhileStmt:
            size := stmt.Block.maxFrameSize()
            if stmt.Def != nil {
                size += stmt.Def.Type.Size()
            }
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *IfStmt:
            size := stmt.Block.maxFrameSize()

            if stmt.Else != nil {
                size2 := stmt.Else.Block.maxFrameSize()
                if size < size2 {
                    size = size2
                }
            }

            if size > maxInnerSize {
                maxInnerSize = size
            }
        }
    }

    res += maxInnerSize

    return res
}
