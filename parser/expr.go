package prs

import (
	"fmt"
	"gamma/ast"
	"gamma/ast/identObj"
	"gamma/cmpTime"
	"gamma/cmpTime/constVal"
	"gamma/token"
	"gamma/types"
	"gamma/types/array"
	"gamma/types/char"
	"gamma/types/str"
	"os"
	"reflect"
	"strconv"
)

type precedence int
const (
    LOGICAL_PRECEDENCE precedence = iota // &&, ||
    COMPARE_PRECEDENCE            = iota // ==, !=, <, <=, >, >=
    CAST_PRECEDENCE               = iota // as
    BITWISE_PRECEDENCE            = iota // <<, >>, &, |, ^, ~
    ADD_SUB_PRECEDENCE            = iota // +, -
    MUL_DIV_PRECEDENCE            = iota // *, /, %
    PAREN_PRECEDENCE              = iota // ()
)


func prsExpr(tokens *token.Tokens) ast.Expr {
    return prsExprWithPrecedence(tokens, 0)
}

func prsExprWithPrecedence(tokens *token.Tokens, precedence precedence) ast.Expr {
    var expr ast.Expr
    switch tokens.Cur().Type {
    case token.Number, token.Boolean, token.Char:
        expr = prsBasicLit(tokens)

    case token.Str:
        expr = prsStrLit(tokens)

    case token.BrackL:
        if tokens.Peek().Type == token.XSwitch {
            expr = prsVecLit(tokens)
        } else {
            expr = prsArrayLit(tokens)
        }

    case token.Typename:
        ident := prsIdentExpr(tokens)
        tokens.Next()
        expr = prsCallInterfaceFn(tokens, ident, nil)

    case token.Name, token.Self, token.SelfType:
        ident := prsIdentExpr(tokens)
        expr = prsPostNameExpr(tokens, ident, nil)

    case token.XSwitch:
        expr = prsXSwitch(tokens)

    case token.UndScr:
        expr = prsIdentExpr(tokens)

    case token.ParenL:
        expr = prsParenExpr(tokens)

    case token.Plus, token.Minus, token.Mul, token.Amp, token.BitNot:
        expr = prsUnaryExpr(tokens)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression (got \"%v\")\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)

        return &ast.BadExpr{}
    }

    for {
        if isBinaryExpr(tokens) {
            expr = prsBinary(tokens, expr, precedence)
        }

        if tokens.Peek().Type == token.BrackL {
            tokens.Next()
            expr = prsIndexExpr(tokens, expr)
        } else if tokens.Peek().Type == token.Dot {
            tokens.Next()
            expr = prsDotExpr(tokens, expr)
        } else if tokens.Peek().Type == token.As && CAST_PRECEDENCE >= precedence {
            tokens.Next()
            expr = prsCast(tokens, expr)
        } else {
            break
        }
    }

    return expr
}

func isUnaryExpr(tokens *token.Tokens) bool {
    return  tokens.Cur().Type == token.Plus || tokens.Cur().Type == token.Minus ||
            tokens.Cur().Type == token.Mul  || tokens.Cur().Type == token.Amp   ||
            tokens.Cur().Type == token.BitNot
}

func isParenExpr(tokens *token.Tokens) bool {
    return  tokens.Cur().Type == token.ParenL
}

func isBinaryExpr(tokens *token.Tokens) bool {
    if tokens.Peek().Pos.Line > tokens.Cur().Pos.Line {
        return false
    }

    return  tokens.Peek().Type == token.Plus || tokens.Peek().Type == token.Minus  ||
            tokens.Peek().Type == token.Mul  || tokens.Peek().Type == token.Div    ||
            tokens.Peek().Type == token.Mod  ||
            tokens.Peek().Type == token.And  || tokens.Peek().Type == token.Or     ||
            isBitwise(tokens)                ||
            isComparison(tokens)
}

func isBitwise(tokens *token.Tokens) bool {
    return tokens.Peek().Type == token.Amp  || tokens.Peek().Type == token.BitOr  ||
           tokens.Peek().Type == token.Xor  || tokens.Peek().Type == token.BitNot ||
           tokens.Peek().Type == token.Shl  || tokens.Peek().Type == token.Shr
}

func isComparison(tokens *token.Tokens) bool {
    return  tokens.Peek().Type == token.Eql || tokens.Peek().Type == token.Neq ||
            tokens.Peek().Type == token.Grt || tokens.Peek().Type == token.Lss ||
            tokens.Peek().Type == token.Geq || tokens.Peek().Type == token.Leq
}

func getPrecedence(tokens *token.Tokens) precedence {
    switch {
    case tokens.Peek().Type == token.And || tokens.Peek().Type == token.Or:
        return LOGICAL_PRECEDENCE
    case isComparison(tokens):
        return COMPARE_PRECEDENCE
    case tokens.Peek().Type == token.As:
        return CAST_PRECEDENCE
    case tokens.Peek().Type == token.Plus || tokens.Peek().Type == token.Minus:
        return ADD_SUB_PRECEDENCE
    case tokens.Peek().Type == token.Mul || tokens.Peek().Type == token.Div || tokens.Peek().Type == token.Mod:
        return MUL_DIV_PRECEDENCE
    case isBitwise(tokens):
        return BITWISE_PRECEDENCE
    case isParenExpr(tokens):
        return PAREN_PRECEDENCE
    default:
        return precedence(0)
    }
}

func prsName(tokens *token.Tokens) token.Token {
    name := tokens.Cur()

    if name.Type == token.SelfType {
        if identObj.CurSelfType == nil {
            fmt.Fprintln(os.Stderr, "[ERROR] Self used outside of impl and interface")
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

        name.Str = identObj.CurSelfType.String()
        name.Type = token.Name
    }

    if name.Type == token.Self {
        if identObj.CurSelfType == nil {
            fmt.Fprintln(os.Stderr, "[ERROR] self used outside of impl and interface")
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

        name.Type = token.Name
    }

    if name.Type != token.Name && name.Type != token.UndScr && name.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    return name
}

func checkDefined(ident *ast.Ident) {
    if ident.Obj == nil && ident.Name == "_" {
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not defined\n", ident.Name)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }
}

func prsPostNameExpr(tokens *token.Tokens, ident *ast.Ident, insetType types.Type) ast.Expr {
    switch tokens.Peek().Type {
    case token.ParenL:
        tokens.Next()
        return prsCallFn(tokens, ident, insetType)

    case token.Dot:
        checkDefined(ident)
        tokens.Next()
        // TODO generic
        return prsDotExpr(tokens, ident)

    case token.BraceL:
        checkDefined(ident)
        if isStruct(ident.Obj) {
            tokens.Next()
            return prsStructLit(tokens, ident, insetType)
        }

    case token.DefConst:
        tokens.Next()
        if tokens.Peek().Type == token.Lss {
            insetType := prsInsetType(tokens)
            return prsPostNameExpr(tokens, ident, insetType)

        } else if tokens.Peek().Type == token.Name {
            checkDefined(ident)
            if isEnumLit(ident.Name, tokens.Peek().Str) {
                return prsEnumLit(tokens, ident, insetType)
            } else {
                return prsCallInterfaceFn(tokens, ident, insetType)
            }

        } else {
            if isGenericFunc(ident.Name) {
                fmt.Fprintf(os.Stderr, "[ERROR] expected \"<\" after \"::\" for a generic function but got %v\n", tokens.Peek())
                fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())

            } else if isStruct(ident.Obj) {
                fmt.Fprintf(os.Stderr, "[ERROR] expected an interface func name after \"::\" for a struct but got %v\n", tokens.Peek())
                fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())

            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] unexpected \"::\"")
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            }

            os.Exit(1)
        }
    }

    if ident.Obj == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is not defined\n", ident.Name)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }
    
    return ident
}

