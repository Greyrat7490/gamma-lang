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
System V AMD64 ABI calling convention
  * [x] int:    rdi, rsi, rdx, rcx, r8, r9
  * [ ] float:  xmm0 - xmm7
  * [ ] more on stack (right to left)

  * [x] struct:         use int/float fields
  * [x] big struct:     on stack (right to left)
    * for now always if more than 2 fields
    * [ ] bigger than 16Byte or unaligned fields (more than 2 regs needed)

  * [ ] return value:
    * [x] int: rax, rdx
    * [ ] float: xmm0, xmm1
    * [ ] stack (addr in rdi) if more/bigger (see big struct)
  * [x] caller cleans stack
  * [x] callee reserves space (multiple of 16)
  * [ ] stack always 16bit aligned
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

func DefArg(file *os.File, regIdx int, v vars.Var) bool {
    t := v.GetType()

    switch t.GetKind() {
    case types.Str:
        setArg(file, v.Addr(0), regIdx, types.Ptr_Size)
        setArg(file, v.Addr(1), regIdx+1, types.I32_Size)

    case types.Struct:
        t := t.(types.StructType)

        if len(t.Types) > 2 { // skip defining big structs
            return false
        }

        for i,fieldType := range t.Types {
            setArg(file, v.Addr(i), regIdx+i, fieldType.Size())
        }

    default:
        setArg(file, v.Addr(0), regIdx, t.Size())
    }
    return true
}

func setArg(file *os.File, addr string, regIdx int, size uint) {
    if regIdx >= len(regs) {
        fmt.Fprintf(os.Stderr, "[ERROR] not enough regs left to set args (max 6) %d more needed\n", regIdx - len(regs) + 1)
        os.Exit(1)
    }

    asm.MovDerefReg(file, addr, size, regs[regIdx])
}

func PassVal(file *os.File, regIdx int, value token.Token, valtype types.Type) {
    switch t := valtype.(type) {
    case types.StrType:
        strIdx := str.Add(value)

        asm.MovRegVal(file, regs[regIdx],   types.Ptr_Size, fmt.Sprintf("_str%d", strIdx))
        asm.MovRegVal(file, regs[regIdx+1], types.I32_Size, fmt.Sprint(str.GetSize(strIdx)))

    case types.ArrType:
        if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
            asm.MovRegVal(file, regs[regIdx], types.Ptr_Size, fmt.Sprintf("_arr%d", idx))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected array literal converted to a Number but got %v\n", value)
            fmt.Fprintln(os.Stderr, "\t" + value.At())
            os.Exit(1)
        }

    case types.StructType:
        if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
            fields := structLit.GetValues(idx)
            for i,fieldType := range t.Types {
                asm.MovRegVal(file, regs[regIdx+i], fieldType.Size(), fields[i].Str)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected struct literal converted to a Number but got %v\n", value)
            fmt.Fprintln(os.Stderr, "\t" + value.At())
            os.Exit(1)
        }

    case types.BoolType:
        if value.Str == "true" {
            asm.MovRegVal(file, regs[regIdx], valtype.Size(), "1")
        } else {
            asm.MovRegVal(file, regs[regIdx], valtype.Size(), "0")
        }

    case types.I32Type, types.PtrType:
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
        asm.MovRegReg(file, regs[regIdx],   asm.RegGroup(0), types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), types.I32_Size)

    case types.StructType:
        for i,t := range t.Types {
            asm.MovRegReg(file, regs[regIdx+i], asm.RegGroup(i), t.Size())
        }

    default:
        asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), argType.Size())
    }
}

func PassBigStructLit(file *os.File, t types.StructType, value token.Token) {
    if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
        fields := structLit.GetValues(idx)

        for i := range t.Types {
            asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, int(types.Ptr_Size)*i), types.Ptr_Size, fields[i].Str)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected struct literal converted to a Number but got %v\n", value)
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }

}

func PassBigStructVar(file *os.File, t types.StructType, v vars.Var) {
    for i := range t.Types {
        asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, int(types.Ptr_Size)*i), v.Addr(i), types.Ptr_Size, asm.RegA)
    }
}

func PassBigStructReg(file *os.File, t types.StructType) {
    for i := range t.Types {
        asm.MovDerefReg(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, int(types.Ptr_Size)*i), types.Ptr_Size, asm.RegGroup(i))
    }
}
