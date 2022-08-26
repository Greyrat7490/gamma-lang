package vars

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/struct"
    "gamma/asm/x86_64"
    "gamma/asm/x86_64/nasm"
)

type GlobalVar struct {
    decPos token.Pos
    name string
    vartype types.Type
}

func CreateGlobalVar(name token.Token, t types.Type) GlobalVar {
    return GlobalVar{ name: name.Str, decPos: name.Pos, vartype: t }
}

func (v *GlobalVar) SetType(t types.Type) {
    if v.vartype != nil {
        fmt.Println("[ERROR] setting the type of a var again is not allowed")
        os.Exit(1)
    }

    v.vartype = t
}

func (v *GlobalVar) String() string {
    return fmt.Sprintf("{%s %s}", v.name, v.vartype)
}

func (v *GlobalVar) GetName() string {
    return v.name
}

func (v *GlobalVar) GetPos () token.Pos {
    return v.decPos
}

func (v *GlobalVar) GetType() types.Type {
    return v.vartype
}

func (v *GlobalVar) OffsetedAddr(offset int) string {
    if offset > 0 {
        return fmt.Sprintf("%s+%d", v.name, offset)
    } else if offset < 0 {
        return fmt.Sprintf("%s%d", v.name, offset)
    } else {
        return v.name
    }
}

func (v *GlobalVar) Addr(fieldNum int) string {
    switch t := v.vartype.(type) {
    case types.StrType:
        if fieldNum == 1 {
            return fmt.Sprintf("%s+%d", v.name, types.Ptr_Size)
        }

    case types.StructType:
        if fieldNum != 0 {
            return fmt.Sprintf("%s+%d", v.name, t.GetOffset(uint(fieldNum)))
        }
    }

    return v.name
}


func (v *GlobalVar) DefVal(file *os.File, val token.Token) {
    nasm.AddData(fmt.Sprintf("%s:\n", v.name))

    switch t := v.vartype.(type) {
    case types.StrType:
        defStr(val)
    case types.ArrType:
        defArr(val.Str)
    case types.StructType:
        defStruct(t, val)
    case types.BoolType:
        defBool(val.Str)
    case types.I32Type:
        defInt(val.Str)
    case types.PtrType:
        defPtr(val.Str)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] define global var of typ %v is not supported yet\n", v.vartype)
        fmt.Fprintln(os.Stderr, "\t" + v.decPos.At())
        os.Exit(1)
    }
}


func defStruct(t types.StructType, val token.Token) {
    if val.Type != token.Number {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    idx,_ := strconv.ParseUint(val.Str, 10, 64)

    for i,v := range structLit.GetValues(idx) {
        switch t := t.Types[i].(type) {
        case types.StrType:
            defStr(v)
        case types.I32Type:
            defInt(v.Str)
        case types.BoolType:
            defBool(v.Str)
        case types.PtrType:
            defPtr(v.Str)
        case types.ArrType:
            defArr(v.Str)
        case types.StructType:
            defStruct(t, v)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet\n", t)
            os.Exit(1)
        }
    }
}

func defInt(val string) {
    nasm.AddData(fmt.Sprintf("  %s %s\n", asm.GetDataSize(types.I32_Size), val))
}

func defPtr(val string) {
    nasm.AddData(fmt.Sprintf("  %s %s\n", asm.GetDataSize(types.Ptr_Size), val))
}

func defBool(val string) {
    if val == "true" {
        val = "1"
    } else {
        val = "0"
    }

    nasm.AddData(fmt.Sprintf("  %s %s\n", asm.GetDataSize(types.Bool_Size), val))
}

func defStr(val token.Token) {
    strIdx := str.Add(val)
    nasm.AddData(fmt.Sprintf("  %s _str%d\n  %s %d", asm.GetDataSize(types.Ptr_Size), strIdx, asm.GetDataSize(types.I32_Size), str.GetSize(strIdx)))
}

func defArr(val string) {
    nasm.AddData(fmt.Sprintf("  %s _arr%s\n", asm.GetDataSize(types.Ptr_Size), val))
}
