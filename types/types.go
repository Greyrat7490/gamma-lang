package types

import (
    "fmt"
    "strconv"
)

type TypeKind int
const (
    I32    TypeKind = iota
    Bool   TypeKind = iota
    Ptr    TypeKind = iota
    Arr    TypeKind = iota
    Str    TypeKind = iota
    Struct TypeKind = iota
)

const (
    I32_Size  uint = 4
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

type I32Type  struct {}
type BoolType struct {}
type PtrType  struct {
    BaseType Type
}
type ArrType  struct {
    Ptr PtrType
    Lens []uint64
}
type StrType  struct {
    ptr  PtrType
    size I32Type
}
type StructType struct {
    Name string
    Types []Type
    isBigStruct bool
    isAligned bool
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
            a,r := isAligned([]Type{ PtrType{}, I32Type{} }, size)
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

    return StructType{ Name: name, Types: types, isBigStruct: isBigStruct, isAligned: aligned }
}

func IsBigStruct(t Type) bool {
    if t,ok := t.(StructType); ok {
        return t.isBigStruct
    }

    return false
}

func (t I32Type)    GetKind() TypeKind { return I32  }
func (t BoolType)   GetKind() TypeKind { return Bool }
func (t StrType)    GetKind() TypeKind { return Str  }
func (t PtrType)    GetKind() TypeKind { return Ptr  }
func (t ArrType)    GetKind() TypeKind { return Arr  }
func (t StructType) GetKind() TypeKind { return Struct }

func (t I32Type)    Size() uint { return I32_Size }
func (t BoolType)   Size() uint { return Bool_Size }
func (t StrType)    Size() uint { return t.ptr.Size() + t.size.Size() }
func (t PtrType)    Size() uint { return Ptr_Size }
func (t ArrType)    Size() uint { return Arr_Size }
func (t StructType) Size() uint {
    var res uint = 0
    for _,t := range t.Types {
        res += t.Size()
    }
    return res
}

func (t I32Type)  String() string { return "i32"  }
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
    case "str":
        return StrType{} // TODO: set ptr to *char
    case "i32":
        return I32Type{}
    case "bool":
        return BoolType{}
    default:
        return nil
    }
}

func TypeOfVal(val string) Type {
    if val[0] == '"' && val[len(val) - 1] == '"' {
        return StrType{}
    } else if _, err := strconv.Atoi(val); err == nil {
        return I32Type{}
    } else if val == "true" || val == "false" {
        return BoolType{}
    } else {
        return nil
    }
}

func Check(destType Type, srcType Type) bool {
    switch t := destType.(type) {
    case ArrType:
        if t2,ok := srcType.(ArrType); ok {
            if t.Ptr.BaseType == t2.Ptr.BaseType {
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
        // allow generic ptr with any other pointer
        if t.BaseType == nil && srcType.GetKind() == Ptr {
            return true
        }
        return destType == srcType

    case StructType:
        if t2,ok := srcType.(StructType); ok {
            for i,t := range t.Types {
                if !Check(t, t2.Types[i]) {
                    return false
                }
            }

            return true
        }

    default:
        return destType == srcType
    }

    return false
}
