package ast

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/asm/x86_64"
    "gamma/asm/x86_64/conditions"
    "gamma/ast/identObj"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)

type Expr interface {
    Node
    Compile(file *os.File)
    GetType() types.Type
    typeCheck()
    ConstEval() token.Token
}

type BadExpr struct{}

type FnCall struct {
    Ident Ident
    ParenLPos token.Pos
    Values []Expr
    ParenRPos token.Pos
}

type Lit struct {
    Val token.Token
    Type types.Type
}

type ArrayLit struct {
    Idx int
    Pos token.Pos
    Type types.ArrType
    BraceLPos token.Pos
    Values []Expr
    BraceRPos token.Pos
}

type StructLit struct {
    Idx int
    Pos token.Pos
    StructType types.StructType
    BraceLPos token.Pos
    Fields []FieldLit
    BraceRPos token.Pos
}

type FieldLit struct {
    Name token.Token
    Pos token.Pos
    Value Expr
}

type Indexed struct {
    ArrExpr Expr
    BrackLPos token.Pos
    Indices []Expr
    BrackRPos token.Pos
}

type Field struct {
    Pos token.Pos
    Obj identObj.IdentObj
    DotPos token.Pos
    FieldName token.Token
}

type Ident struct {
    Name string
    Pos token.Pos
    Obj identObj.IdentObj
}

type Unary struct {
    Operator token.Token
    Operand Expr
}

type Binary struct {
    Pos token.Pos
    OperandL Expr
    Operator token.Token
    OperandR Expr
}

type Paren struct {
    ParenLPos token.Pos
    Expr Expr
    ParenRPos token.Pos
}

type XSwitch struct {
    Pos token.Pos
    BraceLPos token.Pos
    Cases []XCase
    BraceRPos token.Pos
}

type XCase struct {
    Cond Expr
    ColonPos token.Pos
    Expr Expr
}


func (e *Lit) Compile(file *os.File) {
    switch e.Val.Type {
    case token.Str:
        strIdx := str.Add(e.Val)

        asm.MovRegVal(file, asm.RegA, types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovRegVal(file, asm.RegB, types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx)))

    case token.Boolean:
        if e.Val.Str == "true" { e.Val.Str = "1" } else { e.Val.Str = "0" }
        fallthrough

    default:
        asm.MovRegVal(file, asm.RegA, e.Type.Size(), e.Val.Str)
    }
}

