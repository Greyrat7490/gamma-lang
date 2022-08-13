package std

import (
    "os"
    "gamma/types"
    "gamma/ast/identObj"
)

func Declare() {
    identObj.AddBuildIn("printStr",  "s", types.StrType{}, nil)
    identObj.AddBuildIn("printInt",  "i", types.I32Type{}, nil)
    identObj.AddBuildIn("printPtr",  "i", types.PtrType{}, nil)
    identObj.AddBuildIn("printBool", "b", types.BoolType{}, nil)

    identObj.AddBuildIn("exit", "i", types.I32Type{}, nil)
}

func Define(file *os.File) {
    defineItoS(file)
    defineBtoS(file)

    definePrintStr(file)
    definePrintInt(file)
    definePrintPtr(file)
    definePrintBool(file)

    defineExit(file)
}
