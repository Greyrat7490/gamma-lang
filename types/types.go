package types

import "strconv"

// TODO: correct sizes for i32 and bool (not just 64bit)

type Type interface {
    Size() int
    String() string
}

type I32Type struct {}
type BoolType struct {}

type PtrType struct {
    BaseType Type
}
type StrType struct {
    ptr PtrType
    size I32Type
}


func (t I32Type)  Size() int { return 8 }
func (t BoolType) Size() int { return 8 }
func (t PtrType)  Size() int { return 8 }
func (t StrType)  Size() int { return t.ptr.Size() + t.size.Size() }

func (t I32Type)  String() string { return "i32" }
func (t BoolType) String() string { return "bool" }
func (t PtrType)  String() string { return "*" + t.BaseType.String() }
func (t StrType)  String() string { return "str" }


func ToType(s string) Type {
    isPtr := false
    if s[0] == '*' {
        s = s[1:]
        isPtr = true
    }
    
    var base Type
    switch s {
    case "str":
        base = StrType{} // TODO: set ptr to *char
    case "i32":
        base = I32Type{}
    case "bool":
        base = BoolType{}
    default:
        base = nil
    }

    if isPtr {
        return PtrType{ BaseType: base }
    }

    return base
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
