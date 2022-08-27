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
  * [x] more on stack (right to left)

  * [x] struct:         use int/float fields
  * [x] big struct:     on stack (right to left)
    * [x] bigger than 16Byte or unaligned fields (more than 2 regs needed)

  * [x] return value:
    * [x] int: rax, rdx
    * [ ] float: xmm0, xmm1
    * [x] big struct: stack (addr in rdi)
  * [x] caller cleans stack
  * [x] callee reserves space

  * [x] stack always 16bit aligned
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

func (f *Func) Addr(field uint) string {
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

func DefArg(file *os.File, regIdx uint, v vars.Var) {
    switch t := v.GetType().(type) {
    case types.StrType:
        asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, regs[regIdx])
        asm.MovDerefReg(file, v.Addr(1), types.I32_Size, regs[regIdx+1])

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovDerefReg(file, v.Addr(0), types.Ptr_Size, regs[regIdx])
            asm.MovDerefReg(file, v.OffsetedAddr(int(types.Ptr_Size)), t.Types[len(t.Types)-1].Size(), regs[regIdx+1])
        } else {
            asm.MovDerefReg(file, v.Addr(0), t.Size(), regs[regIdx])
        }

    default:
        asm.MovDerefReg(file, v.Addr(0), t.Size(), regs[regIdx])
    }
}

func packValues(file *os.File, valtypes []types.Type, values []token.Token) int {
    offset := uint(0)
    for i,t := range valtypes {
        asm.MovDerefVal(file, fmt.Sprintf("rsp+%d", offset), t.Size(), values[i].Str)
        offset += t.Size()
        if offset == types.Ptr_Size {
            return i+1
        }
    }

    return len(valtypes)
}

func PassVal(file *os.File, regIdx uint, value token.Token, valtype types.Type) {
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

            if len(t.Types) == 1 {
                asm.MovRegVal(file, regs[regIdx], t.Size(), fields[0].Str)
            } else {
                // TODO only if necessary
                file.WriteString(fmt.Sprintf("sub rsp, %d\n", types.Ptr_Size))

                if t.Size() > uint(8) {
                    i := packValues(file, t.Types, fields)
                    asm.MovRegDeref(file, regs[regIdx], "rsp", types.Ptr_Size)
                    // TODO pack values here too if necessary
                    asm.MovRegVal(file, regs[regIdx+1], t.Types[i].Size(), fields[i].Str)
                } else {
                    packValues(file, t.Types, fields)
                    asm.MovRegDeref(file, regs[regIdx], "rsp", t.Size())
                }

                file.WriteString(fmt.Sprintf("add rsp, %d\n", types.Ptr_Size))
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

func PassVar(file *os.File, regIdx uint, otherVar vars.Var) {
    switch t := otherVar.GetType().(type) {
    case types.StrType:
        asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), types.Ptr_Size)
        asm.MovRegDeref(file, regs[regIdx+1], otherVar.Addr(1), types.I32_Size)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), types.Ptr_Size)
            asm.MovRegDeref(file, regs[regIdx+1], otherVar.OffsetedAddr(int(types.Ptr_Size)), t.Types[1].Size())
        } else {
            asm.MovRegDeref(file, regs[regIdx],   otherVar.Addr(0), t.Size())
        }

    case types.BoolType, types.I32Type, types.PtrType, types.ArrType:
        asm.MovRegDeref(file, regs[regIdx], otherVar.Addr(0), t.Size())

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot pass var %s of type %v yet\n", otherVar.GetName(), t)
        os.Exit(1)
    }
}

func PassReg(file *os.File, regIdx uint, argType types.Type) {
    switch t := argType.(type) {
    case types.StrType:
        asm.MovRegReg(file, regs[regIdx],   asm.RegGroup(0), types.Ptr_Size)
        asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), types.I32_Size)

    case types.StructType:
        if t.Size() > uint(8) {
            asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), types.Ptr_Size)
            asm.MovRegReg(file, regs[regIdx+1], asm.RegGroup(1), t.Types[1].Size())
        } else {
            asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), t.Size())
        }

    default:
        asm.MovRegReg(file, regs[regIdx], asm.RegGroup(0), argType.Size())
    }
}

