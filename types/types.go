package types

import (
    "os"
    "fmt"
    "strconv"
)

type TypeKind int
const (
    Int     TypeKind = iota
    Uint    TypeKind = iota
    Char    TypeKind = iota
    Bool    TypeKind = iota
    Ptr     TypeKind = iota
    Arr     TypeKind = iota
    Vec     TypeKind = iota
    Str     TypeKind = iota
    Struct  TypeKind = iota
    Generic TypeKind = iota
)

const (
    I8_Size   uint = 1
    I16_Size  uint = 2
    I32_Size  uint = 4
    I64_Size  uint = 8

    U8_Size   uint = 1
    U16_Size  uint = 2
    U32_Size  uint = 4
    U64_Size  uint = 8

    Char_Size uint = 1
    Bool_Size uint = 1
    Ptr_Size  uint = 8
    Arr_Size  uint = 8
    Str_Size  uint = Ptr_Size + U32_Size
    Vec_Size  uint = Ptr_Size + U64_Size + U64_Size
)

type Type interface {
    Size()        uint
    String()      string
    GetKind()     TypeKind
}

type CharType struct {}
type BoolType struct {}
type UintType struct {
    size uint
}
type IntType struct {
    size uint
}
type PtrType struct {
    BaseType Type
}
type ArrType struct {
    BaseType Type
    Lens []uint64
}
type VecType struct {
    BaseType Type
}
type StrType struct {}
type StructType struct {
    Name string
    Types []Type
    names map[string]int
    isBigStruct bool
    isAligned bool
    size uint
}
type GenericType struct {
    Name string
    CurUsedType Type
    UsedTypes []Type
}

func isAligned(types []Type, size uint) (aligned bool, rest uint)  {
    for _,t := range types {
        switch t := t.(type) {
        case StructType:
            a,r := isAligned(t.Types, size)
            if !a {
                return false, 0
            }
            size += r

        case VecType:
            return false, 0

        case StrType:
            a,r := isAligned([]Type{ PtrType{}, UintType{ size: U32_Size } }, size)
            if !a {
                return false, 0
            }
            size += r

        default:
            size += t.Size()
        }

        if size > 8 {
            return false, 0
        } else if size == 8 {
            size = 0
        }
    }

    return true, size
}

func CreateInt(intSize uint) IntType {
    return IntType{ size: intSize }
}

func CreateUint(uintSize uint) UintType {
    return UintType{ size: uintSize }
}

func CreateStructType(name string, types []Type, names []string) StructType {
    ns := map[string]int{}
    for i, n := range names {
        ns[n] = i
    }

    size := uint(0)
    for _,t := range types {
        size += t.Size()
    }

    isBigStruct := false
    if size > 16 {
        isBigStruct = true
    }

    aligned,_ := isAligned(types, 0)
    if !aligned {
        isBigStruct = true
    }

    return StructType{ Name: name, Types: types, isBigStruct: isBigStruct, isAligned: aligned, size: size, names: ns }
}

func (t *StructType) GetOffset(field string) (offset int64) {
    if fieldNum, ok := t.names[field]; ok {
        for i := 0; i < fieldNum; i++ {
            offset += int64(t.Types[i].Size())
        }
        return
    }

    return -1
}

func (t *StructType) GetType(field string) Type {
    if i, ok := t.names[field]; ok {
        return t.Types[i]
    }

    return nil
}

func (t *StructType) GetFieldNum(field string) int {
    if i, ok := t.names[field]; ok {
        return i
    }

    return -1
}

func (t *StructType) GetFields() []string {
    fields := make([]string, len(t.Types))

    for name, idx := range t.names {
        fields[idx] = name
    }

    return fields
}

func IsBigStruct(t Type) bool {
    if _,ok := t.(VecType); ok {
        return true
    }

    if t,ok := t.(StructType); ok {
        return t.isBigStruct
    }

    return false
}

func RegCount(t Type) uint {
    switch t.GetKind() {
    case Str:
        return 2

    case Vec:
        return 3

    case Struct:
        if IsBigStruct(t) {
            return 0
        }

        if t.Size() > 8 {
            return 2
        } else {
            return 1
        }

    default:
        return 1
    }
}

func (t IntType)     GetKind() TypeKind { return Int }
func (t UintType)    GetKind() TypeKind { return Uint }
func (t CharType)    GetKind() TypeKind { return Char }
func (t BoolType)    GetKind() TypeKind { return Bool }
func (t StrType)     GetKind() TypeKind { return Str  }
func (t PtrType)     GetKind() TypeKind { return Ptr  }
func (t ArrType)     GetKind() TypeKind { return Arr  }
func (t VecType)     GetKind() TypeKind { return Vec  }
func (t StructType)  GetKind() TypeKind { return Struct }
func (t GenericType) GetKind() TypeKind { 
    if t.CurUsedType != nil {
        return t.CurUsedType.GetKind()
    }

    return Generic
}

