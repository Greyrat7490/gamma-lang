package cond

import (
	"os"
	"fmt"
	"gorec/token"
	"gorec/vars"
)

var logCount int = 0

func LogicalOp(file *os.File, t token.Token) int {
    logCount++

    if t.Type == token.And {
        vars.Write(file, "cmp al, 0\n")
    } else if t.Type == token.Or {
        vars.Write(file, "cmp al, 1\n")
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\"(%s) is not a valid logical op (expected && or ||)\n", t.Str, t.Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)
    }

    vars.Write(file, fmt.Sprintf("je .cond%d\n", logCount))

    return logCount
}

func LogicalOpEnd(file *os.File, count int) {
    vars.Write(file, fmt.Sprintf(".cond%d\n", count))
}
