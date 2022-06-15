package ast

import (
    "os"
    "fmt"
    "strings"
    "gorec/arithmetic"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

type OpExpr interface {
    Op
    Compile(asm *os.File)
    GetType() types.Type
}

type BadExpr struct{}

type OpFnCall struct {
    FnName token.Token
    Values []OpExpr
}

type LitExpr struct {
    Val token.Token
    Type types.Type
}

type IdentExpr struct {
    Ident token.Token
}

type UnaryExpr struct {
    Operator token.Token
    Operand OpExpr
}

type BinaryExpr struct {
    OperandL OpExpr
    Operator token.Token
    OperandR OpExpr
}

type ParenExpr struct {
    ParenLPos token.Pos
    Expr OpExpr
    ParenRPos token.Pos
}

func (o *BadExpr)   GetType() types.Type { return nil }
func (o *OpFnCall)  GetType() types.Type { return nil }
func (o *LitExpr)   GetType() types.Type { return o.Type }
func (o *ParenExpr) GetType() types.Type { return o.Expr.GetType() }
func (o *IdentExpr) GetType() types.Type {
    v := vars.GetVar(o.Ident.Str)
    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared)\n", o.Ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + o.Ident.At())
        os.Exit(1)
    }

    return v.GetType()
}

func (o *UnaryExpr) GetType() types.Type {
    if o.Operator.Type == token.Amp {
        return types.PtrType{ BaseType: o.Operand.GetType() }
    }

    if o.Operator.Type == token.Mul {
        if ptr, ok := o.Operand.GetType().(types.PtrType); ok {
            return ptr.BaseType
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] you cannot deref this expre (expected a pointer/address)")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }
    }

    return o.Operand.GetType()
}

func (o *BinaryExpr) GetType() types.Type {
    if  o.Operator.Type == token.Eql || o.Operator.Type == token.Neq ||
        o.Operator.Type == token.Grt || o.Operator.Type == token.Lss ||
        o.Operator.Type == token.Geq || o.Operator.Type == token.Leq {
        return types.BoolType{}
    }

    t := o.OperandL.GetType()
    if t == nil {
        return o.OperandR.GetType()
    }

    // check for cases like 420 + &v1
    if t.GetKind() == types.I32 {
        if other := o.OperandR.GetType(); other.GetKind() != types.I32 {
            return other
        }
    }

    // check for cases like ptr1 - ptr2
    if t.GetKind() == types.Ptr {
        if other := o.OperandR.GetType(); other.GetKind() == types.Ptr {
            return types.I32Type{}
        }
    }

    return t
}


