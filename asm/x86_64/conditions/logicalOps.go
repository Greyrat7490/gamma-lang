package cond

import (
	"os"
	"fmt"
	"gamma/token"
)

var logCount int = 0

func LogicalOp(file *os.File, t token.Token) int {
    logCount++

    if t.Type == token.And {
        file.WriteString("cmp al, 0\n")
    } else if t.Type == token.Or {
        file.WriteString("cmp al, 1\n")
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not a valid logical op (expected && or ||)\n", t)
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)
    }

    file.WriteString(fmt.Sprintf("je .cond%d\n", logCount))

    return logCount
}

func LogicalOpEnd(file *os.File, count int) {
    file.WriteString(fmt.Sprintf(".cond%d\n", count))
}
