package addr

import "fmt"

type Addr struct {
    BaseAddr string
    Offset int64
}

func (a Addr) String() string {
    if a.Offset > 0 {
        return fmt.Sprintf("%s+%d", a.BaseAddr, a.Offset)
    } else if a.Offset < 0 {
        return fmt.Sprintf("%s%d", a.BaseAddr, a.Offset)
    } else {
        return a.BaseAddr
    }
}

func (a Addr) Offseted(offset int64) Addr {
    a.Offset += offset
    return a
}
