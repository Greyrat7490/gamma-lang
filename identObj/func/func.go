package fn

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/asm/x86_64"
    "gorec/identObj/vars"
)

/*
System V AMD64 ABI calling convention
  * int   args1-6: rdi, rsi, rdx, rcx, r8, r9
  * float args1-8: xmm0 - xmm7

  * push args from right to left for more args
  * return value in eax/rax
  * caller cleans stack
  * callee reserves space (multiple of 16)
*/

var regs []asm.RegGroup = []asm.RegGroup{ asm.RegDi, asm.RegSi, asm.RegD, asm.RegC, asm.RegR8, asm.RegR9 }

type Func struct {
    name token.Token
    args []types.Type
    frameSize int
}

func CreateFunc(name token.Token) Func {
    return Func{ name: name, args: nil, frameSize: -1 }
}
func CreateFuncWithArgs(name token.Token, args []types.Type) Func {
    return Func{ name: name, args: args, frameSize: -1 }
}
func (f *Func) SetArgs(args []types.Type) {
    if f.args != nil {
        fmt.Println("[ERROR] setting the arguments of a function again is not allowed")
        os.Exit(1)
    }

    f.args = args
}
func (f *Func) GetArgs() []types.Type {
    return f.args
}

func (f *Func) GetName() token.Token {
    return f.name
}

func (f *Func) At() string {
    return f.name.At()
}

func (f *Func) SetFrameSize(frameSize int) {
    if f.frameSize != -1 {
        fmt.Println("[ERROR] setting the frameSize of a function again is not allowed")
        os.Exit(1)
    }

    f.frameSize = frameSize
}

func (f *Func) Define(file *os.File) {
    file.WriteString(f.name.Str + ":\n")
    file.WriteString("push rbp\nmov rbp, rsp\n")
    reserveSpace(file, f.frameSize)
}

func reserveSpace(file *os.File, size int) {
    if size > 0 {
        // size has to be the multiple of 16byte
        size = (size + 15) & ^15
        file.WriteString(fmt.Sprintf("sub rsp, %d\n", size))
    }
}

func End(file *os.File) {
    file.WriteString("leave\n")
    file.WriteString("ret\n\n")
}

func (f *Func) Call(file *os.File) {
    file.WriteString("call " + f.name.Str + "\n")
}

func DefArg(file *os.File, regIdx int, v vars.Var) {
    t := v.GetType()

    if t.GetKind() == types.Str {
        setArg(file, v.Addr(0), regIdx, types.Ptr_Size)
        setArg(file, v.Addr(1), regIdx+1, types.I32_Size)
    } else {
        setArg(file, v.Addr(0), regIdx, t.Size())
    }
}

func setArg(file *os.File, addr string, regIdx int, size int) {
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

    case types.Bool:
        if value.Str == "true" { value.Str = "1" } else { value.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        asm.MovRegVal(file, regs[regIdx], valtype.Size(), value.Str)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) could not get type of value \"%s\"\n", value.Str)
        os.Exit(1)
    }
}

func PassVar(file *os.File, regIdx int, otherVar vars.Var) {
    t := otherVar.GetType()

    switch t.GetKind() {
    case types.Str:
        asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), types.Ptr_Size)
        asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr(1), types.I32_Size)

    case types.Bool, types.I32, types.Ptr:
        asm.MovRegDeref(file, regs[regIdx], otherVar.Addr(0), t.Size())

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) type of var \"%s\" is not correct\n", otherVar.GetName())
        os.Exit(1)
    }
}

func PassReg(file *os.File, regIdx int, argType types.Type) {
    if argType.GetKind() == types.Str {
        asm.MovRegReg(file, regs[regIdx],   asm.RegA, types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegB, types.Ptr_Size)
    } else {
        asm.MovRegReg(file, regs[regIdx], asm.RegA, argType.Size())
    }
}
