package types

import (
    "os"
    "fmt"
    "strconv"
)

type TypeKind int
const (
    Int         TypeKind = iota
    Uint        TypeKind = iota
    Char        TypeKind = iota
    Bool        TypeKind = iota
    Ptr         TypeKind = iota
    Arr         TypeKind = iota
    Vec         TypeKind = iota
    Str         TypeKind = iota
    Struct      TypeKind = iota
    Enum        TypeKind = iota
    Generic     TypeKind = iota
    Func        TypeKind = iota
    Interface   TypeKind = iota
    Infer       TypeKind = iota
)

const (
    I8_Size         uint = 1
    I16_Size        uint = 2
    I32_Size        uint = 4
    I64_Size        uint = 8

    U8_Size         uint = 1
    U16_Size        uint = 2
    U32_Size        uint = 4
    U64_Size        uint = 8

    Char_Size       uint = 1
    Bool_Size       uint = 1
    Ptr_Size        uint = 8
    Arr_Size        uint = 8
    Func_Size       uint = Ptr_Size
    Interface_Size  uint = 2 * Ptr_Size
    Str_Size        uint = Ptr_Size + U32_Size
    Vec_Size        uint = Ptr_Size + U64_Size + U64_Size
)

type Type interface {
    Size()          uint
    String()        string
    GetKind()       TypeKind

    // TODO: remove
    GetInterfaces() map[string]InterfaceType
}

type CharType struct { Interfaces map[string]InterfaceType }
type BoolType struct { Interfaces map[string]InterfaceType }
type UintType struct {
    size uint
    Interfaces map[string]InterfaceType
}
type IntType struct {
    size uint
    Interfaces map[string]InterfaceType
}
type PtrType struct {
    BaseType Type
    Interfaces map[string]InterfaceType
}
type ArrType struct {
    BaseType Type
    Len uint64
    Interfaces map[string]InterfaceType
}
type VecType struct {
    BaseType Type
    Interfaces map[string]InterfaceType
}
type StrType struct {
    Interfaces map[string]InterfaceType
}
type StructType struct {
    Name string
    Types []Type
    Interfaces map[string]InterfaceType
    names map[string]int
    isBigStruct bool
    isAligned bool
    size uint
}
type EnumType struct {
    Name string
    IdType Type
    Interfaces map[string]InterfaceType
    ids map[string]uint64
    types map[string]Type        // nil for no type
    size uint
    isBigStruct bool
}
type FuncType struct {
    Name string
    Args []Type
    Ret Type
    Generic *GenericType
    // TODO: isConst?
}
type GenericType struct {
    Name string
    CurUsedType Type
    UsedTypes []Type
    Interfaces map[string]InterfaceType
}
type InterfaceType struct {
    Name string
    Funcs []FuncType
}
type InferType struct {
    Idx uint64
    DefaultType Type
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

        case nil:

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
    return IntType{ size: intSize, Interfaces: make(map[string]InterfaceType) }
}

func CreateUint(uintSize uint) UintType {
    return UintType{ size: uintSize, Interfaces: make(map[string]InterfaceType) }
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

    return StructType{ 
        Name: name,
        Types: types,
        Interfaces: make(map[string]InterfaceType),
        isBigStruct: isBigStruct,
        isAligned: aligned,
        size: size,
        names: ns,
    }
}

func CreateEnumType(name string, idType Type, names []string, types []Type) EnumType {
    size := uint(0)
    for _,t := range types {
        if t != nil && t.Size() > size {
            size = t.Size()
        }
    }
    size += idType.Size()

    if len(names) != len(types) {
        fmt.Fprintln(os.Stderr, "[ERROR] (internal) CreateEnumType len names and types is not equal")
        os.Exit(1)
    }

    ts := make(map[string]Type)
    ids := make(map[string]uint64)
    for i,name := range names {
        ts[name] = types[i]
        ids[name] = uint64(i)
    }

    isBigStruct := false
    if size > 16 {
        isBigStruct = true
    }

    aligned,_ := isAligned(types, idType.Size())
    if !aligned {
        isBigStruct = true
    }

    return EnumType{ Name: name, Interfaces: make(map[string]InterfaceType), IdType: idType, types: ts, ids: ids, size: size, isBigStruct: isBigStruct }
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

func (t *EnumType) GetType(name string) Type {
    return t.types[name]
}

func (t *EnumType) GetTypeWithId(id uint64) Type {
    for name,i := range t.ids {
        if i == id {
            return t.types[name]
        }
    }

    return nil
}

func (t *EnumType) HasElem(name string) bool {
    _, ok := t.types[name]
    return ok
}

func (t *EnumType) GetElems() []string {
    res := make([]string, 0, len(t.types))

    for name := range t.types {
        res = append(res, name)
    }

    return res
}

func (t *EnumType) GetID(name string) uint64 {
    return t.ids[name]
}

func (t *InterfaceType) GetFunc(name string) *FuncType {
    for _,f := range t.Funcs {
        if f.Name == name {
            return &f
        }
    }

    return nil
}

func IsBigStruct(t Type) bool {
    if _,ok := t.(VecType); ok {
        return true
    }

    if t,ok := t.(StructType); ok {
        return t.isBigStruct
    }

    if t,ok := t.(EnumType); ok {
        return t.isBigStruct
    }

    return false
}

func IsResolvable(t Type) bool {
    if t,ok := t.(PtrType); ok {
        return IsResolvable(t.BaseType)
    }

    return t.GetKind() == Infer
}


func ReplaceGeneric(t Type) Type {
    if t,ok := t.(*GenericType); ok && t.CurUsedType != nil {
        return t.CurUsedType
    }

    return t
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

    case Interface:
        return 2

    default:
        return 1
    }
}