func prsIdentExpr(tokens *token.Tokens) *ast.Ident {
    // if wildcard ("_")
    if tokens.Cur().Type == token.UndScr {
        return &ast.Ident{ Name: tokens.Cur().Str, Pos: tokens.Cur().Pos, Obj: nil }
    }

    ident := prsName(tokens)
    obj := identObj.Get(ident.Str)

    return &ast.Ident{ Name: ident.Str, Pos: ident.Pos, Obj: obj }
}

func prsBasicLit(tokens *token.Tokens) ast.Expr {
    val := tokens.Cur()
    t := types.TypeOfVal(val.Str)

    switch t.GetKind() {
    case types.Bool:
        return &ast.BoolLit{ Repr: val.Str == "true", Val: val }

    case types.Char:
        var repr uint8
        if val.Str[1] == '\\' {
            var ok bool
            repr,ok = char.EscapeByte(val.Str[2])
            if !ok {
                fmt.Fprintf(os.Stderr, "[ERROR] unexpected escape sequence %s\n", val.Str)
                os.Exit(1)
            }
        } else {
            repr = uint8(val.Str[1])
        }

        return &ast.CharLit{ Repr: repr, Val: val }

    case types.Int, types.Uint, types.Infer:
        repr,_ := strconv.ParseUint(val.Str, 0, 64)
        return &ast.IntLit{ Repr: repr, Val: val, Type: t }

    default:
        return &ast.BadExpr{}
    }
}

