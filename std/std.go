package std

import (
    "os"
    "gorec/types"
    "gorec/identObj"
)

func Declare() {
    identObj.AddBuildIn("printStr",  "s", types.StrType{})
    identObj.AddBuildIn("printInt",  "i", types.I32Type{})
    identObj.AddBuildIn("printPtr",  "i", types.PtrType{})
    identObj.AddBuildIn("printBool", "b", types.BoolType{})

    identObj.AddBuildIn("exit", "i", types.I32Type{})
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
