package vars

var scopes []Scope

type Scope struct {
    vars []LocalVar
    maxSize int
}

func InGlobalScope() bool {
    return len(scopes) == 0
}

func CreateScope() {
    scopes = append(scopes, Scope{})
}

func RemoveScope() {
    localVarOffset = 0

    if len(scopes) > 0 {
        scopes = scopes[:len(scopes)-1]
    }
}