func prsStrLit(tokens *token.Tokens) *ast.StrLit {
    val := tokens.Cur()

    idx := str.Add(val)
    return &ast.StrLit{ Idx: uint64(idx), Val: val }
}

func prsArrayLit(tokens *token.Tokens) *ast.ArrayLit {
    pos := tokens.Cur().Pos

    typ := prsArrType(tokens)

    braceLPos := tokens.Next().Pos
    lit := prsArrayLitExprs(tokens, typ)
    braceRPos := tokens.Cur().Pos

    lit.BraceLPos = braceLPos
    lit.BraceRPos = braceRPos
    lit.Pos = pos
    lit.Idx = array.Add(typ, constEvalExprs(lit.Values))
    return &lit
}

func prsVecLit(tokens *token.Tokens) *ast.VectorLit {
    lit := ast.VectorLit{ Pos: tokens.Cur().Pos }
    lit.Type = prsVecType(tokens)

    lit.BraceLPos = tokens.Next().Pos
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    tokens.Next()
    if tokens.Cur().Type == token.Name && tokens.Peek().Type == token.Colon {
        prsVecLitField(tokens, &lit)
        if tokens.Peek().Type == token.Comma {
            tokens.Next()
            tokens.Next()
            prsVecLitField(tokens, &lit)
        }

        if lit.Cap == nil {
            lit.Cap = &ast.Binary{ 
                OperandL: lit.Len, 
                OperandR: &ast.IntLit{
                    Type: types.CreateUint(types.Ptr_Size), 
                    Repr: 2,
                    Val: token.Token{Type: token.Number, Str: "2"},
                },
                Type: types.CreateUint(types.Ptr_Size),
                Operator: token.Token{Type: token.Mul, Str: "*"},
            }
        }
    } else if tokens.Cur().Type != token.BraceR {
        lit.Cap = prsExpr(tokens)
    }

    lit.BraceRPos = tokens.Next().Pos
    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    return &lit
}

func prsVecLitField(tokens *token.Tokens, lit *ast.VectorLit) {
    switch tokens.Cur().Str {
    case "cap":
        tokens.Next()
        tokens.Next()
        lit.Cap = prsExpr(tokens)
    case "len":
        tokens.Next()
        tokens.Next()
        lit.Len = prsExpr(tokens)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] vec has no field \"%s\" (only len and cap)\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
}

func prsStructLit(tokens *token.Tokens, ident *ast.Ident, insetType types.Type) *ast.StructLit {
    s := ident.Obj.(*identObj.Struct)
    t := types.ReplaceGeneric(ident.GetType(), insetType).(types.StructType)

    braceL := tokens.Cur()
    if braceL.Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    if tokens.Peek().Type == token.BraceR {
        return &ast.StructLit{
            Pos: ident.Pos,
            StructType: t,
            BraceLPos: braceL.Pos,
            BraceRPos: tokens.Next().Pos,
            Fields: []ast.FieldLit{},
        }
    }

    omitNames := omitNames(tokens)

    var fields []ast.FieldLit
    if tokens.Next().Type != token.BraceR {
        f := prsFieldLit(tokens, omitNames)
        fields = append(fields, f)

        for tokens.Next().Type == token.Comma {
            tokens.Next()
            f := prsFieldLit(tokens, omitNames)
            fields = append(fields, f)
        }
    }

    if !omitNames {
        orderedFields := make([]ast.FieldLit, len(fields))
        for _,f := range fields {
            if idx := t.GetFieldNum(f.Name.Str); idx == -1 {
                fmt.Fprintf(os.Stderr, "[ERROR] struct \"%s\" has no field called \"%s\"\n", ident.Name, f.Name.Str)
                fmt.Fprintf(os.Stderr, "\tfields: %v\n", s.GetFieldNames())
                fmt.Fprintln(os.Stderr, "\t" + f.At())
                os.Exit(1)
            } else {
                orderedFields[idx] = f
            }
        }
        fields = orderedFields
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if len(fields) != len(t.Types) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %d fields for struct \"%s\" but got %d\n", len(t.Types), t.Name, len(fields))
        fmt.Fprintf(os.Stderr, "\texpected: %v\n", t.Types)
        fmt.Fprintf(os.Stderr, "\tgot:      %v\n", fieldsToTypes(fields))
        fmt.Fprintln(os.Stderr, "\t" + braceL.Pos.At())
        os.Exit(1)
    }

    return &ast.StructLit{
        Pos: ident.Pos,
        StructType: t,
        BraceLPos: braceL.Pos,
        BraceRPos: tokens.Cur().Pos,
        Fields: fields,
    }
}

