package array

import (
    "os"
    "fmt"
    "gorec/types"
    "gorec/asm/x86_64"
    "gorec/asm/x86_64/nasm"
)

var arrLits []arrLit

type arrLit struct {
    typ types.Type
    len uint64
    values []string
}

func Add(typ types.Type, length uint64, values []string) (i int) {
    switch typ.GetKind() {
    case types.Str, types.Ptr, types.Arr:
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not supported yet (in work)\n", typ)
        os.Exit(1)
    }

    i = len(arrLits)
    arr := arrLit{ typ: typ, len: length, values: values }
    arrLits = append(arrLits, arr)

    if len(arr.values) == 0 {
        nasm.AddBss(fmt.Sprintf("_arr%d: %s %d", i, asm.GetBssSize(arr.typ.Size()), arr.len))
    } else {
        reserve := asm.GetDataSize(arr.typ.Size())

        res := fmt.Sprintf("_arr%d:", i)
        for _, v := range arr.values {
            res += fmt.Sprintf("\n  %s %s", reserve, v)
        }

        nasm.AddData(res)
    }

    return
}
