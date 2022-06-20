package std

import (
    "os"
    "gorec/func"
    "gorec/types"
)

func Declare() {
    fn.AddBuildIn("printStr",  "s", types.StrType{})
    fn.AddBuildIn("printInt",  "i", types.I32Type{})
    fn.AddBuildIn("printPtr",  "i", types.PtrType{})
    fn.AddBuildIn("printBool", "b", types.BoolType{})

    fn.AddBuildIn("exit", "i", types.I32Type{})
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