func omitNames(tokens *token.Tokens) bool {
    return tokens.Peek().Type != token.Name || tokens.Peek2().Type != token.Colon
}

func fieldsToTypes(fields []ast.FieldLit) []types.Type {
    res := make([]types.Type, len(fields))
    for i, f := range fields {
        res[i] = f.GetType()
    }

    return res
}

func prsFieldLit(tokens *token.Tokens, omitNames bool) ast.FieldLit {
    var name token.Token
    var pos token.Pos

    if !omitNames {
        name = prsName(tokens)

        if tokens.Next().Type != token.Colon {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a \":\" but got %v\n", tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }
        pos = name.Pos

        tokens.Next()
    } else {
        pos = tokens.Cur().Pos
    }

    return ast.FieldLit{ Name: name, Pos: pos, Value: prsExpr(tokens) }
}

func constEvalExprs(values []ast.Expr) []constVal.ConstVal {
    res := make([]constVal.ConstVal, 0, len(values))

    for _,v := range values {
        res = append(res, cmpTime.ConstEvalArrWithNils(v))
    }

    return res
}

func prsArrayLitExprs(tokens *token.Tokens, t types.ArrType) ast.ArrayLit {
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    exprs := []ast.Expr{}

    if tokens.Next().Type == token.BraceR {
        return ast.ArrayLit{ Values: exprs, Type: t, Idx: ^uint64(0) }
    }

    parsedLen := uint64(1)
    if baseType,ok := t.BaseType.(types.ArrType); ok {
        es := prsArrayLitExprs(tokens, baseType)
        exprs = append(exprs, &es)

        for tokens.Next().Type == token.Comma {
            if tokens.Next().Type == token.BraceR { break } // trailing comma
            es := prsArrayLitExprs(tokens, baseType)
            exprs = append(exprs, &es)
            parsedLen++
        }
    } else {
        exprs = append(exprs, prsExpr(tokens))

        for tokens.Next().Type == token.Comma {
            if tokens.Next().Type == token.BraceR { break } // trailing comma
            exprs = append(exprs, prsExpr(tokens))
            parsedLen++
        }
    }

    // check missing ,
    if parsedLen < t.Len && tokens.Cur().Type != token.Comma && tokens.Cur().Type != token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
        fmt.Fprintf(os.Stderr, "\tarray type: %v\n", t)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Last().At())
        os.Exit(1)
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return ast.ArrayLit{ Values: exprs, Type: t, Idx: ^uint64(0) }
}

func prsIndexExpr(tokens *token.Tokens, e ast.Expr) *ast.Indexed {
    res := ast.Indexed{ ArrExpr: e, BrackLPos: tokens.Cur().Pos, ArrType: e.GetType() }

    tokens.Next()
    res.Index = prsExpr(tokens)

    posR := tokens.Next()
    if posR.Type != token.BrackR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"]\" but got %v\n", posR)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    switch t := res.ArrType.(type) {
    case types.ArrType:
        res.Type = t.BaseType
    case types.VecType:
        res.Type = t.BaseType
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] you cannot index %v", t)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    return &res
}

func getStructFromExpr(expr ast.Expr) *identObj.Struct {
    typ := expr.GetType()

    if t,ok := typ.(types.StructType); ok {
        if s,ok := identObj.Get(t.Name).(*identObj.Struct); ok {
            return s           
        }
    } else if _,ok := typ.(*types.StructType); ok {
        if s,ok := identObj.Get(t.Name).(*identObj.Struct); ok {
            return s           
        }
    }

    return nil
}

