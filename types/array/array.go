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

var arrayData []arrData

type arrData struct {
    typ types.ArrType
    elems []constVal.ConstVal
}

func SetElem(arrIdx uint64, idx uint64, val constVal.ConstVal) {
    if len(arrayData[arrIdx].elems) == 0 {
        sum := uint64(0)
        for _,l := range arrayData[arrIdx].typ.Lens {
            sum += l
        }
        arrayData[arrIdx].elems = make([]constVal.ConstVal, sum)
    }

    arrayData[arrIdx].elems[idx] = val
}

func GetValues(arrIdx uint64) []constVal.ConstVal {
    return arrayData[arrIdx].elems
}

func Add(elems []constVal.ConstVal, typ types.ArrType) uint64 {
    i := uint64(len(arrayData))
    arrayData = append(arrayData, arrData{ typ: typ, elems: elems })
    return i
}

func Gen() {
    for i,arr := range arrayData {
        if len(arr.elems) == 0 {
            sum := uint64(0)
            for _,l := range arr.typ.Lens {
                sum += l
            }

            nasm.AddBss(fmt.Sprintf("_arr%d: %s %d", i, asm.GetBssSize(arr.typ.BaseType.Size()), sum))
        } else {
            nasm.AddData(fmt.Sprintf("_arr%d:", i))
            addArr(arr.typ.BaseType, arr.elems)
        }
    }
}

func addBasic(size uint, val constVal.ConstVal) {
    nasm.AddData(fmt.Sprintf("  %s %s", asm.GetDataSize(size), val.GetVal()))
}

func addStr(val constVal.StrConst) {
    nasm.AddData(fmt.Sprintf("  %s _str%d", asm.GetDataSize(types.Ptr_Size), uint64(val)))
    nasm.AddData(fmt.Sprintf("  %s %d", asm.GetDataSize(types.I32_Size), str.GetSize(uint64(val))))
}

func addStrPtr(val constVal.StrConst) {
    nasm.AddData(fmt.Sprintf("  %s _str%d", asm.GetDataSize(types.Ptr_Size), uint64(val)))
}

func addArr(baseType types.Type, elems []constVal.ConstVal) {
    switch t := baseType.(type) {
    case types.StrType:
        for _, v := range elems {
            addStr(*v.(*constVal.StrConst))
        }

    case types.StructType:
        for _, v := range elems {
            addStruct(t, *v.(*constVal.StructConst))
        }

    case types.IntType, types.UintType, types.BoolType, types.CharType, types.PtrType:
        for _, v := range elems {
            addBasic(t.Size(), v)
        }

    case types.ArrType:
        for _, v := range elems {
            addArr(t, GetValues(uint64(*v.(*constVal.ArrConst))))
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", baseType)
        os.Exit(1)
    }
}

func addStruct(t types.StructType, val constVal.StructConst) {
    for i,v := range val.Fields {
        switch v := v.(type) {
        case *constVal.StrConst:
            if t.Types[i].GetKind() == types.Str {
                addStr(*v)
            } else {        // *char cast
                addStrPtr(*v)
            }

        case *constVal.StructConst:
            addStruct(t.Types[i].(types.StructType), *v)

        case *constVal.IntConst, *constVal.UintConst, *constVal.BoolConst, *constVal.CharConst, *constVal.PtrConst:
            addBasic(t.Types[i].Size(), v)

        case *constVal.ArrConst:
            addArr(t, GetValues(uint64(*v)))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", t)
            os.Exit(1)
        }
    }
}

