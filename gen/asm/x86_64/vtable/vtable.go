package vtable

import (
	"fmt"
	"gamma/types"
	"gamma/gen/asm/x86_64"
	"gamma/gen/asm/x86_64/nasm"
)

func Create(implName string, fnNames []string) {
    nasm.AddData(GetVTableName(implName) + ":")
    for _,name := range fnNames {
        nasm.AddData(fmt.Sprintf("%s %s.%s", asm.GetDataSize(types.Ptr_Size), implName, name))
    }
}

func GetVTableName(implName string) string {
    return fmt.Sprintf("_vtable_%s", implName)
}