func prsDotCallFn(tokens *token.Tokens, obj ast.Expr, dotPos token.Pos, typ types.Type, name token.Token, f *identObj.Func) ast.Expr {
    tokens.Next()
    insetType := prsInsetType(tokens)

    posL := tokens.Cur().Pos
    vals := prsPassArgs(tokens)
    posR := tokens.Cur().Pos

    vals = addSelfArg(vals, f, obj)

    if insetType != nil {
        if !f.IsGeneric() {
            fmt.Fprintf(os.Stderr, "[ERROR] %s (from %s) is not generic\n", name.Str, typ)
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }
        f.AddTypeToGeneric(insetType)
    }

    ident := ast.Ident{ Name: name.Str, Pos: name.Pos, Obj: f }
    return &ast.FnCall{ 
        Ident: ident, FnSrc: typ, F: f, InsetType: insetType,
        Values: vals, ParenLPos: posL, ParenRPos: posR, 
    }
}

func prsDotField(tokens *token.Tokens, t types.Type, obj ast.Expr, dotPos token.Pos, name token.Token) *ast.Field {
    switch typ := t.(type) {
    case *types.StructType:
        field := &ast.Field{ Obj: obj, DotPos: dotPos, FieldName: name }
        field.StructType = *typ
        field.Type = field.StructType.GetType(field.FieldName.Str)
        return field
    case types.StructType:
        field := &ast.Field{ Obj: obj, DotPos: dotPos, FieldName: name }
        field.StructType = typ
        field.Type = field.StructType.GetType(field.FieldName.Str)
        return field

    case types.ArrType:
        field := &ast.Field{ Obj: obj, DotPos: dotPos, FieldName: name }
        field.Type = types.CreateUint(types.U64_Size)
        return field

    case types.VecType:
        field := &ast.Field{ Obj: obj, DotPos: dotPos, FieldName: name }
        field.Type = types.CreateUint(types.U64_Size)
        return field

    case types.StrType:
        field := &ast.Field{ Obj: obj, DotPos: dotPos, FieldName: name }
        field.Type = types.CreateUint(types.U32_Size)
        return field

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] type %s has no field/function called %s\n", typ, name.Str)
        fmt.Fprintln(os.Stderr, "\t" + obj.At())
        os.Exit(1)
        return nil
    }
}

func prsDotExpr(tokens *token.Tokens, obj ast.Expr) ast.Expr {
    dot := tokens.Cur()
    if dot.Type != token.Dot {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \".\" but got %v\n", dot)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    t := obj.GetType()
    obj, t = autoDeref(obj, name, t)

    if f := getInterfaceFunc(t, name.Str); f != nil {
        return prsDotCallFn(tokens, obj, dot.Pos, t, name, f)
    } else {
        return prsDotField(tokens, t, obj, dot.Pos, name)
    }
}

func getInterfaceFunc(t types.Type, name string) *identObj.Func {
    if impl := identObj.GetImplObj(t.String()); impl != nil {
        return impl.GetFunc(name)
    }

    if I,ok := identObj.Get(t.String()).(*identObj.Interface); ok {
        return I.GetFunc(name)
    }

    return nil
}

func autoDeref(obj ast.Expr, field token.Token, t types.Type) (derefedObj ast.Expr, baseType types.Type)  {
    if identObj.HasFunc(t, field.Str) {
        return obj, t
    }

    if typ, ok := t.(types.PtrType); ok {
        derefedObj := &ast.Unary{ 
            Type: typ.BaseType,
            Operator: token.Token{ Pos: field.Pos,
            Type: token.Mul, Str: "*" },
            Operand: obj,
        }

        return autoDeref(derefedObj, field, typ.BaseType)
    } else {
        return obj, t
    }
}

