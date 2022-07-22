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
    Ident Ident
    Type types.Type
    TypePos token.Pos
}

type DefVar struct {
    Ident Ident
    Type types.Type
    ColPos token.Pos
    Value Expr
}

type DefConst struct {
    Ident Ident
    Type types.Type
    ColPos token.Pos
    Value Expr
}

type DefFn struct {
    Pos token.Pos
    FnName token.Token
    Args []DecVar
    Block Block
}


func (d *DecVar) Compile(file *os.File) {}

func (d *DefConst) Compile(file *os.File) {
    if d.Ident.C.Type == nil {
        d.Ident.C.Type = d.Value.GetType()
        d.Type = d.Ident.C.Type
    }

    d.typeCheck()

    val := d.Value.constEval()

    if val.Type == token.Unknown {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot evaluate expr at compile time (not const)")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    d.Ident.C.Define(val)
}

func (d *DefVar) Compile(file *os.File) {
    if d.Ident.V.GetType() == nil {
        if g,ok := d.Ident.V.(*vars.GlobalVar); ok {
            g.Type = d.Value.GetType()
            d.Type = g.Type
        } else if l,ok := d.Ident.V.(*vars.LocalVar); ok {
            l.Type = d.Value.GetType()
            d.Type = l.Type
        }
    }

    d.typeCheck()


    // compile time evaluation
    if val := d.Value.constEval(); val.Type != token.Unknown {
        d.Ident.V.DefVal(file, val)
        return
    }

    if _,ok := d.Ident.V.(*vars.GlobalVar); ok {
        fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    d.Value.Compile(file)
    d.Ident.V.DefExpr(file)
}

func (d *DefFn) Compile(file *os.File) {
    fn.Define(file, d.FnName, argsSize(d.Args), d.Block.maxFrameSize())

    regIdx := 0
    for _,a := range d.Args {
        if v,ok := a.Ident.V.(*vars.LocalVar); ok {
            fn.DefArg(file, regIdx, v)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] decl.go Compile DefFn: (unreachable) expected argument to be local var")
            os.Exit(1)
        }

        if a.Type.GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    d.Block.Compile(file)

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
                if def.Type == nil {
                    def.Type = def.Value.GetType()
                }
                res += def.Type.Size()
            }
            if def, ok := stmt.Decl.(*DefConst); ok {
                if def.Type == nil {
                    def.Type = def.Value.GetType()
                }
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


func (d *BadDecl)  decl() {}
func (d *DecVar)   decl() {}
func (d *DefVar)   decl() {}
func (d *DefConst) decl() {}
func (d *DefFn)    decl() {}

func (d *BadDecl)  At() string { return "" }
func (d *DecVar)   At() string { return d.Ident.Ident.At() }
func (d *DefVar)   At() string { return d.ColPos.At() }
func (d *DefConst) At() string { return d.ColPos.At() }
func (d *DefFn)    At() string { return d.Pos.At() }

func (d *BadDecl)  End() string { return "" }
func (d *DecVar)   End() string { return d.TypePos.At() }
func (d *DefVar)   End() string { return d.Value.End() }
func (d *DefConst) End() string { return d.Value.End() }
func (d *DefFn)    End() string { return d.Block.End() }
