package ast

import (
    "os"
    "fmt"
    "strings"
    "gamma/types"
    "gamma/token"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)


type Decl interface {
    Node
    decl()  // to distinguish Decl from Stmt and Expr
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

type Import struct {
    Pos token.Pos
    Path token.Token
    Decls []Decl
}


func (o *DecVar) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    return s + "DEC_VAR:\n" +
          s2 + fmt.Sprintf("%s(Name)\n", o.V.GetName()) +
          s2 + fmt.Sprintf("%v(Typename)\n", o.Type)
}

func (o *DefVar) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    res := s + "DEF_VAR:\n" +
        s2 + fmt.Sprintf("%v(Name)\n", o.V.GetName())

    if o.Type == nil {
        res += s2 + "infer type\n"
    } else {
        res += s2 + fmt.Sprintf("%v(Typename)\n", o.Type)
    }

    return res + o.Value.Readable(indent+1)
}

func (o *DefConst) Readable(indent int) string {
    s  := strings.Repeat("   ", indent)
    s2 := strings.Repeat("   ", indent+1)

    res := s + "DEF_CONST:\n" +
        s2 + fmt.Sprintf("%s(Name)\n", o.C.GetName())

    if o.Type == nil {
        res += s2 + "infer type\n"
    } else {
        res += s2 + fmt.Sprintf("%v(Typename)\n", o.Type)
    }

    return res + o.Value.Readable(indent+1)
}

func (o *DefFn) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "DEF_FN:\n"

    args := ""
    for _,a := range o.Args {
        args += fmt.Sprintf("%s(Name) %v(Type), ", a.V.GetName(), a.Type)
    }
    if len(args) > 0 { args = args[:len(args)-2] }

    s := strings.Repeat("   ", indent+1)

    res += fmt.Sprintf("%sName: %s\n", s, o.F.GetName()) +
        fmt.Sprintf("%sArgs: [%s]\n", s, args)

    if o.RetType != nil {
        res += fmt.Sprintf("%sRet: %v\n", s, o.RetType)
    }

    return res + o.Block.Readable(indent+2)
}

func (o *DefStruct) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "DEF_STRUCT:\n" +
        strings.Repeat("   ", indent+1) + o.Name.String() + "\n"

    for _,f := range o.Fields {
        res += f.Readable(indent+1)
    }

    return res
}

func (d *Import) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "IMPORT:\n" +
        strings.Repeat("   ", indent+1) + d.Path.Str + "\n"
    for _,d := range d.Decls {
        res += d.Readable(indent+1)
    }

    return res
}

func (o *BadDecl) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
    return ""
}


func (d *BadDecl)   decl() {}
func (d *DecVar)    decl() {}
func (d *DefVar)    decl() {}
func (d *DefConst)  decl() {}
func (d *DefFn)     decl() {}
func (d *DefStruct) decl() {}
func (d *Import)    decl() {}

func (d *BadDecl)   At() string { return "" }
func (d *DecVar)    At() string { return d.V.GetPos().At() }
func (d *DefVar)    At() string { return d.ColPos.At() }
func (d *DefConst)  At() string { return d.ColPos.At() }
func (d *DefFn)     At() string { return d.Pos.At() }
func (d *DefStruct) At() string { return d.Pos.At() }
func (d *Import)    At() string { return d.Pos.At() }

func (d *BadDecl)   End() string { return "" }
func (d *DecVar)    End() string { return d.TypePos.At() }
func (d *DefVar)    End() string { return d.Value.End() }
func (d *DefConst)  End() string { return d.Value.End() }
func (d *DefFn)     End() string { return d.Block.End() }
func (d *DefStruct) End() string { return d.BraceRPos.At() }
func (d *Import)    End() string { return d.Path.At() }