func prsEnumLit(tokens *token.Tokens, ident *ast.Ident, insetType types.Type) *ast.EnumLit {
    if tokens.Cur().Type != token.DefConst {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"::\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    elemName := tokens.Next()
    if elemName.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", elemName)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    var content *ast.Paren = nil
    if tokens.Peek().Type == token.ParenL {
        tokens.Next()
        content = prsParenExpr(tokens)
    }

    enumType := types.ReplaceGeneric(ident.GetType(), insetType).(types.EnumType)
    t := enumType.GetType(elemName.Str)
    return &ast.EnumLit{ Pos: ident.Pos, Type: enumType, ElemName: elemName, ContentType: t, Content: content }
}

func prsUnwrapElem(tokens *token.Tokens, unwrap *ast.Unwrap) *ast.Unwrap {
    if unwrap.ElemName.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", unwrap.ElemName)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.ParenL {
        parenLPos := tokens.Next().Pos

        tokens.Next()
        ident := prsName(tokens)

        if tokens.Next().Type != token.ParenR {
            fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }
        parenRPos := tokens.Cur().Pos

        t := unwrap.EnumType.GetType(unwrap.ElemName.Str)
        ununsedObj := ident.Type == token.UndScr

        var obj identObj.IdentObj = nil
        if !ununsedObj {
            if c := cmpTime.ConstEval(unwrap.SrcExpr); c != nil {
                if c,ok := c.(*constVal.EnumConst); ok {
                    obj = identObj.DecConst(ident, t, c.Elem)
                }
            } else {
                obj = identObj.DecVar(ident, t) 
            }
        }

        unwrap.ParenLPos = parenLPos
        unwrap.ParenRPos = parenRPos
        unwrap.Obj = obj
        unwrap.UnusedObj = ununsedObj
    }

    return unwrap
}

