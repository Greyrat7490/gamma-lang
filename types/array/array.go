package array

import (
	"fmt"
	"gorec/asm/x86_64"
	"gorec/asm/x86_64/nasm"
	"gorec/token"
	"gorec/types"
	"gorec/types/str"
	"os"
)

var arrLits []arrLit

type arrLit struct {
    typ types.Type
    len uint64
    values []token.Token
}

func Add(typ types.Type, length uint64, values []token.Token) (i int) {
    i = len(arrLits)
    arr := arrLit{ typ: typ, len: length, values: values }
    arrLits = append(arrLits, arr)

    if len(arr.values) == 0 {
        nasm.AddBss(fmt.Sprintf("_arr%d: %s %d", i, asm.GetBssSize(arr.typ.Size()), arr.len))
    } else {
        switch typ.GetKind() {
        case types.Str:
            d1size := asm.GetDataSize(types.Ptr_Size)
            d2size := asm.GetDataSize(types.I32_Size)

            res := fmt.Sprintf("_arr%d:", i)
            for _, v := range arr.values {
                strIdx := str.Add(v)

                res += fmt.Sprintf("\n  %s _str%d", d1size, strIdx)
                res += fmt.Sprintf("\n  %s %d", d2size, str.GetSize(strIdx))
            }
            nasm.AddData(res)

        case types.Bool, types.I32, types.Ptr, types.Arr:
            dsize := asm.GetDataSize(arr.typ.Size())

            res := fmt.Sprintf("_arr%d:", i)
            for _, v := range arr.values {
                res += fmt.Sprintf("\n  %s %s", dsize, v.Str)
            }
            nasm.AddData(res)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", typ)
            os.Exit(1)
        }
    }

    return
}
