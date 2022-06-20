package std

import (
    "os"
    "fmt"
)

const SYS_WRITE = 1
const SYS_EXIT = 60

// linux syscall calling convention
// arg: 0    1    2    3   4   5
//     rdi, rsi, rdx, r10, r8, r9
// return: rax
func syscall(file *os.File, syscallNum uint) {
    file.WriteString(fmt.Sprintf("mov rax, %d\n", syscallNum))

    file.WriteString("push rcx\n")
    file.WriteString("push r11\n")   // syscall can change r11 and rcx

    file.WriteString("syscall\n")

    file.WriteString("pop r11\n")
    file.WriteString("pop rcx\n")
}

func defineExit(file *os.File) {
    file.WriteString("exit:\n")
    file.WriteString(fmt.Sprintf("mov rax, %d\n", SYS_EXIT))
    file.WriteString("syscall\n\n")
}