func (t IntType)     Size() uint { return t.size }
func (t UintType)    Size() uint { return t.size }
func (t CharType)    Size() uint { return Char_Size }
func (t BoolType)    Size() uint { return Bool_Size }
func (t StrType)     Size() uint { return Str_Size }
func (t PtrType)     Size() uint { return Ptr_Size }
func (t ArrType)     Size() uint { return Arr_Size }
func (t VecType)     Size() uint { return Vec_Size }
func (t StructType)  Size() uint { return t.size }
func (t GenericType) Size() uint { 
    if t.CurUsedType != nil {
        return t.CurUsedType.Size()
    }

    return 0
}

func (t IntType)  String() string {
    switch t.size {
    case I8_Size:
        return "i8"
    case I16_Size:
        return "i16"
    case I32_Size:
        return "i32"
    case I64_Size:
        return "i64"
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected int size %d", t.size)
        os.Exit(1)
        return ""
    }
}
func (t UintType)  String() string {
    switch t.size {
    case U8_Size:
        return "u8"
    case U16_Size:
        return "u16"
    case U32_Size:
        return "u32"
    case U64_Size:
        return "u64"
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected uint size %d", t.size)
        os.Exit(1)
        return ""
    }
}
func (t CharType) String() string { return "char" }
func (t BoolType) String() string { return "bool" }
func (t StrType)  String() string { return "str"  }
func (t PtrType)  String() string {
    return "*" + t.BaseType.String()
}
func (t ArrType)  String() string {
    res := ""
    for _,l := range t.Lens {
        res += fmt.Sprintf("[%d]", l)
    }

    return res + t.BaseType.String()
}
func (t VecType) String() string {
    return "[$]" + t.BaseType.String()
}
func (t StructType) String() string { return t.Name }
func (t GenericType) String() string {
    if t.CurUsedType != nil {
        return t.CurUsedType.String()
    }
    return t.Name
}

func ToBaseType(s string) Type {
    switch s {
    case "i8":
        return IntType{ size: I8_Size }
    case "i16":
        return IntType{ size: I16_Size }
    case "i32":
        return IntType{ size: I32_Size }
    case "i64":
        return IntType{ size: I64_Size }
    case "u8":
        return UintType{ size: U8_Size }
    case "u16":
        return UintType{ size: U16_Size }
    case "u32":
        return UintType{ size: U32_Size }
    case "u64":
        return UintType{ size: U64_Size }
    case "char":
        return CharType{}
    case "bool":
        return BoolType{}
    case "str":
        return StrType{}
    default:
        return nil
    }
}

func TypeOfVal(val string) Type {
    switch {
    case val[0] == '"' && val[len(val) - 1] == '"':
        return StrType{}
    case val[0] == '\'' && val[len(val) - 1] == '\'':
        return CharType{}
    case val == "true" || val == "false":
        return BoolType{}
    case len(val) > 2 && val[0:2] == "0x":
        if _, err := strconv.ParseUint(val, 0, 64); err == nil {
            return UintType{ size: U64_Size }
        }
    default:
        if _, err := strconv.ParseInt(val, 10, 32); err == nil {
            return IntType{ size: I32_Size }
        }
        if _, err := strconv.ParseInt(val, 10, 64); err == nil {
            return IntType{ size: I64_Size }
        }
        if _, err := strconv.ParseUint(val, 0, 64); err == nil {
            return UintType{ size: U64_Size }
        }
    }

    return nil
}


func MinSizeInt(val int64) uint {
    if val < 0 {
        val = -val - 1
    }

    switch {
    case val < 0x80:       // i8
        return 1
    case val < 0x8000:     // i16
        return 2
    case val < 0x80000000: // i32
        return 4
    default:             // i64
        if (val >> 63) == 0 { // not u64
            return 8
        }

        return 9
    }
}

func MinSizeUint(val uint64) uint {
    switch {
    case val <= 0xff:       // 8bit
        return 1
    case val <= 0xffff:     // 16bit
        return 2
    case val <= 0xffffffff: // 32bit
        return 4
    default:                // 64bit
        return 8
    }
}

func Equal(destType Type, srcType Type) bool {
    if srcType == nil {
        return false
    }

    switch t := destType.(type) {
    case VecType:
        if t2,ok := srcType.(VecType); ok {
            return Equal(t.BaseType, t2.BaseType)
        }

    case ArrType:
        if t2,ok := srcType.(ArrType); ok {
            if Equal(t.BaseType, t2.BaseType) {
                if len(t.Lens) == len(t2.Lens) {
                    for i,l := range t.Lens {
                        if l != t2.Lens[i] {
                            return false
                        }
                    }

                    return true
                }
            }
        }

    case PtrType:
        if t2,ok := srcType.(PtrType); ok {
            return Equal(t.BaseType, t2.BaseType)
        }

    case StructType:
        if t2,ok := srcType.(StructType); ok {
            for i,t := range t.Types {
                if !Equal(t, t2.Types[i]) {
                    return false
                }
            }

            return true
        }

    case GenericType:
        if t2,ok := srcType.(GenericType); ok {
            return t.Name == t2.Name
        }

    case IntType:
        if t2,ok := srcType.(IntType); ok {
            return t2.Size() <= destType.Size()
        }

    case StrType:
        return destType.GetKind() == srcType.GetKind()

    default:
        return destType == srcType
    }

    return false
}
