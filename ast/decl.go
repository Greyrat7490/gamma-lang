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
    TypePos token.Pos
}

type DefVar struct {
    Name token.Token
    Type types.Type
    ColPos token.Pos
    Value Expr
}

type DefConst struct {
    Name token.Token
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


func (d *DecVar) Compile(file *os.File) {
    vars.DecVar(d.Name, d.Type)
}

func (d *DefConst) Compile(file *os.File) {
    if d.Type == nil {
        d.Type = d.Value.GetType()
    }

    d.typeCheck()

    val := d.Value.constEval()

    if val.Type == token.Unknown {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot evaluate expr at compile time (not const)")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    vars.DefConst(d.Name, d.Type, val)
}

func (d *DefVar) Compile(file *os.File) {
    if d.Type == nil {
        d.Type = d.Value.GetType()
    }

    vars.DecVar(d.Name, d.Type)

    d.typeCheck()

    v := vars.GetVar(d.Name.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", d.Name.Str)
        fmt.Fprintln(os.Stderr, "\t" + d.Name.At())
        os.Exit(1)
    }

    // compile time evaluation
    if val := d.Value.constEval(); val.Type != token.Unknown {
        v.DefVal(file, val)
        return
    }

    if _,ok := d.Value.(*Ident); ok {
        fmt.Fprintln(os.Stderr, "[ERROR] defining a variable with another variable is not supported yet")
        fmt.Fprintln(os.Stderr, "\t" + d.Name.At())
        os.Exit(1)
    }

    d.Value.Compile(file)
    v.DefExpr(file)
}

func (d *DefFn) Compile(file *os.File) {
    fn.Declare(d.FnName)

    vars.CreateScope()

    regIdx := 0

    fn.Define(file, d.FnName, argsSize(d.Args), d.Block.maxFrameSize())

    for _,a := range d.Args {
        fn.AddArg(a.Type)
        vars.DecVar(a.Name, a.Type)
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
func (d *DecVar)   At() string { return d.Name.At() }
func (d *DefVar)   At() string { return d.ColPos.At() }
func (d *DefConst) At() string { return d.ColPos.At() }
func (d *DefFn)    At() string { return d.Pos.At() }

func (d *BadDecl)  End() string { return "" }
func (d *DecVar)   End() string { return d.TypePos.At() }
func (d *DefVar)   End() string { return d.Value.End() }
func (d *DefConst) End() string { return d.Value.End() }
func (d *DefFn)    End() string { return d.Block.End() }
