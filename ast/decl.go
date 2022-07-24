package ast

import (
    "os"
    "fmt"
    "gorec/types"
    "gorec/token"
    "gorec/identObj/func"
    "gorec/identObj/vars"
    "gorec/identObj/consts"
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
    Ident Ident
    Args []DecVar
    Block Block
}


func (d *DecVar) Compile(file *os.File) {}

func (d *DefConst) Compile(file *os.File) {
    d.typeCheck()

    if c,ok := d.Ident.Obj.(*consts.Const); ok {
        if c.Type == nil {
            c.Type = d.Value.GetType()
            d.Type = c.Type
        }

        val := d.Value.constEval()

        if val.Type == token.Unknown {
            fmt.Fprintln(os.Stderr, "[ERROR] cannot evaluate expr at compile time (not const)")
            fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
            os.Exit(1)
        }

        c.Define(val)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a const (in decl.go DefConst Compile)")
        os.Exit(1)
    }
}

func (d *DefVar) Compile(file *os.File) {
    if v,ok := d.Ident.Obj.(vars.Var); ok {
        if v.GetType() == nil {
            d.Type = d.Value.GetType()
            v.SetType(d.Type)
        }

        d.typeCheck()


        // compile time evaluation
        if val := d.Value.constEval(); val.Type != token.Unknown {
            v.DefVal(file, val)
            return
        }

        if _,ok := v.(*vars.GlobalVar); ok {
            fmt.Fprintln(os.Stderr, "[ERROR] defining a global variable with a non const expr is not allowed")
            fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
            os.Exit(1)
        }

        d.Value.Compile(file)
        vars.VarSetExpr(file, v)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a var (in decl.go DefVar Compile)")
        os.Exit(1)
    }
}

func (d *DefFn) Compile(file *os.File) {
    if f,ok := d.Ident.Obj.(*fn.Func); ok {
        f.Define(file)

        regIdx := 0
        for _,a := range d.Args {
            if v,ok := a.Ident.Obj.(vars.Var); ok {
                fn.DefArg(file, regIdx, v)
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] expected identObj of an argument to be a var (in decl.go DefFn Compile)")
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
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] expected identObj to be a func (in decl.go DefFn Compile)")
        os.Exit(1)
    }
}

func (d *BadDecl) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
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
