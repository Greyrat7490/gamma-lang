package vtable

import (
	"fmt"
	"gamma/types"
	"gamma/gen/asm/x86_64"
	"gamma/gen/asm/x86_64/nasm"
)

func Create(implName string, interfaceName string, fnNames []string) {
    nasm.AddData(GetVTableName(implName, interfaceName) + ":")
    for _,name := range fnNames {
        nasm.AddData(fmt.Sprintf("%s %s.%s", asm.GetDataSize(types.Ptr_Size), implName, name))
    }
}

func GetVTableName(implName string, interfaceName string) string {
    return fmt.Sprintf("_vtable_%s_%s", implName, interfaceName)
}
