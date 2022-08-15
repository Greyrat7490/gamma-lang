package fn

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/struct"
    "gamma/asm/x86_64"
    "gamma/ast/identObj/vars"
)

/*
calling convention
  * int   args1-6: rdi, rsi, rdx, rcx, r8, r9
  * float args1-8: xmm0 - xmm7 (TODO)
  * return: rax, rbx, rcx, rdx, rdi, rsi, r8, r9

  * rest: from right to left on stack (TODO)

  * caller cleans stack (TODO)
  * callee reserves space (multiple of 16)
*/

var regs []asm.RegGroup = []asm.RegGroup{ asm.RegDi, asm.RegSi, asm.RegD, asm.RegC, asm.RegR8, asm.RegR9 }

type Func struct {
    decPos token.Pos
    name string
    args []types.Type
    retType types.Type
    frameSize uint
}

func CreateFunc(name token.Token, args []types.Type, retType types.Type) Func {
    // frameSize = 1 -> invalid value
    return Func{ name: name.Str, decPos: name.Pos, args: args, retType: retType, frameSize: 1 }
}

func (f *Func) GetArgs() []types.Type {
    return f.args
}

func (f *Func) GetName() string {
    return f.name
}

func (f *Func) GetType() types.Type {
    // TODO
    return nil
}

func (f *Func) GetRetType() types.Type {
    return f.retType
}

func (f *Func) GetPos() token.Pos {
    return f.decPos
}

func (f *Func) SetFrameSize(size uint) {
    if f.frameSize != 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] setting the frameSize of a function again is not allowed")
        os.Exit(1)
    }

    // size has to be the multiple of 16byte
    f.frameSize = (size + 15) & ^uint(15)
}

func (f *Func) Addr(fieldNum int) string {
    fmt.Fprintln(os.Stderr, "[ERROR] TODO: func.go Addr()")
    os.Exit(1)
    return ""
}

func (f *Func) Define(file *os.File) {
    file.WriteString(f.name + ":\n")
    file.WriteString("push rbp\nmov rbp, rsp\n")
    if f.frameSize > 0 {
        file.WriteString(fmt.Sprintf("sub rsp, %d\n", f.frameSize))
    }
}

func End(file *os.File) {
    file.WriteString("leave\n")
    file.WriteString("ret\n\n")
}

func (f *Func) Call(file *os.File) {
    file.WriteString("call " + f.name + "\n")
}

func DefArg(file *os.File, regIdx int, v vars.Var) {
    t := v.GetType()

    switch t.GetKind() {
    case types.Str:
        setArg(file, v.Addr(0), regIdx, types.Ptr_Size)
        setArg(file, v.Addr(1), regIdx+1, types.I32_Size)

    case types.Struct:
        t := t.(types.StructType)
        for i,fieldType := range t.Types {
            setArg(file, v.Addr(i), regIdx+i, fieldType.Size())
        }

    default:
        setArg(file, v.Addr(0), regIdx, t.Size())
    }
}

func setArg(file *os.File, addr string, regIdx int, size uint) {
    if regIdx >= len(regs) {
        fmt.Fprintf(os.Stderr, "[ERROR] not enough regs left to set args (max 6) %d more needed\n", regIdx - len(regs) + 1)
        os.Exit(1)
    }

    asm.MovDerefReg(file, addr, size, regs[regIdx])
}

func PassVal(file *os.File, regIdx int, value token.Token, valtype types.Type) {
    switch valtype.GetKind() {
    case types.Str:
        strIdx := str.Add(value)

        asm.MovRegVal(file, regs[regIdx],   types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovRegVal(file, regs[regIdx+1], types.I32_Size, fmt.Sprint(str.GetSize(strIdx)))

    case types.Arr:
        if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
            asm.MovRegVal(file, regs[regIdx], types.Ptr_Size, fmt.Sprintf("_arr%d", idx))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected array literal converted to a Number but got %v\n", value)
            fmt.Fprintln(os.Stderr, "\t" + value.At())
            os.Exit(1)
        }

    case types.Struct:
        if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
            t := valtype.(types.StructType)

            fields := structLit.GetValues(idx)
            for i,fieldType := range t.Types {
                asm.MovRegVal(file, regs[regIdx+i], fieldType.Size(), fields[i].Str)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected struct literal converted to a Number but got %v\n", value)
            fmt.Fprintln(os.Stderr, "\t" + value.At())
            os.Exit(1)
        }

    case types.Bool:
        if value.Str == "true" { value.Str = "1" } else { value.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        asm.MovRegVal(file, regs[regIdx], valtype.Size(), value.Str)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot pass value of type %v yet\n", valtype)
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }
}

func PassVar(file *os.File, regIdx int, otherVar vars.Var) {
    switch t := otherVar.GetType().(type) {
    case types.StrType:
        asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), types.Ptr_Size)
        asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr(1), types.I32_Size)

    case types.StructType:
        for i,fieldType := range t.Types {
            asm.MovRegDeref(file, regs[regIdx+i], otherVar.Addr(i), fieldType.Size())
        }

    case types.BoolType, types.I32Type, types.PtrType, types.ArrType:
        asm.MovRegDeref(file, regs[regIdx], otherVar.Addr(0), t.Size())

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot pass var %s of type %v yet\n", otherVar.GetName(), t)
        os.Exit(1)
    }
}

func PassReg(file *os.File, regIdx int, argType types.Type) {
    switch t := argType.(type) {
    case types.StrType:
        asm.MovRegReg(file, regs[regIdx],   asm.RegA, types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegB, types.I32_Size)

    case types.StructType:
        for i,t := range t.Types {
            asm.MovRegReg(file, regs[regIdx+i], asm.RegGroup(i), t.Size())
        }

    default:
        asm.MovRegReg(file, regs[regIdx], asm.RegA, argType.Size())
    }
}
