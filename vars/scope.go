package vars

var scopes []Scope
var curScope int = -1

type Scope struct {
    localStartIdx int
    maxSize int
}

func InGlobalScope() bool {
    return curScope == -1
}

func CreateScope() {
    curScope = len(scopes)
    scopes = append(scopes, Scope{ maxSize: 0, localStartIdx: len(vars) })
}

func RemoveScope() {
    removeLocalVars()

    if len(scopes) > 0 {
        scopes = scopes[:len(scopes)-1]
        curScope--
    }
}
