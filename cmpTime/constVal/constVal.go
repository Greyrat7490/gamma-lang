package constVal

import (
    "os"
    "fmt"
    "gamma/types"
    "gamma/types/addr"
)

type ConstVal interface {
    GetKind() types.TypeKind
    GetVal() string
}

type IntConst int64;
type UintConst uint64;
type CharConst uint8;
type BoolConst bool;
type StrConst uint64;
type ArrConst  struct {
    Idx uint64
    Type types.ArrType
    Elems []ConstVal
}

type StructConst struct {
    Fields []ConstVal
}

type PtrConst struct {
    Addr addr.Addr
    Local bool
}

func (c *IntConst)    GetKind() types.TypeKind { return types.Int }
func (c *UintConst)   GetKind() types.TypeKind { return types.Uint }
func (c *CharConst)   GetKind() types.TypeKind { return types.Char }
func (c *BoolConst)   GetKind() types.TypeKind { return types.Bool }
func (c *StrConst)    GetKind() types.TypeKind { return types.Str }
func (c *ArrConst)    GetKind() types.TypeKind { return types.Arr }
func (c *StructConst) GetKind() types.TypeKind { return types.Struct }
func (c *PtrConst)    GetKind() types.TypeKind { return types.Ptr }


func (c *IntConst)    GetVal() string { return fmt.Sprint(int64(*c)) }
func (c *UintConst)   GetVal() string { return fmt.Sprint(uint64(*c)) }
func (c *CharConst)   GetVal() string { return fmt.Sprint(uint8(*c)) }
func (c *BoolConst)   GetVal() string { if bool(*c) { return "1" } else { return "0" } }
func (c *StrConst)    GetVal() string { return fmt.Sprintf("_str%d", uint64(*c)) }
func (c *ArrConst)    GetVal() string { return fmt.Sprintf("_arr%d", c.Idx) }
func (c *PtrConst)    GetVal() string { return c.Addr.String() }

func (c *StructConst) GetVal() string {
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) StructConst.GetVal() got called")
    os.Exit(1)
    return ""
}