func PassValStack(file *os.File, value token.Token, valtype types.Type) {
    switch t := valtype.(type) {
    case types.StrType:
        strIdx := str.Add(value)

        asm.PushVal(file, fmt.Sprint(str.GetSize(strIdx)))
        asm.PushVal(file, fmt.Sprintf("_str%d", strIdx))

    case types.StructType:
        if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
            fields := structLit.GetValues(idx)
            if len(t.Types) == 1 {
                asm.PushVal(file, fields[0].Str)
            } else {
                if t.Size() > uint(8) {
                    asm.PushVal(file, fields[0].Str)
                    asm.MovDerefVal(file, fmt.Sprintf("rsp+%d", t.Types[0].Size()), t.Types[1].Size(), fields[1].Str)
                } else {
                    asm.PushVal(file, fields[0].Str)
                    asm.MovDerefVal(file, fmt.Sprintf("rsp+%d", t.Types[0].Size()), t.Types[1].Size(), fields[1].Str)
                }
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected struct literal converted to a Number but got %v\n", value)
            fmt.Fprintln(os.Stderr, "\t" + value.At())
            os.Exit(1)
        }
    default:
        asm.PushVal(file, value.Str)
    }
}

func PassVarStack(file *os.File, otherVar vars.Var) {
    switch t := otherVar.GetType().(type) {
    case types.StrType:
        asm.PushDeref(file, otherVar.Addr(1))
        asm.PushDeref(file, otherVar.Addr(0))

    case types.StructType:
        if t.Size() > uint(8) {
            asm.PushDeref(file, otherVar.Addr(1))
        }
        asm.PushDeref(file, otherVar.Addr(0))

    case types.BoolType, types.I32Type, types.PtrType, types.ArrType:
        asm.PushDeref(file, otherVar.Addr(0))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] cannot pass var %s of type %v yet\n", otherVar.GetName(), t)
        os.Exit(1)
    }
}

func PassRegStack(file *os.File, argType types.Type) {
    switch t := argType.(type) {
    case types.StrType:
        asm.PushReg(file, asm.RegGroup(1))
        asm.PushReg(file, asm.RegGroup(0))

    case types.StructType:
        if t.Size() > uint(8) {
            asm.PushReg(file, asm.RegGroup(1))
        }
        asm.PushReg(file, asm.RegGroup(0))

    default:
        asm.PushReg(file, asm.RegGroup(0))
    }
}

func PassBigStructLit(file *os.File, t types.StructType, value token.Token, offset int) {
    if idx,err := strconv.ParseUint(value.Str, 10, 64); err == nil {
        fields := structLit.GetValues(idx)

        for i,t := range t.Types {
            switch t := t.(type) {
            case types.StrType:
                strIdx := str.Add(fields[i])

                asm.MovDerefVal(file,
                    asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
                    types.Ptr_Size,
                    fmt.Sprintf("_str%d", strIdx))
                asm.MovDerefVal(file,
                    asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset + int(types.Ptr_Size)),
                    types.I32_Size,
                    fmt.Sprint(str.GetSize(strIdx)))

            case types.StructType:
                PassBigStructLit(file, t, fields[i], offset)

            case types.BoolType:
                if fields[i].Str == "true" {
                    asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), types.Bool_Size, "1")
                } else {
                    asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), types.Bool_Size, "0")
                }

            default:
                asm.MovDerefVal(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), t.Size(), fields[i].Str)
            }
            offset += int(t.Size())
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected struct literal converted to a Number but got %v\n", value)
        fmt.Fprintln(os.Stderr, "\t" + value.At())
        os.Exit(1)
    }
}

func PassBigStructVar(file *os.File, t types.StructType, v vars.Var, offset int) {
    for i := 0; i < int(t.Size()/types.Ptr_Size); i++ {
        asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), v.OffsetedAddr(offset), types.Ptr_Size, asm.RegA)
        offset += int(types.Ptr_Size)
    }

    if size := t.Size() % types.Ptr_Size; size != 0 {
        asm.MovDerefDeref(file, asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset), v.OffsetedAddr(offset), size, asm.RegA)
    }
}

func PassBigStructReg(file *os.File, t types.StructType) {
    offset := 0
    for i := 0; i < int(t.Size()/types.Ptr_Size); i++ {
        asm.MovDerefDeref(
            file,
            asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
            asm.GetOffsetedReg(asm.RegA, types.Ptr_Size, offset),
            types.Ptr_Size,
            asm.RegB,
        )

        offset += int(types.Ptr_Size)
    }

    if size := t.Size() % types.Ptr_Size; size != 0 {
        asm.MovDerefDeref(
            file,
            asm.GetOffsetedReg(asm.RegC, types.Ptr_Size, offset),
            asm.GetOffsetedReg(asm.RegA, types.Ptr_Size, offset),
            size,
            asm.RegB,
        )
    }
}

func RetBigStructLit(file *os.File, t types.StructType, val token.Token) {
    PassBigStructLit(file, t, val, 0)
}

func RetBigStructVar(file *os.File, t types.StructType, v vars.Var) {
    PassBigStructVar(file, t, v, 0)
}

func RetBigStructExpr(file *os.File, t types.StructType) {
    PassBigStructReg(file, t)
}
