package array

import (
	"os"
	"fmt"
)

var arrLits []arrLit

type arrLit struct {
    size uint64
}

func Add(size uint64) (idx int) {
    idx = len(arrLits)
    arrLits = append(arrLits, arrLit{ size: size })
    return
}

func WriteArrayLits(file *os.File) {
    for i,a := range arrLits {
        file.WriteString(fmt.Sprintf("_arr%d: resb %d\n", i, a.size))
    }
}
