package ast

import (
    "os"
    "fmt"
    "gamma/types"
    "gamma/token"
    "gamma/asm/x86_64"
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
    V vars.Var
    Type types.Type
    TypePos token.Pos
}

type DefVar struct {
    V vars.Var
    Type types.Type
    ColPos token.Pos
    Value Expr
}

type DefConst struct {
    C *consts.Const
    Type types.Type
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

    val := d.Value.ConstEval()

    if val.Type == token.Unknown {
        fmt.Fprintln(os.Stderr, "[ERROR] cannot evaluate expr at compile time (not const)")
        fmt.Fprintln(os.Stderr, "\t" + d.Value.At())
        os.Exit(1)
    }

    d.C.Define(val)
}

func (d *DefVar) Compile(file *os.File) {
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

    if c,ok := d.Value.(*FnCall); ok {
        if types.IsBigStruct(c.F.GetRetType()) {
            file.WriteString(fmt.Sprintf("lea rdi, [%s]\n", d.V.Addr(0)))
        }
    }

    d.Value.Compile(file)
    if !types.IsBigStruct(d.Value.GetType()) {
        vars.VarSetExpr(file, d.V)
    }
}

func (d *DefFn) Compile(file *os.File) {
    d.F.Define(file)

    regIdx := uint(0)

    if types.IsBigStruct(d.F.GetRetType()) {
        asm.MovDerefReg(file, fmt.Sprintf("rbp-%d", types.Ptr_Size), types.Ptr_Size, asm.RegDi)
        regIdx++
    }

    for _,a := range d.Args {
        if !types.IsBigStruct(a.V.GetType()) {
            i := types.RegCount(a.Type)

            if regIdx+i <= 6 {
                fn.DefArg(file, regIdx, a.V)
                regIdx += i
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
