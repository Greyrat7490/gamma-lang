package types

import (
	"fmt"
	"os"
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
    Size()           uint
    String()         string
    GetMangledName() string
    GetKind()        TypeKind
}

type CharType struct {}
type BoolType struct {}
type UintType struct { size uint }
type IntType struct { size uint }
type PtrType struct { BaseType Type }
type ArrType struct {
    BaseType Type
    Len uint64
}
type VecType struct { BaseType Type }
type StrType struct { } 
type StructType struct {
    Name string
    Types []Type
    genericName string          // empty string means not generic
    insetType Type
    names map[string]int
    isBigStruct bool
    size uint
}
type EnumType struct {
    Name string
    IdType Type
    genericName string          // empty string means not generic
    insetType Type
    ids map[string]uint64
    types map[string]Type       // nil for no type
    size uint
    isBigStruct bool
}
type FuncType struct {
    Name string
    Args []Type
    Ret Type
    Generic GenericType
    // TODO: isConst?
}
type GenericType struct {
    Name string
    SetType Type
    Idx uint64
}
type InterfaceType struct {
    Name string
    Funcs []FuncType
    Generic GenericType
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
            if size != 0 {
                return false, 0
            }

        case StrType:
            if size != 0 {
                return false, 0
            }
            size += U32_Size

        case nil:

        default:
            if size != 0 && size < t.Size() {
                return false, 0
            }
            size += t.Size()
        }

        if size >= 8 {
            size -= 8
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

func CreateEmptyStructType(name string) StructType {
    return StructType{ Name: name }
}
func CreateStructType(name string, types []Type, names []string, genericName string) StructType {
    ns := map[string]int{}
    for i, n := range names {
        ns[n] = i
    }

    size := uint(0)
    for _,t := range types {
        size += t.Size()
    }

    isBigStruct := true
    if size <= 16 {
        aligned,_ := isAligned(types, 0)
        isBigStruct = !aligned
    }

    return StructType{ 
        Name: name,
        Types: types,
        isBigStruct: isBigStruct,
        genericName: genericName,
        size: size,
        names: ns,
    }
}

func CreateEnumType(name string, idType Type, names []string, types []Type, genericName string) EnumType {
    size := uint(0)
    isBigStruct := false
    for _,t := range types {
        if t == nil { continue }

        isBigStruct = true
        if t.Size() > size { 
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

    return EnumType{ 
        Name: name,
        IdType: idType,
        types: ts,
        ids: ids,
        size: size,
        isBigStruct: isBigStruct,
        genericName: genericName,
    }
}

var inferIdx uint64 = 0

func CreateInferType(defaultType Type) InferType {
    t := InferType{ DefaultType: defaultType, Idx: inferIdx } 
    inferIdx++
    return t
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

func (t *StructType) GetInsetType() Type {
    return t.insetType
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

func (t *EnumType) GetInsetType() Type {
    return t.insetType
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

func IsSelfType(t Type, interfaceType InterfaceType) bool {
    switch t := t.(type) {
    case InterfaceType:
        return Equal(t, interfaceType)

    case PtrType:
        return IsSelfType(t.BaseType, interfaceType)

    default:
        return false
    }
}

func IsFnSrcResolvable(src Type, fnName string) bool {
    if src,ok := src.(InterfaceType); ok {
        f := src.GetFunc(fnName)
        if len(f.Args) > 0 {
            return IsSelfType(f.Args[0], src)
        }
    }

    return false
}

func ResolveFnSrc(src Type, fnName string, firstArgType Type) Type {
    if IsFnSrcResolvable(src, fnName) && !IsGeneric(firstArgType) {
        return firstArgType
    }

    return src
}

func IsBigStruct(t Type) bool {
    switch t := t.(type) {
    case VecType:
        return true
    case StructType:
        return t.isBigStruct
    case EnumType:
        return t.isBigStruct
    case *GenericType:
        if ResolveGeneric(t) != nil {
            return IsBigStruct(ResolveGeneric(t))
        }
    }

    return false
}

func IsResolvable(t Type) bool {
    switch t := t.(type) {
    case PtrType:
        return IsResolvable(t.BaseType)
    case ArrType:
        return IsResolvable(t.BaseType)
    case VecType:
        return IsResolvable(t.BaseType)
    case InterfaceType:
        return IsResolvable(t.Generic.SetType)
    case StructType:
        return IsResolvable(t.insetType)
    case EnumType:
        return IsResolvable(t.insetType)
    case InferType:
        return true
    default:
        return false
    }
}

func SolveGeneric(typeWithGeneric Type, srcType Type) Type {
    switch t1 := typeWithGeneric.(type) {
    case PtrType:
        if t2,ok := srcType.(PtrType); ok {
            return SolveGeneric(t1.BaseType, t2.BaseType)
        }

    case ArrType:
        if t2,ok := srcType.(ArrType); ok {
            return SolveGeneric(t1.BaseType, t2.BaseType)
        }

    case VecType:
        if t2,ok := srcType.(VecType); ok {
            return SolveGeneric(t1.BaseType, t2.BaseType)
        }

    case GenericType, *GenericType:
        if t2,ok := srcType.(GenericType); ok {
            return ResolveGeneric(t2)
        }
        if t2,ok := srcType.(*GenericType); ok {
            return ResolveGeneric(t2)
        }

        return srcType
    }

    return nil
}

func IsGeneric(t Type) bool {
    switch t := t.(type) {
    case PtrType:
        return IsGeneric(t.BaseType)

    case ArrType:
        return IsGeneric(t.BaseType)

    case VecType:
        return IsGeneric(t.BaseType)

    case InterfaceType:
        return t.Generic.Name != ""

    case StructType:
        return t.genericName != ""

    case EnumType:
        return t.genericName != ""

    case GenericType, *GenericType:
        return true
    }

    return false
}

func ReplaceGeneric(t Type, insetType Type) Type {
    if insetType == nil || insetType == (*GenericType)(nil) {
        return t
    }

    switch t := t.(type) {
    case PtrType:
        t.BaseType = ReplaceGeneric(t.BaseType, insetType)
        return t

    case ArrType:
        t.BaseType = ReplaceGeneric(t.BaseType, insetType)
        return t

    case VecType:
        t.BaseType = ReplaceGeneric(t.BaseType, insetType)
        return t

    case InterfaceType:
        if t.Generic.Name != "" {
            SetCurInsetType(t.Generic, insetType)  
        }
        return t

    case EnumType:
        if t.insetType == nil || IsGeneric(t.insetType) {
            t.insetType = insetType

            ts := make(map[string]Type)
            for name, elemType := range t.types {
                ts[name] = ReplaceGeneric(elemType, insetType)
            }

            t.types = ts
            t.size = t.IdType.Size() + insetType.Size()
        }

        return t

    case StructType:
        if t.insetType == nil || IsGeneric(t.insetType) {
            t.insetType = insetType

            ts := make([]Type, len(t.Types))
            size := uint(0)
            for i := range t.Types {
                t2 := ReplaceGeneric(t.Types[i], insetType)
                ts[i] = t2
                size += t2.Size()
            }
            t.size = size
            t.Types = ts

            if !t.isBigStruct {
                aligned,_ := isAligned(t.Types, 0)
                t.isBigStruct = aligned && t.Size() <= 16
            }
        }

        return t

    case GenericType:
        return insetType
    case *GenericType:
        return insetType

    default:
        return t
    }
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
    if ResolveGeneric(t) != nil {
        return ResolveGeneric(t).GetKind()
    }
    // TODO: always generic
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
    if ResolveGeneric(t) != nil {
        return ResolveGeneric(t).Size()
    }

    return 0
}

func (t InferType) GetInterfaces() map[string]InterfaceType { 
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) InferType has no interfaces")
    os.Exit(1)
    return nil
}

func (t IntType) String() string {
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
func (t UintType) String() string {
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
func (t StrType) String() string { return "str"  }
func (t PtrType) String() string {
    return fmt.Sprintf("*%v", t.BaseType)
}
func (t ArrType) String() string {
    return fmt.Sprintf("[%d]%s", t.Len, t.BaseType)
}
func (t VecType) String() string {
    return "[$]" + t.BaseType.String()
}
func (t StructType) String() string { 
    if t.genericName != "" {
        if t.insetType != nil {
            return fmt.Sprintf("%s<%s>", t.Name, t.insetType)
        }
        return fmt.Sprintf("%s<%s>", t.Name, t.genericName)
    }
    return t.Name
}
func (t EnumType) String() string { 
    if t.genericName != "" {
        if t.insetType != nil {
            return fmt.Sprintf("%s<%s>", t.Name, t.insetType)
        }
        return fmt.Sprintf("%s<%s>", t.Name, t.genericName)
    }
    return t.Name
}
func (t InterfaceType) String() string { 
    if t.Generic.Name != "" {
        return fmt.Sprintf("%s<%s>", t.Name, t.Generic)
    }
    return t.Name 
}
func (t InferType) String() string { return t.DefaultType.String() }
func (t GenericType) String() string {
    if ResolveGeneric(t) != nil {
        return ResolveGeneric(t).String()
    }
    return t.Name
}
func (t FuncType) String() string {
    generic := ""
    if t.Generic.Name != "" {
        generic = fmt.Sprintf("<%s>", t.Generic)
    }

    ret := ""
    if t.Ret != nil {
        ret = fmt.Sprintf(" -> %s", t.Ret)
    }

    return fmt.Sprintf("%s%s(%v)%s", t.Name, generic, t.Args, ret)
}


func (t IntType)        GetMangledName() string { return t.String() }
func (t UintType)       GetMangledName() string { return t.String() }
func (t CharType)       GetMangledName() string { return t.String() }
func (t BoolType)       GetMangledName() string { return t.String() }
func (t StrType)        GetMangledName() string { return t.String() }
func (t InferType)      GetMangledName() string { return t.String() }
func (t PtrType)        GetMangledName() string { return "$ptr_" + t.BaseType.GetMangledName() }
func (t ArrType)        GetMangledName() string { return "$arr_" + t.BaseType.GetMangledName() }
func (t VecType)        GetMangledName() string { return "$vec_" + t.BaseType.GetMangledName() }
func (t InterfaceType)  GetMangledName() string { 
    if t.Generic.Name != "" {
        return t.Generic.GetMangledName() + "$" + t.Name
    } 

    return t.String() 
}
func (t GenericType)    GetMangledName() string {
    if ResolveGeneric(t) != nil {
        return ResolveGeneric(t).GetMangledName()
    }
    return t.Name
}
func (t StructType)     GetMangledName() string { 
    if t.genericName != "" {
        if t.insetType != nil {
            return fmt.Sprintf("%s$%s", t.Name, t.insetType)
        }
        return fmt.Sprintf("%s$%s", t.Name, t.genericName)
    }
    return t.Name
}
func (t EnumType)       GetMangledName() string {
    if t.genericName != "" {
        if t.insetType != nil {
            return fmt.Sprintf("%s$%s", t.Name, t.insetType)
        }
        return fmt.Sprintf("%s$%s", t.Name, t.genericName)
    }
    return t.Name
}
func (t FuncType)       GetMangledName() string { 
    generic := ""
    if t.Generic.Name != "" {
        generic = "$gen_" + t.Generic.GetMangledName()
    }

    ret := ""
    if t.Ret != nil {
        ret = "$ret_" + t.Ret.GetMangledName()
    }

    args := ""
    if len(t.Args) > 0 {
        args = "$args_"
    }
    for _,a := range t.Args {
        args += a.GetMangledName()
    }

    return t.Name + generic + args + ret
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
    case len(val) > 2 && val[:2] == "0x":
        if _, err := strconv.ParseUint(val, 0, 64); err == nil {
            return CreateInferType(UintType{ size: U64_Size })
        }
    default:
        if _, err := strconv.ParseInt(val, 10, 32); err == nil {
            return CreateInferType(IntType{ size: I32_Size })
        }
        if _, err := strconv.ParseInt(val, 10, 64); err == nil {
            return CreateInferType(IntType{ size: I64_Size })
        }
        if _, err := strconv.ParseUint(val, 0, 64); err == nil {
            return CreateInferType(UintType{ size: U64_Size })
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

func EqualCustom(destType Type, srcType Type, interfaceCompareFn func(Type, Type)bool) bool {
    if s := ResolveGeneric(srcType); s != nil {
        srcType = s
    }
    if d := ResolveGeneric(destType); d != nil {
        destType = d
    }

    switch t := destType.(type) {
    case VecType:
        if t2,ok := srcType.(VecType); ok {
            return EqualCustom(t.BaseType, t2.BaseType, interfaceCompareFn)
        }

    case ArrType:
        if t2,ok := srcType.(ArrType); ok {
            if t.Len == t2.Len {
                return EqualCustom(t.BaseType, t2.BaseType, interfaceCompareFn)
            }
        }

    case PtrType:
        if t2,ok := srcType.(PtrType); ok {
            return EqualCustom(t.BaseType, t2.BaseType, interfaceCompareFn)
        }

    case StructType:
        if t2,ok := srcType.(StructType); ok {
            return t.Name == t2.Name && EqualCustom(t.insetType, t2.insetType, interfaceCompareFn)
        }

    case EnumType:
        if t2,ok := srcType.(EnumType); ok {
            return t.Name == t2.Name && EqualCustom(t.insetType, t2.insetType, interfaceCompareFn)
        }

    case InterfaceType:
        if t2,ok := srcType.(InterfaceType); ok {
            return t.Name == t2.Name
        }

        return interfaceCompareFn(t, srcType)

    case GenericType:
        if t2,ok := srcType.(GenericType); ok {
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

            if !EqualCustom(t.Generic, t2.Generic, interfaceCompareFn) {
                return false
            }

            if !EqualCustom(t.Ret, t2.Ret, interfaceCompareFn) {
                return false
            }

            if len(t.Args) != len(t2.Args) {
                return false
            }
            for i := range t.Args {
                if !EqualCustom(t.Args[i], t2.Args[i], interfaceCompareFn) {
                    return false
                }
            }

            return true
        }

    case IntType:
        if t2,ok := srcType.(IntType); ok {
            return t2.Size() <= destType.Size()
        }

    case UintType:
        if t2,ok := srcType.(UintType); ok {
            return t2.Size() <= destType.Size()
        }

    case InferType:
        if t2,ok := srcType.(InferType); ok {
            return t.Idx == t2.Idx
        }

    case nil:
        return srcType == nil

    default:
        return srcType != nil && destType.GetKind() == srcType.GetKind()
    }

    return false
}

func Equal(destType Type, srcType Type) bool {
    return EqualCustom(destType, srcType, func(t1 Type, t2 Type)bool{ return false })
}
