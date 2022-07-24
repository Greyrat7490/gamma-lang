package std

import (
    "os"
    "gorec/types"
    "gorec/identObj/scope"
)

func Declare() {
    scope.AddBuildIn("printStr",  "s", types.StrType{})
    scope.AddBuildIn("printInt",  "i", types.I32Type{})
    scope.AddBuildIn("printPtr",  "i", types.PtrType{})
    scope.AddBuildIn("printBool", "b", types.BoolType{})

    scope.AddBuildIn("exit", "i", types.I32Type{})
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