func (t IntType)        GetKind() TypeKind { return Int }
func (t UintType)       GetKind() TypeKind { return Uint }
func (t CharType)       GetKind() TypeKind { return Char }
func (t BoolType)       GetKind() TypeKind { return Bool }
func (t StrType)        GetKind() TypeKind { return Str  }
func (t PtrType)        GetKind() TypeKind { return Ptr  }
func (t ArrType)        GetKind() TypeKind { return Arr  }
func (t VecType)        GetKind() TypeKind { return Vec  }
func (t StructType)     GetKind() TypeKind { return Struct }
func (t EnumType)       GetKind() TypeKind { return Enum }
func (t InterfaceType)  GetKind() TypeKind { return Interface }
func (t FuncType)       GetKind() TypeKind { return Func }
func (t InferType)      GetKind() TypeKind { return Infer }
func (t GenericType)    GetKind() TypeKind {
    if t.CurUsedType != nil {
        return t.CurUsedType.GetKind()
    }

    return Generic
}

func (t IntType)        Size() uint { return t.size }
func (t UintType)       Size() uint { return t.size }
func (t CharType)       Size() uint { return Char_Size }
func (t BoolType)       Size() uint { return Bool_Size }
func (t StrType)        Size() uint { return Str_Size }
func (t PtrType)        Size() uint { return Ptr_Size }
func (t ArrType)        Size() uint { return Arr_Size }
func (t VecType)        Size() uint { return Vec_Size }
func (t StructType)     Size() uint { return t.size }
func (t EnumType)       Size() uint { return t.size }
func (t InterfaceType)  Size() uint { return Interface_Size }
func (t FuncType)       Size() uint { return Func_Size }
func (t InferType)      Size() uint { 
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) InferType Size() should never get called")
    os.Exit(1)
    return 0 
}
func (t GenericType) Size() uint {
    if t.CurUsedType != nil {
        return t.CurUsedType.Size()
    }

    return 0
}

func (t IntType)        GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t UintType)       GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t CharType)       GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t BoolType)       GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t StrType)        GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t PtrType)        GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t ArrType)        GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t VecType)        GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t StructType)     GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t EnumType)       GetInterfaces() map[string]InterfaceType { return t.Interfaces }
func (t InterfaceType) GetInterfaces() map[string]InterfaceType { return nil }
func (t FuncType) GetInterfaces() map[string]InterfaceType { return nil }
func (t GenericType)    GetInterfaces() map[string]InterfaceType {
    if t.CurUsedType != nil {
        return t.CurUsedType.GetInterfaces()
    }

    return t.Interfaces
}
func (t InferType) GetInterfaces() map[string]InterfaceType { 
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) InferType has no interfaces")
    os.Exit(1)
    return nil
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
    return fmt.Sprintf("[%d]%s", t.Len, t.BaseType)
}
func (t VecType) String() string {
    return "[$]" + t.BaseType.String()
}
func (t StructType) String() string { return t.Name }
func (t EnumType) String() string { return t.Name }
func (t InterfaceType) String() string { return t.Name }
func (t InferType) String() string { return t.DefaultType.String() }
func (t GenericType) String() string {
    if t.CurUsedType != nil {
        return t.CurUsedType.String()
    }
    return t.Name
}
func (t FuncType) String() string {
    generic := ""
    if t.Generic != nil {
        generic = fmt.Sprintf("<%s>", t.Generic)
    }

    ret := ""
    if t.Ret != nil {
        ret = fmt.Sprintf(" -> %s", t.Ret)
    }

    return fmt.Sprintf("%s%s(%v)%s", t.Name, generic, t.Args, ret)
}

