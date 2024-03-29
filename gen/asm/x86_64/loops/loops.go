package loops

import (
    "fmt"
    "bufio"
    "gamma/types/addr"
)


var whileCount uint = 0
var forCount   uint = 0

var inForLoop bool = false
var inWhileLoop bool = false

func InLoop() bool {
    return inForLoop || inWhileLoop
}

func ResetCount() {
    whileCount = 0
    forCount   = 0
}

func WhileStart(file *bufio.Writer) uint {
    inWhileLoop = true
    whileCount++
    file.WriteString(fmt.Sprintf(".while%d:\n", whileCount))
    return whileCount
}

func WhileVar(file *bufio.Writer, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", addr))
    file.WriteString(fmt.Sprintf("jne .while%dEnd\n", whileCount))
}

func WhileExpr(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf("cmp al, 1\njne .while%dEnd\n", whileCount))
}

func WhileEnd(file *bufio.Writer, count uint) {
    inWhileLoop = false
    file.WriteString(fmt.Sprintf("jmp .while%d\n", count))
    file.WriteString(fmt.Sprintf(".while%dEnd:\n", count))
}


func ForStart(file *bufio.Writer) uint {
    inForLoop = true

    forCount++
    file.WriteString(fmt.Sprintf(".for%d:\n", forCount))
    return forCount
}

func ForExpr(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf("cmp al, 1\njne .for%dEnd\n", forCount))
}

func ForBlockEnd(file *bufio.Writer, count uint) {
    file.WriteString(fmt.Sprintf(".for%dBlockEnd:\n", count))
}
func ForEnd(file *bufio.Writer, count uint) {
    file.WriteString(fmt.Sprintf("jmp .for%d\n", count))
    file.WriteString(fmt.Sprintf(".for%dEnd:\n", count))

    inForLoop = false
}

func Break(file *bufio.Writer) {
    if inForLoop {
        file.WriteString(fmt.Sprintf("jmp .for%dEnd\n", forCount))
    } else {
        file.WriteString(fmt.Sprintf("jmp .while%dEnd\n", whileCount))
    }
}

func Continue(file *bufio.Writer) {
    if inForLoop {
        file.WriteString(fmt.Sprintf("jmp .for%dBlockEnd\n", forCount))
    } else {
        file.WriteString(fmt.Sprintf("jmp .while%d\n", whileCount))
    }
}
