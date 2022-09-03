package types

import (
    "fmt"
    "os"
    "strconv"
)

type TypeKind int
const (
    Int    TypeKind = iota
    Char   TypeKind = iota
    Bool   TypeKind = iota
    Ptr    TypeKind = iota
    Arr    TypeKind = iota
    Str    TypeKind = iota
    Struct TypeKind = iota
)

const (
    I8_Size   uint = 1
    I16_Size  uint = 2
    I32_Size  uint = 4
    I64_Size  uint = 8
    Char_Size uint = 1
    Bool_Size uint = 1
    Ptr_Size  uint = 8
    Arr_Size  uint = 8
    Str_Size  uint = Ptr_Size + I32_Size
)

type Type interface {
    Size()        uint
    String()      string
    GetKind()     TypeKind
}

type CharType struct {}
type BoolType struct {}
type IntType struct {
    size uint
}
type PtrType struct {
    BaseType Type
}
type ArrType  struct {
    Ptr PtrType
    Lens []uint64
}
type StrType  struct {
    ptr  PtrType
    size IntType
}
type StructType struct {
    Name string
    Types []Type
    isBigStruct bool
    isAligned bool
    size uint
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

        case StrType:
            a,r := isAligned([]Type{ PtrType{}, IntType{ size: I32_Size } }, size)
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

func CreateStructType(name string, types []Type) StructType {
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

    return StructType{ Name: name, Types: types, isBigStruct: isBigStruct, isAligned: aligned, size: size }
}

func (t StructType) GetOffset(fieldNum uint) (offset int) {
    for i := uint(0); i < fieldNum; i++ {
        offset += int(t.Types[i].Size())
    }

    return
}

func IsBigStruct(t Type) bool {
    if t,ok := t.(StructType); ok {
        return t.isBigStruct
    }

    return false
}

func RegCount(t Type) uint {
    switch t.GetKind() {
    case Str:
        return 2

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

func (t IntType)    GetKind() TypeKind { return Int }
func (t CharType)   GetKind() TypeKind { return Char }
func (t BoolType)   GetKind() TypeKind { return Bool }
func (t StrType)    GetKind() TypeKind { return Str  }
func (t PtrType)    GetKind() TypeKind { return Ptr  }
func (t ArrType)    GetKind() TypeKind { return Arr  }
func (t StructType) GetKind() TypeKind { return Struct }

func (t IntType)    Size() uint { return t.size }
func (t CharType)   Size() uint { return Char_Size }
func (t BoolType)   Size() uint { return Bool_Size }
func (t StrType)    Size() uint { return t.ptr.Size() + t.size.Size() }
func (t PtrType)    Size() uint { return Ptr_Size }
func (t ArrType)    Size() uint { return Arr_Size }
func (t StructType) Size() uint { return t.size }

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
func (t CharType) String() string { return "char" }
func (t BoolType) String() string { return "bool" }
func (t StrType)  String() string { return "str"  }
func (t PtrType)  String() string {
    if t.BaseType == nil {
        return "ptr(generic)"
    }
    return "*" + t.BaseType.String()
}
func (t ArrType)  String() string {
    res := ""
    for _,l := range t.Lens {
        res += fmt.Sprintf("[%d]", l)
    }

    return res + t.Ptr.BaseType.String()
}
func (t StructType) String() string { return t.Name }

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
    case "char":
        return CharType{}
    case "bool":
        return BoolType{}
    case "str":
        return StrType{ ptr: PtrType{ BaseType: CharType{} } }
    default:
        return nil
    }
}

func TypeOfVal(val string) Type {
    if val[0] == '"' && val[len(val) - 1] == '"' {
        return StrType{}
    } else if val[0] == '\'' && val[len(val) - 1] == '\'' {
        return CharType{}
    } else if _, err := strconv.Atoi(val); err == nil {
        return IntType{ size: I32_Size }
    } else if val == "true" || val == "false" {
        return BoolType{}
    } else {
        return nil
    }
}
