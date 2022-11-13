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

var arrayData []constVal.ArrConst

func SetElem(arrIdx uint64, idx uint64, val constVal.ConstVal) {
    if len(arrayData[arrIdx].Elems) == 0 {
        sum := uint64(0)
        for _,l := range arrayData[arrIdx].Type.Lens {
            sum += l
        }
        arrayData[arrIdx].Elems = make([]constVal.ConstVal, sum)
    }

    arrayData[arrIdx].Elems[idx] = val
}

func GetValues(arrIdx uint64) []constVal.ConstVal {
    return arrayData[arrIdx].Elems
}

func Add(t types.ArrType, elems []constVal.ConstVal) uint64 {
    arr := constVal.ArrConst{ Idx: uint64(len(arrayData)), Type: t, Elems: elems }
    arrayData = append(arrayData, arr)
    return arr.Idx
}

func Gen() {
    for i,arr := range arrayData {
        if len(arr.Elems) == 0 {
            sum := uint64(0)
            for _,l := range arr.Type.Lens {
                sum += l
            }

            nasm.AddBss(fmt.Sprintf("_arr%d: %s %d", i, asm.GetBssSize(arr.Type.BaseType.Size()), sum))
        } else {
            nasm.AddData(fmt.Sprintf("_arr%d:", i))
            addArr(&arr)
        }
    }
}

func addBasic(size uint, val constVal.ConstVal) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(size), val.GetVal()))
}

func addStr(val *constVal.StrConst) {
    nasm.AddData(fmt.Sprintf("  %s _str%d", asm.GetDataSize(types.Ptr_Size), uint64(*val)))
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(types.I32_Size), str.GetSize(uint64(*val))))
}

func addStrPtr(val *constVal.StrConst) {
    nasm.AddData(fmt.Sprintf("  %s _str%d", asm.GetDataSize(types.Ptr_Size), uint64(*val)))
}

func addDefault(t types.Type) {
    switch t := t.(type) {
    case types.StrType:
        nasm.AddData(fmt.Sprintf("  %s 0", asm.GetDataSize(types.Ptr_Size)))
        nasm.AddData(fmt.Sprintf("  %s 0", asm.GetDataSize(types.U32_Size)))
    case types.StructType:
        for _,t := range t.Types {
            addDefault(t)
        }
    default:
        nasm.AddData(fmt.Sprintf("  %s 0", asm.GetDataSize(t.Size())))
    }
}

func addArr(arr *constVal.ArrConst) {
    switch t := arr.Type.BaseType.(type) {
    case types.StrType:
        for _, v := range arr.Elems {
            if v == nil {
                addDefault(t)
            } else {
                addStr(v.(*constVal.StrConst))
            }
        }

    case types.StructType:
        for _, v := range arr.Elems {
            if v == nil {
                addDefault(t)
            } else {
                addStruct(t, v.(*constVal.StructConst))
            }
        }

    case types.ArrType:
        for _, v := range arr.Elems {
            if v == nil {
                addDefault(t)
            } else {
                addArr(v.(*constVal.ArrConst))
            }
        }

    case types.IntType, types.UintType, types.BoolType, types.CharType, types.PtrType:
        for _, v := range arr.Elems {
            if v == nil {
                addDefault(t)
            } else {
                addBasic(t.Size(), v)
            }
        }


    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", arr.Type.BaseType)
        os.Exit(1)
    }
}

func addStruct(t types.StructType, val *constVal.StructConst) {
    for i,v := range val.Fields {
        switch v := v.(type) {
        case *constVal.StrConst:
            if t.Types[i].GetKind() == types.Str {
                addStr(v)
            } else {        // *char cast
                addStrPtr(v)
            }

        case *constVal.StructConst:
            addStruct(t.Types[i].(types.StructType), v)

        case *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst, *constVal.PtrConst:
            addBasic(t.Types[i].Size(), v)

        case *constVal.ArrConst:
            addArr(v)

        case nil:
            addDefault(t)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", t)
            os.Exit(1)
        }
    }
}

