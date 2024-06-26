package ast

import (
    "os"
    "fmt"
    "strings"
    "gamma/types"
    "gamma/token"
    "gamma/ast/identObj"
    "gamma/ast/identObj/vars"
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

type DecField struct {
    Name token.Token
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
    C *identObj.Const
    Type types.Type
    ColPos token.Pos
    Value Expr
}

type DefFn struct {
    Pos token.Pos
    FnHead FnHead 
    Block Block
}

type FnHead struct {
    F *identObj.Func
    Generic *identObj.Generic  // can be nil
    Name token.Token
    Args []DecVar
    RetType types.Type
    IsConst bool
}

type DefStruct struct {
    S *identObj.Struct
    Pos token.Pos
    Name token.Token
    Generic token.Token // empty if IsGeneric == false
    BraceLPos token.Pos
    Fields []DecField
    BraceRPos token.Pos
}

type DefInterface struct {
    I *identObj.Interface
    Pos token.Pos
    Name token.Token
    Generic token.Token // empty if IsGeneric == false
    BraceLPos token.Pos
    FnHeads []FnHead
    BraceRPos token.Pos
}

type DefEnum struct {
    E *identObj.Enum
    Pos token.Pos
    Name token.Token
    Generic token.Token // empty if IsGeneric == false
    IdType types.Type
    BraceLPos token.Pos
    Elems []EnumElem
    BraceRPos token.Pos
}
type EnumElem struct {
    Name token.Token
    Type *EnumElemType     // can be nil
}
type EnumElemType struct {
    ParenL token.Pos
    Type types.Type
    ParenR token.Pos
}

type Impl struct {
    Impl identObj.Impl
    Pos token.Pos
    BraceLPos token.Pos
    FnDefs []DefFn
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

func (d *DecField) Readable(indent int) string {
    s := strings.Repeat("   ", indent+1)

    return strings.Repeat("   ", indent) + "DEC_FIELD:\n" +
        s + fmt.Sprintf("%v(Name)\n", d.Name.Str) +
        s + fmt.Sprintf("%v(Typename)\n", d.Type)
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

func (o *FnHead) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "FN_HEAD:\n"

    args := ""
    for _,a := range o.Args {
        args += fmt.Sprintf("%s(Name) %v(Type), ", a.V.GetName(), a.Type)
    }
    if len(args) > 0 { args = args[:len(args)-2] }

    s := strings.Repeat("   ", indent+1)

    generic := ""
    if o.F.IsGeneric() {
        generic = fmt.Sprintf("%sGeneric: %s\n", s, o.F.GetGeneric().Name)
    }

    res += fmt.Sprintf("%sName: %s\n%s%sArgs: [%s]\n", s, o.Name, generic, s, args)

    if o.RetType != nil {
        res += fmt.Sprintf("%sRet: %v\n", s, o.RetType)
    }

    return res + fmt.Sprintf("%sIsConst: %t\n", s, o.IsConst)
}

func (o *DefFn) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "DEF_FN:\n" + 
        o.FnHead.Readable(indent+1) +
        o.Block.Readable(indent+1)
}

func (o *DefStruct) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "DEF_STRUCT:\n" +
        strings.Repeat("   ", indent+1) + o.Name.String() + "\n"

    if o.Generic.Type != 0 {
        res += fmt.Sprintf("%sGeneric: %s\n", strings.Repeat("   ", indent+1), o.Generic.Str)
    }

    for _,f := range o.Fields {
        res += f.Readable(indent+1)
    }

    return res
}

func (o *DefInterface) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "DEF_INTERFACE:\n" +
        strings.Repeat("   ", indent+1) + o.Name.String() + "\n"

    if o.Generic.Type != 0 {
        res += fmt.Sprintf("%sGeneric: %s\n", strings.Repeat("   ", indent+1), o.Generic.Str)
    }

    for _,f := range o.FnHeads {
        res += f.Readable(indent+1)
    }

    return res
}

