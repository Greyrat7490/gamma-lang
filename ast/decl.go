package ast

import (
    "os"
    "fmt"
    "gamma/types"
    "gamma/token"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)


type Decl interface {
    Node
    Compile(file *os.File)
    decl()  // to distinguish Decl from Stmt
}

type BadDecl struct {}

type DecVar struct {
    V vars.Var          // TODO remove Var,Const,Func
    Type types.Type
    TypePos token.Pos
}

type DefVar struct {
    V vars.Var
    Type types.Type     // nil -> infer type
    ColPos token.Pos
    Value Expr
}

type DefConst struct {
    C *consts.Const
    Type types.Type     // nil -> infer type
    ColPos token.Pos
    Value Expr
}

type DefFn struct {
    F *fn.Func
    Pos token.Pos
    Args []DecVar
    RetType types.Type
    Block Block
}

type DefStruct struct {
    S *structDec.Struct
    Pos token.Pos
    Name token.Token
    BraceLPos token.Pos
    Fields []DecVar
    BraceRPos token.Pos
}


func (d *DecVar) Compile(file *os.File) {}

func (d *DefConst) Compile(file *os.File) {
    d.typeCheck()

    if d.C.GetType() == nil {
        d.Type = d.Value.GetType()
        d.C.SetType(d.Type)
    }

    val := d.Value.ConstEval()

    if val.Type == token.Unknown {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot evaluate expr at compile time (not const)")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    d.C.Define(val)
}

func (d *DefVar) Compile(file *os.File) {
    if d.V.GetType() == nil {
        d.Type = d.Value.GetType()
        d.V.SetType(d.Type)
    }

    d.typeCheck()


    // compile time evaluation
    if val := d.Value.ConstEval(); val.Type != token.Unknown {
        d.V.DefVal(file, val)
        return
    }

    if _,ok := d.V.(*vars.GlobalVar); ok {
        fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    d.Value.Compile(file)
    vars.VarSetExpr(file, d.V)
}

func (d *DefFn) Compile(file *os.File) {
    d.F.Define(file)

    regIdx := 0
    for _,a := range d.Args {
        if fn.DefArg(file, regIdx, a.V) {
            switch t := a.Type.(type) {
            case types.StrType:
                regIdx += 2

            case types.StructType:
                regIdx += len(t.Types)

            default:
                regIdx++
            }
        }
    }

    d.Block.Compile(file)

    fn.End(file);
}

func (d *DefStruct) Compile(file *os.File) {}

func (d *BadDecl) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
}


func (d *BadDecl)   decl() {}
func (d *DecVar)    decl() {}
func (d *DefVar)    decl() {}
func (d *DefConst)  decl() {}
func (d *DefFn)     decl() {}
func (d *DefStruct) decl() {}

func (d *BadDecl)   At() string { return "" }
func (d *DecVar)    At() string { return d.V.GetPos().At() }
func (d *DefVar)    At() string { return d.ColPos.At() }
func (d *DefConst)  At() string { return d.ColPos.At() }
func (d *DefFn)     At() string { return d.Pos.At() }
func (d *DefStruct) At() string { return d.Pos.At() }

func (d *BadDecl)   End() string { return "" }
func (d *DecVar)    End() string { return d.TypePos.At() }
func (d *DefVar)    End() string { return d.Value.End() }
func (d *DefConst)  End() string { return d.Value.End() }
func (d *DefFn)     End() string { return d.Block.End() }
func (d *DefStruct) End() string { return d.BraceRPos.At() }