func (e *StructLit) Compile(file *os.File) {
    for i := 0; i < len(e.Fields); i++ {
        l := e.Fields[i].Value
        switch l.GetType().GetKind() {
        case types.Str:
            strIdx := str.Add(l.ConstEval())

            asm.MovRegVal(file, uint8(i), types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
            i++
            asm.MovRegVal(file, uint8(i), types.I32_Size, fmt.Sprintf("%d", str.GetSize(strIdx)))

        case types.Bool:
            if l.ConstEval().Str == "true" {
                asm.MovRegVal(file, uint8(i), types.Bool_Size, "1")
            } else {
                asm.MovRegVal(file, uint8(i), types.Bool_Size, "0")
            }

        default:
            asm.MovRegVal(file, uint8(i), l.GetType().Size(), l.ConstEval().Str)
        }
    }
}

func (e *FieldLit) Compile(file *os.File) {}

func (e *Indexed) flatten() Expr {
    // for dim = 1 (no need to flatten)
    if len(e.Indices) == 1 {
        return e.Indices[0]
    }

    res := Binary{ Operator: token.Token{ Str: "+", Type: token.Plus }, OperandR: e.Indices[len(e.Indices)-1] }
    expr := &res.OperandL

    // for dim > 3
    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    var innerLen uint64 = arrType.Lens[0]
    for i := 1; i < len(e.Indices)-1; i++ {
        b := Binary{
            Operator: token.Token{ Str: "+", Type: token.Plus }, OperandR: &Binary{
                Operator: token.Token{ Str: "*", Type: token.Mul },
                OperandL: e.Indices[len(e.Indices)-1-i],
                OperandR: &Lit{ Val: token.Token{ Str: fmt.Sprint(innerLen), Type: token.Number }, Type: types.I32Type{} },
        } }

        *expr = &b
        expr = &b.OperandL

        innerLen *= arrType.Lens[i]
    }

    *expr = &Binary{
        Operator: token.Token{ Str: "*", Type: token.Mul },
        OperandL: e.Indices[0],
        OperandR: &Lit{ Val: token.Token{ Str: fmt.Sprint(innerLen), Type: token.Number }, Type: types.I32Type{} },
    }

    return &res
}

func (e *Indexed) AddrToRdx(file *os.File) {
    e.typeCheck()

    e.ArrExpr.Compile(file)

    arrType,_ := e.ArrExpr.GetType().(types.ArrType)
    baseTypeSize := uint64(arrType.Ptr.BaseType.Size())

    if len(arrType.Lens) < len(e.Indices){
        fmt.Fprintf(os.Stderr, "[ERROR] dimension of the array is %d but got %d\n", len(arrType.Lens), len(e.Indices))
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    idxExpr := e.flatten()
    val := idxExpr.ConstEval()
    if val.Type != token.Unknown {
        if val.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
            fmt.Fprintln(os.Stderr, "\t" + idxExpr.At())
            os.Exit(1)
        }

        idx,_ := strconv.ParseUint(val.Str, 10, 64)
        file.WriteString(fmt.Sprintf("lea rdx, [rax+%d]\n", idx * baseTypeSize))
    } else {
        asm.MovRegReg(file, asm.RegD, asm.RegA, types.Ptr_Size)
        idxExpr.Compile(file)

        asm.Mul(file, fmt.Sprint(baseTypeSize), types.Ptr_Size)
        asm.Add(file, asm.GetReg(asm.RegD, types.Ptr_Size), types.Ptr_Size)

        asm.MovRegReg(file, asm.RegD, asm.RegA, types.Ptr_Size)
    }
}

func (e *Indexed) Compile(file *os.File) {
    e.typeCheck()

    arrType,_ := e.ArrExpr.GetType().(types.ArrType)

    e.AddrToRdx(file)

    if arrType.Ptr.BaseType.GetKind() == types.Str {
        asm.MovRegDeref(file, asm.RegA, asm.GetReg(asm.RegD, types.Ptr_Size), types.Ptr_Size)
        asm.MovRegDeref(file, asm.RegB, asm.GetOffsetedReg(asm.RegD, types.Ptr_Size, 8), types.I32_Size)
    } else {
        asm.MovRegDeref(file, asm.RegA, asm.GetReg(asm.RegD, types.Ptr_Size), arrType.Ptr.BaseType.Size())
    }
}

func (e *Field) Compile(file *os.File) {
    t := e.Obj.GetType()

    if sType,ok := t.(types.StructType); ok {
        obj := identObj.Get(sType.Name)
        if s,ok := obj.(*structDec.Struct); ok {
            i := s.GetFieldNum(e.FieldName.Str)
            s := s.GetTypes()[i].Size()
            asm.MovRegDeref(file, asm.RegA, e.Obj.Addr(i), s)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is not a struct but a %v\n", e.Obj.GetName(), t)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }
}

func (e *ArrayLit) Compile(file *os.File) {}

func (e *Ident) Compile(file *os.File) {
    if c,ok := e.Obj.(*consts.Const); ok {
        l := Lit{ Val: c.GetVal(), Type: c.GetType() }
        l.Compile(file)
        return
    }

    if v,ok := e.Obj.(vars.Var); ok {
        asm.MovRegDeref(file, asm.RegA, v.Addr(0), v.GetType().Size())
        return
    }

    if _,ok := e.Obj.(*fn.Func); ok {
        fmt.Fprintf(os.Stderr, "[ERROR] TODO: expr.go compile Ident for functions\n")
        os.Exit(1)
        return
    }

    fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", e.Name)
    fmt.Fprintln(os.Stderr, "\t" + e.Pos.At())
    os.Exit(1)
}
func (e *Paren) Compile(file *os.File) { e.Expr.Compile(file) }
func (e *Unary) Compile(file *os.File) {
    e.typeCheck()

    // compile time evaluation
    if c := e.ConstEval(); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, e.Operand.GetType().Size(), c.Str)
        return
    }

    e.Operand.Compile(file)

    switch e.Operator.Type {
    case token.Minus:
        size := e.Operand.GetType().Size()
        asm.Neg(file, asm.GetReg(asm.RegA, size), size)

    case token.Mul:
        if _,ok := e.Operand.(*Ident); !ok {
            if _,ok := e.Operand.(*Paren); !ok {
                fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
                fmt.Fprintln(os.Stderr, "\t" + e.Operator.At())
                os.Exit(1)
            }
        }

        asm.DerefRax(file, e.GetType().Size())
    }
}
func (e *Binary) Compile(file *os.File) {
    e.typeCheck()

    size := e.OperandL.GetType().Size()
    if sizeR := e.OperandR.GetType().Size(); sizeR > size {
        size = sizeR
    }

    // compile time evaluation (constEval whole expr)
    if c := e.ConstEval(); c.Type != token.Unknown {
        if c.Type == token.Boolean {
            if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
        }

        asm.MovRegVal(file, asm.RegA, size, c.Str)
        return
    }


    // +,-,*,/, <,<=,>,>=,==,!=
    if e.Operator.Type != token.And && e.Operator.Type != token.Or {
        // compile time evaluation (constEval only left expr)
        if c := e.OperandL.ConstEval(); c.Type != token.Unknown {
            if c.Type == token.Boolean {
                if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
            }

            asm.MovRegVal(file, asm.RegA, size, c.Str)
        } else {
            e.OperandL.Compile(file)

            // compile time evaluation (constEval only right expr)
            if c := e.OperandR.ConstEval(); c.Type != token.Unknown {
                if c.Type == token.Boolean {
                    if c.Str == "true" { c.Str = "1" } else { c.Str = "0" }
                }

                asm.BinaryOp(file, e.Operator.Type, c.Str, size)
                return
            }
        }

        if ident,ok := e.OperandR.(*Ident); ok {
            if v,ok := ident.Obj.(vars.Var); ok {
                if t := v.GetType(); t.Size() < size {
                    asm.MovRegDeref(file, asm.RegC, v.Addr(0), t.Size())
                    asm.BinaryOpReg(file, e.Operator.Type, asm.RegC, size)
                } else {
                    asm.BinaryOp(file, e.Operator.Type, fmt.Sprintf("%s [%s]", asm.GetWord(t.Size()), v.Addr(0)), size)
                }
            }
        } else {
            asm.Push(file, asm.RegA)

            e.OperandR.Compile(file)
            asm.MovRegReg(file, asm.RegB, asm.RegA, size)

            asm.Pop(file, asm.RegA)
            asm.BinaryOpReg(file, e.Operator.Type, asm.RegB, size)
        }

    // &&, ||
    } else {
        // compile time evaluation
        if c := e.OperandL.ConstEval(); c.Type != token.Unknown {
            if e.Operator.Type == token.And && c.Str == "false" {
                asm.MovRegVal(file, asm.RegA, size, "0")
                return
            }
            if e.Operator.Type == token.Or && c.Str == "true" {
                asm.MovRegVal(file, asm.RegA, size, "1")
                return
            }

            e.OperandR.Compile(file)
        } else {
            e.OperandL.Compile(file)

            count := cond.LogicalOp(file, e.Operator)
            e.OperandR.Compile(file)
            cond.LogicalOpEnd(file, count)
        }
    }
}

func (e *FnCall) Compile(file *os.File) {
    e.typeCheck()

    regIdx := 0
    for _, val := range e.Values {
        // compile time evaluation:
        if v := val.ConstEval(); v.Type != token.Unknown {
            fn.PassVal(file, regIdx, v, val.GetType())

        } else if ident,ok := val.(*Ident); ok {
            fn.PassVar(file, regIdx, ident.Obj.(vars.Var))

        } else {
            val.Compile(file)
            fn.PassReg(file, regIdx, val.GetType())
        }

        if val.GetType().GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    e.Ident.Obj.(*fn.Func).Call(file)
}

func (e *XCase) Compile(file *os.File, switchCount uint) {
    cond.CaseStart(file)

    if e.Cond == nil {
        cond.CaseBody(file)
        e.Expr.Compile(file)
        return
    }

    // compile time evaluation
    if val := e.Cond.ConstEval(); val.Type != token.Unknown {
        if val.Str == "true" {
            cond.CaseBody(file)
            e.Expr.Compile(file)
            cond.CaseBodyEnd(file, switchCount)
        }

        return
    }

    if i,ok := e.Cond.(*Ident); ok {
        cond.CaseVar(file, i.Obj.Addr(0))
    } else {
        e.Cond.Compile(file)
        cond.CaseExpr(file)
    }

    cond.CaseBody(file)
    e.Expr.Compile(file)
    cond.CaseBodyEnd(file, switchCount)
}

func (e *XSwitch) Compile(file *os.File) {
    e.typeCheck()

    // compile time evaluation
    if c := e.ConstEval(); c.Type != token.Unknown {
        asm.MovRegVal(file, asm.RegA, types.TypeOfVal(c.Str).Size(), c.Str)
        return
    }

    count := cond.StartSwitch()

    for i := 0; i < len(e.Cases)-1; i++ {
        e.Cases[i].Compile(file, count)
    }
    cond.InLastCase()
    e.Cases[len(e.Cases)-1].Compile(file, count)

    cond.EndSwitch(file)
}

func (e *BadExpr) Compile(file *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}


func (e *BadExpr)   At() string { return "" }
func (e *FnCall)    At() string { return e.Ident.At() }
func (e *Lit)       At() string { return e.Val.At() }
func (e *FieldLit)  At() string { return e.Name.At() }
func (e *StructLit) At() string { return e.Pos.At() }
func (e *ArrayLit)  At() string { return e.Pos.At() }
func (e *Indexed)   At() string { return e.ArrExpr.At() }
func (e *Field)     At() string { return e.Pos.At() }
func (e *Ident)     At() string { return e.Pos.At() }
func (e *Unary)     At() string { return e.Operator.At() }
func (e *Binary)    At() string { return e.OperandL.At() }    // TODO: At() of Operand with higher precedence
func (e *Paren)     At() string { return e.ParenLPos.At() }
func (e *XSwitch)   At() string { return e.Pos.At() }
func (e *XCase)     At() string { return e.ColonPos.At() }

func (e *BadExpr)   End() string { return "" }
func (e *FnCall)    End() string { return e.ParenRPos.At() }
func (e *Lit)       End() string { return e.Val.At() }
func (e *FieldLit)  End() string { return e.Value.End() }
func (e *StructLit) End() string { return e.BraceRPos.At() }
func (e *ArrayLit)  End() string { return e.BraceRPos.At() }
func (e *Indexed)   End() string { return e.BrackRPos.At() }
func (e *Field)     End() string { return e.FieldName.At() }
func (e *Ident)     End() string { return e.Pos.At() }
func (e *Unary)     End() string { return e.Operand.At() }
func (e *Binary)    End() string { return e.OperandR.At() }
func (e *Paren)     End() string { return e.ParenRPos.At() }
func (e *XSwitch)   End() string { return e.BraceRPos.At() }
func (e *XCase)     End() string { return e.Expr.At() }