func ToBaseType(s string) Type {
    switch s {
    case "i8":
        return IntType{ size: I8_Size, Interfaces: make(map[string]InterfaceType) }
    case "i16":
        return IntType{ size: I16_Size, Interfaces: make(map[string]InterfaceType) }
    case "i32":
        return IntType{ size: I32_Size, Interfaces: make(map[string]InterfaceType) }
    case "i64":
        return IntType{ size: I64_Size, Interfaces: make(map[string]InterfaceType) }
    case "u8":
        return UintType{ size: U8_Size, Interfaces: make(map[string]InterfaceType) }
    case "u16":
        return UintType{ size: U16_Size, Interfaces: make(map[string]InterfaceType) }
    case "u32":
        return UintType{ size: U32_Size, Interfaces: make(map[string]InterfaceType) }
    case "u64":
        return UintType{ size: U64_Size, Interfaces: make(map[string]InterfaceType) }
    case "char":
        return CharType{ Interfaces: make(map[string]InterfaceType) }
    case "bool":
        return BoolType{ Interfaces: make(map[string]InterfaceType) }
    case "str":
        return StrType{ Interfaces: make(map[string]InterfaceType) }
    default:
        return nil
    }
}

var inferIdx uint64 = 0

func TypeOfVal(val string) Type {
    switch {
    case val[0] == '"' && val[len(val) - 1] == '"':
        return StrType{ Interfaces: make(map[string]InterfaceType) }
    case val[0] == '\'' && val[len(val) - 1] == '\'':
        return CharType{ Interfaces: make(map[string]InterfaceType) }
    case val == "true" || val == "false":
        return BoolType{ Interfaces: make(map[string]InterfaceType) }
    case len(val) > 2 && val[:2] == "0x":
        if _, err := strconv.ParseUint(val, 0, 64); err == nil {
            t := InferType{ Idx: inferIdx, DefaultType: UintType{ size: U64_Size } } 
            inferIdx++
            return t
        }
    default:
        if _, err := strconv.ParseInt(val, 10, 32); err == nil {
            t := InferType{ Idx: inferIdx, DefaultType: IntType{ size: I32_Size } } 
            inferIdx++
            return t
        }
        if _, err := strconv.ParseInt(val, 10, 64); err == nil {
            t := InferType{ Idx: inferIdx, DefaultType: IntType{ size: I64_Size } } 
            inferIdx++
            return t
        }
        if _, err := strconv.ParseUint(val, 0, 64); err == nil {
            t := InferType{ Idx: inferIdx, DefaultType: UintType{ size: U64_Size } } 
            inferIdx++
            return t
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

func EqualBinary(t1 Type, t2 Type) bool {
    switch t := t1.(type) {
    case IntType:
        if t2,ok := t2.(IntType); ok {
            return t2.Size() == t.Size()
        }

    case UintType:
        if t2,ok := t2.(UintType); ok {
            return t2.Size() == t.Size()
        }
        if _,ok := t2.(PtrType); ok {
            return t1.Size() == Ptr_Size
        }

    case PtrType:
        if _,ok := t2.(PtrType); ok {
            return true
        }
        return Equal(CreateUint(Ptr_Size), t2)

    default:
        return Equal(t1, t2)
    }

    return false
}

func Equal(destType Type, srcType Type) bool {
    srcType = ReplaceGeneric(srcType)
    destType = ReplaceGeneric(destType)

    switch t := destType.(type) {
    case VecType:
        if t2,ok := srcType.(VecType); ok {
            return Equal(t.BaseType, t2.BaseType)
        }

    case ArrType:
        if t2,ok := srcType.(ArrType); ok {
            if t.Len == t2.Len {
                return Equal(t.BaseType, t2.BaseType)
            }
        }

    case PtrType:
        if t2,ok := srcType.(PtrType); ok {
            return Equal(t.BaseType, t2.BaseType)
        }

    case StructType:
        if t2,ok := srcType.(StructType); ok {
            return t.Name == t2.Name
        }

    case EnumType:
        if t2,ok := srcType.(EnumType); ok {
            return t.Name == t2.Name
        }

    case InterfaceType:
        if interfaces := srcType.GetInterfaces(); interfaces != nil {
            if t,ok := interfaces[t.Name]; ok {
                return Equal(destType, t)
            }
        } else if t2,ok := srcType.(InterfaceType); ok {
            return t.Name == t2.Name
        }

    case *GenericType:
        if t2,ok := srcType.(*GenericType); ok {
            return t.Name == t2.Name
        }

    case FuncType:
        if t2,ok := srcType.(FuncType); ok {
            if t.Name != t2.Name {
                return false
            }

            if t.Generic != nil && t2.Generic != nil {
                if !Equal(*t.Generic, *t2.Generic) {
                    return false
                }
            } else if !(t.Generic == nil && t2.Generic == nil) {
                return false
            }

            if !Equal(t.Ret, t2.Ret) {
                return false
            }

            if len(t.Args) == len(t2.Args) {
                for i := range t.Args {
                    if !Equal(t.Args[i], t2.Args[i]) {
                        return false
                    }
                }

                return true
            }
        }

    case IntType:
        if t2,ok := srcType.(IntType); ok {
            return t2.Size() <= destType.Size()
        }

    case UintType:
        if t2,ok := srcType.(UintType); ok {
            return t2.Size() <= destType.Size()
        }

    case nil:
        return srcType == nil

    default:
        return srcType != nil && destType.GetKind() == srcType.GetKind()
    }

    return false
}