func prsUnwrapHead(tokens *token.Tokens, srcExpr ast.Expr) *ast.Unwrap {
    colonPos := tokens.Next().Pos

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    var insetType types.Type = nil
    if tokens.Peek2().Type == token.Lss {
        tokens.Next()
        insetType = prsInsetType(tokens)
    }

    if tokens.Next().Type != token.DefConst {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"::\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    var enum *identObj.Enum = nil
    if obj := identObj.Get(name.Str); obj != nil {
        if e,ok := obj.(*identObj.Enum); ok {
            enum = e
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not an enum (got %v)\n", name.Str, reflect.TypeOf(obj))
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] enum \"%s\" is not defined\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
    enumType := types.ReplaceGeneric(enum.GetType(), insetType).(types.EnumType)

    return &ast.Unwrap{ SrcExpr: srcExpr, ColonPos: colonPos, EnumType: enumType }
}

func prsUnwrap(tokens *token.Tokens, srcExpr ast.Expr) *ast.Unwrap {
    unwrap := prsUnwrapHead(tokens, srcExpr)

    elemName := tokens.Next()
    if elemName.Type == token.BraceL {
        return unwrap
    } else {
        unwrap.ElemName = elemName
        return prsUnwrapElem(tokens, unwrap)
    }
}

func addSelfArg(values []ast.Expr, f *identObj.Func, obj ast.Expr) []ast.Expr {
    values = append(values, nil)
    copy(values[1:], values)

    expectedSelfType := f.GetArgs()[0]

    if expectedSelfType.GetKind() == types.Ptr &&  obj.GetType().GetKind() != types.Ptr {
        values[0] = &ast.Unary { 
            Type: types.PtrType{ BaseType: obj.GetType() },
            Operator: token.Token{ Type: token.Amp, Str: "&" },
            Operand: obj,
        }
    } else {
        values[0] = obj
    }

    return values
}

func prsParenExpr(tokens *token.Tokens) *ast.Paren {
    expr := ast.Paren{ ParenLPos: tokens.Cur().Pos }

    tokens.Next()
    expr.Expr = prsExpr(tokens)

    expr.ParenRPos = tokens.Next().Pos

    if tokens.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return &expr
}

func prsUnaryExpr(tokens *token.Tokens) *ast.Unary {
    expr := ast.Unary{ Operator: tokens.Cur() }

    switch expr.Operator.Type {
    case token.Mul:
        tokens.Next()
        if tokens.Cur().Type == token.ParenL {
            expr.Operand = prsParenExpr(tokens)
        } else if tokens.Cur().Type == token.Mul {
            expr.Operand = prsUnaryExpr(tokens)
        } else {
            expr.Operand = prsIdentExpr(tokens)
        }
    case token.Amp:
        tokens.Next()
        expr.Operand = prsIdentExpr(tokens)
        if tokens.Peek().Type == token.Dot {
            expr.Operand = prsDotExpr(tokens, expr.Operand)
            for tokens.Peek().Type == token.Dot {
                expr.Operand = prsDotExpr(tokens, expr.Operand)
            }
        }
    case token.BitNot:
        tokens.Next()
        expr.Operand = prsExpr(tokens)
    default:
        switch tokens.Next().Type {
        case token.Name:
            expr.Operand = prsIdentExpr(tokens)
        case token.ParenL:
            expr.Operand = prsParenExpr(tokens)
        default:
            expr.Operand = prsBasicLit(tokens)
        }
    }

    expr.Type = getTypeUnary(&expr)
    return &expr
}

func prsCaseCond(tokens *token.Tokens, condBase ast.Expr, placeholder *ast.Expr) (conds ast.Expr, colonPos token.Pos) {
    if tokens.Cur().Type == token.Colon {
        if tokens.Last().Pos.Line == tokens.Cur().Pos.Line {
            fmt.Fprintln(os.Stderr, "[ERROR] missing case body for this case")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Last2().Pos.At())
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        }
        os.Exit(1)
    }
    if tokens.Cur().Type == token.Comma {
        fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \",\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    if tokens.Last().Pos.Line == tokens.Cur().Pos.Line && tokens.Last().Type != token.SemiCol && tokens.Last().Type != token.BraceL {
        fmt.Fprintln(os.Stderr, "[ERROR] cases should always start in a new line or after a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    cond := prsExpr(tokens)
    conds = completeCond(placeholder, condBase, cond, conds)

    for tokens.Next().Type == token.Comma {
        if tokens.Peek().Type == token.Colon || tokens.Peek().Type == token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: no expr after \",\"")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        tokens.Next()
        cond = prsExpr(tokens)
        conds = completeCond(placeholder, condBase, cond, conds)
    }

    if tokens.Cur().Type != token.Colon {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \":\" at the end of case condition")
        fmt.Fprintln(os.Stderr, "\t" + conds.End())
        os.Exit(1)
    }
    colonPos = tokens.Cur().Pos

    return
}

func prsXCaseUnwrap(tokens *token.Tokens, condBase ast.Expr) ast.XCase {
    cond, colonPos := prsUnwrapCaseCond(tokens, condBase)

    tokens.Next()
    expr := prsExpr(tokens)

    if colonPos.Line == tokens.Peek().Pos.Line && tokens.Peek().Type != token.SemiCol && tokens.Peek().Type != token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.SemiCol { tokens.Next() }

    return ast.XCase{ Cond: cond, ColonPos: colonPos, Expr: expr }
}

func prsXCase(tokens *token.Tokens, condBase ast.Expr, placeholder *ast.Expr) ast.XCase {
    cond, colonPos := prsCaseCond(tokens, condBase, placeholder)

    tokens.Next()
    expr := prsExpr(tokens)

    if colonPos.Line == tokens.Peek().Pos.Line && tokens.Peek().Type != token.SemiCol && tokens.Peek().Type != token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Peek().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.SemiCol { tokens.Next() }

    return ast.XCase{ Cond: cond, ColonPos: colonPos, Expr: expr }
}

func prsXCases(tokens *token.Tokens, condBase ast.Expr) (cases []ast.XCase) {
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" at the end of conditon for the xswitch (got \"%s\")\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if _,ok := condBase.(*ast.Unwrap); ok { 
        for tokens.Next().Type != token.BraceR {
            cases = append(cases, prsXCaseUnwrap(tokens, condBase))
        }
    } else {
        placeholder := getPlaceholder(condBase)
        for tokens.Next().Type != token.BraceR {
            cases = append(cases, prsXCase(tokens, condBase, placeholder))
        }
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" at the end of the switch (got \"%s\")\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return
}

func prsXSwitch(tokens *token.Tokens) *ast.XSwitch {
    switchExpr := ast.XSwitch{ Pos: tokens.Cur().Pos }

    var condBase ast.Expr = nil
    if tokens.Next().Type != token.BraceL {
        condBase = prsExpr(tokens)
    }

    if tokens.Peek().Type == token.BraceR {
        fmt.Fprintln(os.Stderr, "[ERROR] empty switch")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.Colon {
        condBase = prsUnwrapHead(tokens, condBase)
        tokens.Next()
    }

    identObj.StartScope()
    switchExpr.BraceLPos = tokens.Cur().Pos
    switchExpr.Cases = prsXCases(tokens, condBase)
    switchExpr.BraceRPos = tokens.Cur().Pos
    identObj.EndScope()

    switchExpr.Type = switchExpr.Cases[0].GetType()
    return &switchExpr
}

func prsBinary(tokens *token.Tokens, expr ast.Expr, min_precedence precedence) ast.Expr {
    for isBinaryExpr(tokens) && getPrecedence(tokens) >= min_precedence {
        var b ast.Binary
        b.OperandL = expr

        precedence := getPrecedence(tokens)
        b.Operator = tokens.Next()

        tokens.Next()

        // switch/xswitch
        if tokens.Cur().Type == token.BraceL {
            b.Type = types.BoolType{}
            return &b
        }

        b.OperandR = prsExprWithPrecedence(tokens, precedence+1)
        b.Type = getTypeBinary(&b)

        expr = &b
    }

    return expr
}

func prsInsetType(tokens *token.Tokens) types.Type {
    if tokens.Cur().Type == token.DefConst {
        if tokens.Next().Type != token.Lss {
            fmt.Fprintf(os.Stderr, "[ERROR] expected \"<\" but got %v\n", tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        tokens.Next()
        typ := prsType(tokens)

        if tokens.Next().Type != token.Grt {
            fmt.Fprintf(os.Stderr, "[ERROR] expected \">\" but got %v\n", tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        return typ
    }

    return nil
}

func prsCallInterfaceFn(tokens *token.Tokens, ident *ast.Ident, insetType types.Type) *ast.FnCall {
    if tokens.Cur().Type != token.DefConst {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"::\" but got %s\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    name := tokens.Next()

    posL := tokens.Next().Pos
    vals := prsPassArgs(tokens)
    posR := tokens.Cur().Pos

    var f *identObj.Func = nil
    if obj,ok := ident.Obj.(*identObj.Interface); ok {
        f = obj.GetFunc(name.Str)
        f = f.FromNewFnSrc(vals[0].GetType())

    } else if obj := identObj.GetImplObj(ident.Name); obj != nil {
        f = obj.GetFunc(name.Str)

    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] expected an interface or implementable obj before %s but got %s\n", ident.Name, ident.GetType())
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if f != nil {
        if insetType != nil {
            if !f.IsGeneric() {
                fmt.Fprintf(os.Stderr, "[ERROR] function %s is not generic\n", ident.Name)
                fmt.Fprintln(os.Stderr, "\t" + ident.At())
                os.Exit(1)
            }
            f.AddTypeToGeneric(insetType)
        }

        resvSpace := identObj.ReserveSpace(f.GetRetType())

        fnSrc := ident.GetType()
        ident := ast.Ident{ Name: name.Str, Pos: name.Pos, Obj: f }
        return &ast.FnCall{ Ident: ident, FnSrc: fnSrc, F: f, ResvSpace: resvSpace, InsetType: insetType, Values: vals, ParenLPos: posL, ParenRPos: posR }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %s does not implement function %s\n", ident.Name, name.Str)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    return nil
}

func prsCallFn(tokens *token.Tokens, ident *ast.Ident, insetType types.Type) *ast.FnCall {
    posL := tokens.Cur().Pos
    vals := prsPassArgs(tokens)
    posR := tokens.Cur().Pos

    if ident.Obj != nil {
        if f,ok := ident.Obj.(*identObj.Func); ok {
            if insetType != nil {
                if !f.IsGeneric() {
                    fmt.Fprintf(os.Stderr, "[ERROR] function %s is not generic\n", ident.Name)
                    fmt.Fprintln(os.Stderr, "\t" + ident.At())
                    os.Exit(1)
                }
                f.AddTypeToGeneric(insetType)
            }

            resvSpace := identObj.ReserveSpace(f.GetRetType())

            return &ast.FnCall{ Ident: *ident, F: f, ResvSpace: resvSpace, InsetType: insetType, Values: vals, ParenLPos: posL, ParenRPos: posR }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only call a function (%s is not a function)\n", ident.Name)
            fmt.Fprintln(os.Stderr, "\t" + ident.At())
            os.Exit(1)
        }
    }

    f := identObj.CreateUnresolvedFunc(ident.Name)
    return &ast.FnCall{ Ident: *ident, F: &f, Values: vals, ParenLPos: posL, ParenRPos: posR }
}

func prsPassArgs(tokens *token.Tokens) []ast.Expr {
    if tokens.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    var values []ast.Expr

    if tokens.Next().Type == token.ParenR {
        return values
    }

    values = append(values, prsExpr(tokens))
    for tokens.Next().Type == token.Comma {
        tokens.Next()
        values = append(values, prsExpr(tokens))
    }

    if tokens.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return values
}

func prsCast(tokens *token.Tokens, e ast.Expr) *ast.Cast {
    c := ast.Cast{ Expr: e, AsPos: tokens.Cur().Pos }

    tokens.Next()
    c.DestType = prsType(tokens)

    return &c
}