func (o *LitExpr)   Compile(asm *os.File) {}
func (o *IdentExpr) Compile(asm *os.File) {}
func (o *ParenExpr) Compile(asm *os.File) { o.Expr.Compile(asm) }
func (o *UnaryExpr) Compile(asm *os.File) {
    switch o.Operator.Type {
    case token.Mul:
        switch e := o.Operand.(type) {
        case *IdentExpr:
            vars.ValToRegA(asm, e.Ident)

        case *ParenExpr:
            o.Operand.Compile(asm)

        default:
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable or parentheses after \"*\"")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }

    case token.Amp:
        if e,ok := o.Operand.(*IdentExpr); ok {
            vars.AddrToRax(asm, e.Ident)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] expected a variable after \"&\"")
            fmt.Fprintln(os.Stderr, "\t" + o.Operator.At())
            os.Exit(1)
        }

    default:
        size := o.Operand.GetType().Size()

        switch e := o.Operand.(type) {
        case *IdentExpr:
            vars.ValToRegA(asm, e.Ident)

        case *LitExpr:
            vars.Write(asm, fmt.Sprintf("mov %s, %s\n", vars.GetReg(vars.RegA, size), e.Val.Str))

        default:
            o.Operand.Compile(asm)
        }

        if o.Operator.Type == token.Minus {
            vars.Write(asm, fmt.Sprintf("neg %s\n", vars.GetReg(vars.RegA, size)))
        }
    }
}
func (o *BinaryExpr) Compile(asm *os.File) {
    size := o.OperandL.GetType().Size()
    if sizeR := o.OperandR.GetType().Size(); sizeR > size {
        size = sizeR
    }

    switch e := o.OperandL.(type) {
    case *LitExpr:
        vars.Write(asm, fmt.Sprintf("mov %s, %s\n", vars.GetReg(vars.RegA, size), e.Val.Str))
    case *IdentExpr:
        vars.ValToRegA(asm, e.Ident)
    case *UnaryExpr:
        o.OperandL.Compile(asm)
        if e.Operator.Type == token.Mul {
            vars.Write(asm, fmt.Sprintf("mov %s, %s [rax]\n", vars.GetReg(vars.RegA, size), vars.GetWord(size)))
        }
    default:
        o.OperandL.Compile(asm)
    }


    switch e := o.OperandR.(type) {
    case *LitExpr:
        arith.BinaryOp(asm, o.Operator.Type, e.Val.Str, size)
    case *IdentExpr:
        v := vars.GetVar(e.Ident.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] variable %s is not declared\n", e.Ident.Str)
            fmt.Fprintln(os.Stderr, "\t" + e.Ident.At())
            os.Exit(1)
        }

        arith.BinaryOp(asm, o.Operator.Type, v.Get(), size)

    default:
        vars.Write(asm, fmt.Sprintf("push %s\n", vars.GetReg(vars.RegA, 8)))

        o.OperandR.Compile(asm)
        if u,ok := e.(*UnaryExpr); ok && u.Operator.Type == token.Mul {
            vars.Write(asm, fmt.Sprintf("mov %s, %s [rax]\n", vars.GetReg(vars.RegB, size), vars.GetWord(size)))
        } else {
            vars.Write(asm, fmt.Sprintf("mov %s, %s\n", vars.GetReg(vars.RegB, size), vars.GetReg(vars.RegA, size)))
        }

        vars.Write(asm, fmt.Sprintf("pop %s\n", vars.GetReg(vars.RegA, 8)))
        arith.BinaryOpReg(asm, o.Operator.Type, vars.RegB, size)
    }
}

func (o *OpFnCall) Compile(asm *os.File) {
    regIdx := 0
    for _, val := range o.Values {
        switch e := val.(type) {
        case *LitExpr:
            fn.PassVal(asm, o.FnName, regIdx, e.Val)

        case *IdentExpr:
            fn.PassVar(asm, regIdx, e.Ident)

        case *UnaryExpr:
            size := val.GetType().Size()

            val.Compile(asm)
            if e.Operator.Type == token.Mul {
                vars.Write(asm, fmt.Sprintf("mov %s, %s [rax]\n", vars.GetReg(vars.RegA, size), vars.GetWord(size)))
            }
            fn.PassReg(asm, regIdx, size)

        default:
            val.Compile(asm)
            fn.PassReg(asm, regIdx, val.GetType().Size())
        }

        if val.GetType().GetKind() == types.Str {
            regIdx += 2
        } else {
            regIdx++
        }
    }

    fn.CallFunc(asm, o.FnName)
}

func (o *BadExpr) Compile(asm *os.File) {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
}


func (o *LitExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("%s(%v)\n", o.Val.Str, o.Type)
}

func (o *IdentExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + o.Ident.Str + "(Name)\n"
}

func (o *OpFnCall) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    res := fmt.Sprintf("%sOP_CALL_FN:\n%s%s\n", s, s2, o.FnName.Str)
    for _, e := range o.Values {
        res += e.Readable(indent+1)
    }

    return res
}

func (o *UnaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return fmt.Sprintf("%sOP_UNARY:\n%s%s(%s)\n", s, s2, o.Operator.Str, o.Operator.Type.Readable()) +
        o.Operand.Readable(indent+1)
}

func (o *BinaryExpr) Readable(indent int) string {
    s := strings.Repeat("   ", indent)
    s2 := s + "   "

    return s + "OP_BINARY:\n" +
        o.OperandL.Readable(indent+1) +
        s2 + fmt.Sprintf("%s(%s)\n", o.Operator.Str, o.Operator.Type.Readable()) +
        o.OperandR.Readable(indent+1)
}

func (o *ParenExpr) Readable(indent int) string {
    return strings.Repeat("   ", indent) + "PAREN:\n" + o.Expr.Readable(indent+1)
}

func (o *BadExpr) Readable(indent int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] bad expression")
    os.Exit(1)
    return ""
}
