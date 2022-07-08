package ast

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/func"
    "gorec/types"
    "gorec/token"
)


type Decl interface {
    Node
    Compile(file *os.File)
    decl()  // to distinguish Decl from Stmt
}

type BadDecl struct {}

type DecVar struct {
    Name token.Token
    Type types.Type
}

type DefVar struct {
    Name token.Token
    Type types.Type
    ColPos token.Pos
    Value Expr
}

type DefFn struct {
    FnName token.Token
    Args []DecVar
    Block Block
}


func (d *BadDecl) decl() {}
func (d *DecVar)  decl() {}
func (d *DefVar)  decl() {}
func (d *DefFn)   decl() {}


func (d *DecVar) Compile(file *os.File) {
    vars.Declare(d.Name, d.Type)
}

func (d *DefVar) Compile(file *os.File) {
    if d.Type == nil {
        d.Type = d.Value.GetType()
    }

    vars.Declare(d.Name, d.Type)

    d.typeCheck()

    v := vars.GetVar(d.Name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", d.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + d.Name.At())
        os.Exit(1)
    }

    switch e := d.Value.(type) {
    case *Lit:
        v.DefVal(file, e.Val)

    case *Ident:
        if ptr, ok := vars.GetVar(d.Name.Str).GetType().(types.PtrType); ok {
            otherType := vars.GetVar(e.Ident.Str).GetType()
            if ptr.BaseType != otherType {
                fmt.Fprintf(os.Stderr, "[ERROR] %s points to %v not should be %v\n", d.Name.Str, otherType, ptr.BaseType)
                fmt.Fprintln(os.Stderr, "\t" + d.Name.At())
                os.Exit(1)
            }

            vars.DefPtrWithVar(file, d.Name, e.Ident)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only define global pointer with another global var")
            fmt.Fprintln(os.Stderr, "\t" + d.Name.At())
            os.Exit(1)
        }

    default:
        d.Value.Compile(file)
        v.DefExpr(file)
    }
}

func (d *DefFn) Compile(file *os.File) {
    fn.Declare(d.FnName)

    vars.CreateScope()

    regIdx := 0

    fn.Define(file, d.FnName, argsSize(d.Args), d.Block.maxFrameSize())

    for _,a := range d.Args {
        fn.AddArg(a.Type)
        vars.Declare(a.Name, a.Type)
        fn.DefArg(file, regIdx, a.Type)

        if a.Type.GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    d.Block.Compile(file)

    vars.RemoveScope()
    fn.End(file);
}

func (d *BadDecl) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
}


func argsSize(args []DecVar) (res int) {
    for _,a := range args {
        res += a.Type.Size()
    }

    return res
}

func (d *Block) maxFrameSize() int {
    res := 0
    maxInnerSize := 0

    for _,s := range d.Stmts {
        switch stmt := s.(type) {
        case *DeclStmt:
            if def, ok := stmt.Decl.(*DefVar); ok {
                res += def.Type.Size()
            }
        case *Block:
            size := stmt.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *For:
            size := stmt.Def.Type.Size()
            size += stmt.Block.maxFrameSize()
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *While:
            size := stmt.Block.maxFrameSize()
            if stmt.Def != nil {
                size += stmt.Def.Type.Size()
            }
            if size > maxInnerSize {
                maxInnerSize = size
            }
        case *If:
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
