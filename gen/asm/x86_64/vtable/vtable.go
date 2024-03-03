package vtable

import (
	"fmt"
	"gamma/types"
	"gamma/gen/asm/x86_64"
	"gamma/gen/asm/x86_64/nasm"
)

func Create(implDstType types.Type, interfaceType types.InterfaceType, fnNames []string) {
    nasm.AddData(GetVTableName(implDstType, interfaceType) + ":")
    for _,name := range fnNames {
        nasm.AddData(fmt.Sprintf("%s %s.%s", asm.GetDataSize(types.Ptr_Size), implDstType.GetMangledName(), name))
    }
}

func GetVTableName(implDstType types.Type, interfaceType types.InterfaceType) string {
    return fmt.Sprintf("_vtable_%s_%s", implDstType.GetMangledName(), interfaceType.GetMangledName())
}