func (d *DefEnum) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "DEF_ENUM:\n" +
        strings.Repeat("   ", indent+1) + d.IdType.String() + "\n" +
        strings.Repeat("   ", indent+1) + d.Name.String() + "\n"

    if d.Generic.Type != 0 {
        res += fmt.Sprintf("%sGeneric: %s\n", strings.Repeat("   ", indent+1), d.Generic.Str)
    }

    for _,e := range d.Elems {
        res += e.Readable(indent+1)
    }

    return res
}
func (d *EnumElem) Readable(indent int) string {
    s := strings.Repeat("   ", indent+1)
    
    res := strings.Repeat("   ", indent) + "ENUM_ELEM:\n" +
        s + fmt.Sprintf("%v(Name)\n", d.Name.Str)

    if d.Type != nil {
        res += s + fmt.Sprintf("%v(Type)\n", d.Type.Type)
    }

    return res
}

func (d *Impl) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "IMPL:\n"

    if d.Impl.GetGeneric() != nil {
        res += fmt.Sprintf("%sGeneric: %s\n", strings.Repeat("   ", indent+1), d.Impl.GetGeneric().Typ.Name)
    }

    res += strings.Repeat("   ", indent+1) + "Interface: " + d.Impl.GetInterfaceName() + "\n" +
        strings.Repeat("   ", indent+1) + "DestType: " + d.Impl.GetDstType().String() + "\n"

    for _,f := range d.FnDefs {
        res += f.Readable(indent+1)
    }

    return res
}

func (d *Import) Readable(indent int) string {
    res := strings.Repeat("   ", indent) + "IMPORT:\n" +
        strings.Repeat("   ", indent+1) + d.Path.Str + "\n"

    if d.Path.Str != "\"std.gma\"" {
        for _,d := range d.Decls {
            res += d.Readable(indent+1)
        }
    }

    return res
}

func (o *BadDecl) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad declaration")
    os.Exit(1)
    return ""
}


func (d *BadDecl)       decl() {}
func (d *DecVar)        decl() {}
func (d *DefVar)        decl() {}
func (d *DefConst)      decl() {}
func (d *DefFn)         decl() {}
func (d *DefStruct)     decl() {}
func (d *DefInterface)  decl() {}
func (d *DefEnum)       decl() {}
func (d *Impl)          decl() {}
func (d *DecField)      decl() {}
func (d *FnHead)        decl() {}
func (d *Import)        decl() {}

func (d *BadDecl)       At() string { return "" }
func (d *DecVar)        At() string { return d.V.GetPos().At() }
func (d *DefVar)        At() string { return d.ColPos.At() }
func (d *DefConst)      At() string { return d.ColPos.At() }
func (d *DefFn)         At() string { return d.Pos.At() }
func (d *DefStruct)     At() string { return d.Pos.At() }
func (d *DefInterface)  At() string { return d.Pos.At() }
func (d *DefEnum)       At() string { return d.Pos.At() }
func (d *Impl)          At() string { return d.Pos.At() }
func (d *DecField)      At() string { return d.Name.At() }
func (d *Import)        At() string { return d.Pos.At() }

func (d *BadDecl)       End() string { return "" }
func (d *DecVar)        End() string { return d.TypePos.At() }
func (d *DefVar)        End() string { return d.Value.End() }
func (d *DefConst)      End() string { return d.Value.End() }
func (d *DefFn)         End() string { return d.Block.End() }
func (d *DefStruct)     End() string { return d.BraceRPos.At() }
func (d *DefInterface)  End() string { return d.BraceRPos.At() }
func (d *DefEnum)       End() string { return d.BraceRPos.At() }
func (d *Impl)          End() string { return d.BraceRPos.At() }
func (d *DecField)      End() string { return d.TypePos.At() }
func (d *Import)        End() string { return d.Path.At() }
