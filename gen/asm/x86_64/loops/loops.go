package loops

import (
    "os"
    "fmt"
)


var whileCount uint = 0
var forCount   uint = 0

var inForLoop bool = false

func ResetCount() {
    whileCount = 0
    forCount   = 0
}

func WhileStart(file *os.File) uint {
    whileCount++
    file.WriteString(fmt.Sprintf(".while%d:\n", whileCount))
    return whileCount
}

func WhileVar(file *os.File, addr string) {
    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", addr))
    file.WriteString(fmt.Sprintf("jne .while%dEnd\n", whileCount))
}

func WhileExpr(file *os.File) {
    file.WriteString(fmt.Sprintf("cmp al, 1\njne .while%dEnd\n", whileCount))
}

func WhileEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf("jmp .while%d\n", count))
    file.WriteString(fmt.Sprintf(".while%dEnd:\n", count))
}


func ForStart(file *os.File) uint {
    inForLoop = true

    forCount++
    file.WriteString(fmt.Sprintf(".for%d:\n", forCount))
    return forCount
}

func ForExpr(file *os.File) {
    file.WriteString(fmt.Sprintf("cmp al, 1\njne .for%dEnd\n", forCount))
}

func ForBlockEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf(".for%dBlockEnd:\n", count))
}
func ForEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf("jmp .for%d\n", count))
    file.WriteString(fmt.Sprintf(".for%dEnd:\n", count))

    inForLoop = false
}

func Break(file *os.File) {
    if inForLoop {
        file.WriteString(fmt.Sprintf("jmp .for%dEnd\n", forCount))
    } else {
        file.WriteString(fmt.Sprintf("jmp .while%dEnd\n", whileCount))
    }
}

func Continue(file *os.File) {
    if inForLoop {
        file.WriteString(fmt.Sprintf("jmp .for%dBlockEnd\n", forCount))
    } else {
        file.WriteString(fmt.Sprintf("jmp .while%d\n", whileCount))
    }
}
