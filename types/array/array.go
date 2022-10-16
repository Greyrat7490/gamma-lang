package array

import (
    "os"
    "fmt"
    "gamma/types"
    "gamma/types/str"
    "gamma/cmpTime/constVal"
    "gamma/gen/asm/x86_64"
    "gamma/gen/asm/x86_64/nasm"
)

var arrLits []arrLit

type arrLit struct {
    baseType types.Type
    values []constVal.ConstVal
}

func GetValues(arrLitIdx int) []constVal.ConstVal {
    return arrLits[arrLitIdx].values
}

func Add(typ types.ArrType, values []constVal.ConstVal) (i int) {
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
                addStr(*v.(*constVal.StrConst))
            }

        case types.StructType:
            for _, v := range values {
                addStruct(t, *v.(*constVal.StructConst))
            }

        case types.IntType:
            for _, v := range values {
                addInt(t.Size(), *v.(*constVal.IntConst))
            }
        case types.UintType:
            for _, v := range values {
                addUint(t.Size(), *v.(*constVal.UintConst))
            }
        case types.BoolType:
            for _, v := range values {
                addBool(*v.(*constVal.BoolConst))
            }
        case types.CharType:
            for _, v := range values {
                addChar(*v.(*constVal.CharConst))
            }

        case types.PtrType:
            for _, v := range values {
                addPtr(*v.(*constVal.PtrConst))
            }

        case types.ArrType:
            for _, v := range values {
                addArr(t, *v.(*constVal.ArrConst))
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", typ.Ptr.BaseType)
            os.Exit(1)
        }
    }

    return
}

func addInt(size uint, val constVal.IntConst) {
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(size), val))
}

func addUint(size uint, val constVal.UintConst) {
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(size), val))
}

func addBool(val constVal.BoolConst) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Bool_Size), val.GetVal()))
}

func addChar(val constVal.CharConst) {
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(types.Char_Size), val))
}

func addStr(val constVal.StrConst) {
    nasm.AddData(fmt.Sprintf("  %s _str%d", asm.GetDataSize(types.Ptr_Size), val))
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(types.I32_Size), str.GetSize(int(val))))
}

func addArr(baseType types.Type, val constVal.ArrConst) {
    values := GetValues(int(val))

    switch t := baseType.(type) {
    case types.StrType:
        for _, v := range values {
            addStr(*v.(*constVal.StrConst))
        }

    case types.StructType:
        for _, v := range values {
            addStruct(t, *v.(*constVal.StructConst))
        }

    case types.IntType:
        for _, v := range values {
            addInt(t.Size(), *v.(*constVal.IntConst))
        }
    case types.UintType:
        for _, v := range values {
            addUint(t.Size(), *v.(*constVal.UintConst))
        }
    case types.BoolType:
        for _, v := range values {
            addBool(*v.(*constVal.BoolConst))
        }
    case types.CharType:
        for _, v := range values {
            addChar(*v.(*constVal.CharConst))
        }

    case types.PtrType:
        for _, v := range values {
            addPtr(*v.(*constVal.PtrConst))
        }

    case types.ArrType:
        for _, v := range values {
            addArr(t, *v.(*constVal.ArrConst))
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", baseType)
        os.Exit(1)
    }
}

func addPtr(val constVal.PtrConst) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(types.Ptr_Size), val.Addr))
}

func addStruct(t types.StructType, val constVal.StructConst) {
    for i,v := range val.Fields {
        switch v := v.(type) {
        case *constVal.StrConst:
            addStr(*v)

        case *constVal.StructConst:
            addStruct(t.Types[i].(types.StructType), *v)

        case *constVal.IntConst:
            addInt(t.Types[i].Size(), *v)

        case *constVal.UintConst:
            addUint(t.Types[i].Size(), *v)

        case *constVal.BoolConst:
            addBool(*v)

        case *constVal.CharConst:
            addChar(*v)

        case *constVal.PtrConst:
            addPtr(*v)

        case *constVal.ArrConst:
            addArr(t.Types[i], *v)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", t)
            os.Exit(1)
        }
    }
}

