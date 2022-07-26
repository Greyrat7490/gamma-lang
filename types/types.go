package types

import (
	"fmt"
	"strconv"
)

type TypeKind int
const (
    I32  TypeKind = iota
    Bool TypeKind = iota
    Ptr  TypeKind = iota
    Arr  TypeKind = iota
    Str  TypeKind = iota
)

const (
    I32_Size  int = 4
    Bool_Size int = 1
    Ptr_Size  int = 8
    Arr_Size  int = 8
    Str_Size  int = Ptr_Size + I32_Size
)

type Type interface {
    Size()    int
    String()  string
    GetKind() TypeKind
}

type I32Type  struct {}
type BoolType struct {}
type PtrType  struct {
    BaseType Type
}
type ArrType  struct {
    Ptr PtrType
    Len uint64
}
type StrType  struct {
    ptr  PtrType
    size I32Type
}

func (t I32Type)  GetKind() TypeKind { return I32  }
func (t BoolType) GetKind() TypeKind { return Bool }
func (t StrType)  GetKind() TypeKind { return Str  }
func (t PtrType)  GetKind() TypeKind { return Ptr  }
func (t ArrType)  GetKind() TypeKind { return Arr  }

func (t I32Type)  Size() int { return I32_Size }
func (t BoolType) Size() int { return Bool_Size }
func (t StrType)  Size() int { return t.ptr.Size() + t.size.Size() }
func (t PtrType)  Size() int { return Ptr_Size }
func (t ArrType)  Size() int { return Arr_Size }

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
    return fmt.Sprintf("[%d]%v", t.Len, t.Ptr.BaseType)
}


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

func AreCompatible(destType Type, srcType Type) bool {
    if destType == srcType {
        return true
    }

    // allow generic ptr with any other pointer
    if destType.GetKind() == Ptr && srcType.GetKind() == Ptr {
        if p, ok := destType.(PtrType); ok {
            if p.BaseType == nil { return true }
        }
    }

    return false
}
