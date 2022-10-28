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
            nasm.AddBss(fmt.Sprintf("_arr%d: %s %d", i, asm.GetBssSize(arr.typ.Ptr.BaseType.Size()), len(arr.elems)))
        } else {
            nasm.AddData(fmt.Sprintf("_arr%d:", i))
            switch t := arr.typ.Ptr.BaseType.(type) {
            case types.StrType:
                for _, v := range arr.elems {
                    addStr(*v.(*constVal.StrConst))
                }

            case types.StructType:
                for _, v := range arr.elems {
                    addStruct(t, *v.(*constVal.StructConst))
                }

            case types.IntType:
                for _, v := range arr.elems {
                    addInt(t.Size(), *v.(*constVal.IntConst))
                }
            case types.UintType:
                for _, v := range arr.elems {
                    addUint(t.Size(), *v.(*constVal.UintConst))
                }
            case types.BoolType:
                for _, v := range arr.elems {
                    addBool(*v.(*constVal.BoolConst))
                }
            case types.CharType:
                for _, v := range arr.elems {
                    addChar(*v.(*constVal.CharConst))
                }

            case types.PtrType:
                for _, v := range arr.elems {
                    addPtr(*v.(*constVal.PtrConst))
                }

            case types.ArrType:
                for _, v := range arr.elems {
                    addArr(t, GetValues(uint64(*v.(*constVal.ArrConst))))
                }

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", arr.typ.Ptr.BaseType)
                os.Exit(1)
            }
        }
    }
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

    case types.IntType:
        for _, v := range elems {
            addInt(t.Size(), *v.(*constVal.IntConst))
        }
    case types.UintType:
        for _, v := range elems {
            addUint(t.Size(), *v.(*constVal.UintConst))
        }
    case types.BoolType:
        for _, v := range elems {
            addBool(*v.(*constVal.BoolConst))
        }
    case types.CharType:
        for _, v := range elems {
            addChar(*v.(*constVal.CharConst))
        }

    case types.PtrType:
        for _, v := range elems {
            addPtr(*v.(*constVal.PtrConst))
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
            addArr(t, GetValues(uint64(*v)))

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", t)
            os.Exit(1)
        }
    }
}

