package array

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/struct"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/nasm"
)

var arrLits []arrLit

type arrLit struct {
    baseType types.Type
    values []token.Token
}

func GetValues(arrLitIdx int) []token.Token {
    return arrLits[arrLitIdx].values
}

func Add(typ types.ArrType, values []token.Token) (i int) {
    i = len(arrLits)
    arr := arrLit{ baseType: typ.Ptr.BaseType, values: values }
    arrLits = append(arrLits, arr)

    if len(arr.values) == 0 {
        var total uint64 = 1
        for _,l := range typ.Lens {
            total *= l
        }
        nasm.AddBss(fmt.Sprintf("_arr%d: %s %d", i, asm.GetBssSize(arr.baseType.Size()), total))
    } else {
        nasm.AddData(fmt.Sprintf("_arr%d:", i))

        switch t := typ.Ptr.BaseType.(type) {
        case types.StrType:
            for _, v := range values {
                addStr(v)
            }

        case types.StructType:
            for _, v := range values {
                addStruct(t, v)
            }

        case types.BoolType:
            for _, v := range values {
                addBool(v)
            }

        case types.CharType:
            for _, v := range values {
                addChar(v)
            }

        case types.IntType, types.PtrType, types.ArrType:
            for _, v := range values {
                addBasic(t.Size(), v)
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", typ)
            os.Exit(1)
        }
    }

    return
}

func addBasic(size uint, val token.Token) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(size), val.Str))
}

func addBool(val token.Token) {
    if val.Str == "true" {
        nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Bool_Size), "1"))
    } else {
        nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Bool_Size), "0"))
    }
}

func addChar(val token.Token) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Bool_Size), fmt.Sprint(int(val.Str[1]))))
}

func addStr(val token.Token) {
    strIdx := str.Add(val)

    nasm.AddData(fmt.Sprintf("  %s _str%d", asm.GetDataSize(types.Ptr_Size), strIdx))
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(types.I32_Size), str.GetSize(strIdx)))
}

func addStruct(t types.StructType, val token.Token) {
    if val.Type != token.Number {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Number but got %v\n", val)
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    idx,_ := strconv.ParseUint(val.Str, 10, 64)
    values := structLit.GetValues(idx)

    for i, v := range values {
        switch t := t.Types[i].(type) {
        case types.StrType:
            addStr(v)

        case types.StructType:
            addStruct(t, v)

        case types.BoolType:
            addBool(v)

        case types.CharType:
            addChar(v)

        case types.IntType, types.PtrType, types.ArrType:
            addBasic(t.Size(), v)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", t)
            os.Exit(1)
        }
    }
}

