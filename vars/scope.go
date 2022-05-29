package vars

var scopes []Scope
var curScope int = -1

type Scope struct {
    vars []LocalVar
    maxSize int
}

func InGlobalScope() bool {
    return curScope == -1
}

func CreateScope() {
    curScope = len(scopes)
    scopes = append(scopes, Scope{})
}

func RemoveScope() {
    localVarOffset = 0

    if len(scopes) > 0 {
        scopes = scopes[:len(scopes)-1]
        curScope--
    }
}
