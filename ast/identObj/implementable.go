package identObj

type Implementable interface {
    IdentObj
    AddImpl(impl Impl)
    GetFunc(name string) *Func
    GetFuncNames() []string
}
