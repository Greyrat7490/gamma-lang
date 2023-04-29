package prs

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/types/char"
    "gamma/types/array"
    "gamma/cmpTime"
    "gamma/cmpTime/constVal"
    "gamma/ast"
    "gamma/ast/identObj"
)

type precedence int
const (
    LOGICAL_PRECEDENCE precedence = iota // &&, ||
    COMPARE_PRECEDENCE            = iota // ==, !=, <, <=, >, >=
    BITWISE_PRECEDENCE            = iota // <<, >>, &, |, ^, ~
    ADD_SUB_PRECEDENCE            = iota // +, -
    MUL_DIV_PRECEDENCE            = iota // *, /, %
    PAREN_PRECEDENCE              = iota // ()
)

func prsExpr(tokens *token.Tokens) ast.Expr {
    var expr ast.Expr
    switch tokens.Cur().Type {
    case token.Number, token.Boolean, token.Char:
        expr = prsBasicLit(tokens)

    case token.Str:
        expr = prsStrLit(tokens)

    case token.BrackL:
        if tokens.Peek().Type == token.XSwitch {
            return prsVecLit(tokens)
        } else {
            return prsArrayLit(tokens)
        }

    case token.Name:
        switch tokens.Peek().Type {
        case token.ParenL:
            expr = prsCallFn(tokens)

        case token.DefConst:
            expr = prsCallGenericFn(tokens)

        case token.Dot:
            ident := prsIdentExpr(tokens)
            expr = prsField(tokens, ident)
            for tokens.Peek().Type == token.Dot {
                expr = prsField(tokens, expr)
            }

        case token.BraceL:
            if obj := identObj.Get(tokens.Cur().Str); obj != nil {
                if _,ok := obj.(*identObj.Struct); ok {
                    return prsStructLit(tokens)
                }
            }
            fallthrough

        default:
            expr = prsIdentExpr(tokens)
        }

    case token.XSwitch:
        expr = prsXSwitch(tokens)

    case token.UndScr:
        expr = prsIdentExpr(tokens)

    case token.ParenL:
        expr = prsParenExpr(tokens)

    case token.Plus, token.Minus, token.Mul, token.Amp, token.BitNot:
        expr = prsUnaryExpr(tokens)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression (got type %v)\n", tokens.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)

        return &ast.BadExpr{}
    }

    for tokens.Peek().Type == token.BrackL {
        tokens.Next()
        expr = prsIndexExpr(tokens, expr)
    }

    for tokens.Peek().Type == token.Dot {
        expr = prsField(tokens, expr)
    }

    for tokens.Peek().Type == token.As {
        tokens.Next()
        expr = prsCast(tokens, expr)
    }

    if isBinaryExpr(tokens) {
        expr = prsBinary(tokens, expr, 0)
        for tokens.Peek().Type == token.As {
            tokens.Next()
            expr = prsCast(tokens, expr)
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

func prsIdentExpr(tokens *token.Tokens) *ast.Ident {
    ident := tokens.Cur()

    // if wildcard ("_")
    if ident.Type == token.UndScr {
        return &ast.Ident{ Name: "_", Pos: ident.Pos, Obj: nil }
    }

    if ident.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", ident)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if obj := identObj.Get(ident.Str); obj != nil {
        return &ast.Ident{ Name: ident.Str, Pos: ident.Pos, Obj: obj }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", ident.Str)
    fmt.Fprintln(os.Stderr, "\t" + ident.At())
    os.Exit(1)
    return nil
}

func prsBasicLit(tokens *token.Tokens) ast.Expr {
    val := tokens.Cur()
    t := types.TypeOfVal(val.Str)

    switch t.GetKind() {
    case types.Bool:
        repr := false
        if val.Str == "true" {
            repr = true
        }

        return &ast.BoolLit{ Repr: repr, Val: val }

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

    case types.Int:
        repr,_ := strconv.ParseInt(val.Str, 0, 64)
        return &ast.IntLit{ Repr: repr, Val: val, Type: t.(types.IntType) }

    case types.Uint:
        repr,_ := strconv.ParseUint(val.Str, 0, 64)
        return &ast.UintLit{ Repr: repr, Val: val, Type: t.(types.UintType) }

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
    lit := ast.ArrayLit{ Pos: tokens.Cur().Pos }

    lit.Type = prsArrType(tokens)

    lit.BraceLPos = tokens.Next().Pos
    lit.Values = prsArrayLitExprs(tokens, lit.Type, 0)
    lit.BraceRPos = tokens.Cur().Pos

    lit.Idx = array.Add(lit.Type, constEvalExprs(lit.Values))
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

func prsStructLit(tokens *token.Tokens) *ast.StructLit {
    name := tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    var t types.StructType
    var s *identObj.Struct
    if obj := identObj.Get(name.Str); obj != nil {
        if strct,ok := obj.(*identObj.Struct); ok {
            t = strct.GetType().(types.StructType)
            s = strct
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] struct %s is not defined\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    braceL := tokens.Next()
    if braceL.Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    if tokens.Peek().Type == token.BraceR {
        return &ast.StructLit{
            Pos: name.Pos,
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
                fmt.Fprintf(os.Stderr, "[ERROR] struct \"%s\" has no field called \"%s\"\n", name.Str, f.Name.Str)
                fmt.Fprintf(os.Stderr, "\tfields: %v\n", s.GetNames())
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
        Pos: name.Pos,
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
        name = tokens.Cur()
        if name.Type != token.Name {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

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
    res := make([]constVal.ConstVal, len(values))

    for i,v := range values {
        res[i] = cmpTime.ConstEval(v)
    }

    return res
}

func prsArrayLitExprs(tokens *token.Tokens, t types.ArrType, depth int) (exprs []ast.Expr) {
    parsedLen := uint64(1)

    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Peek().Type == token.BraceR {
        tokens.Next()
        return
    }

    if len(t.Lens) - depth == 1 {
        tokens.Next()
        exprs = append(exprs, prsExpr(tokens))

        for tokens.Next().Type == token.Comma {
            if tokens.Next().Type == token.BraceR { break } // trailing comma
            exprs = append(exprs, prsExpr(tokens))
            parsedLen++
        }

    } else {
        tokens.Next()
        es := prsArrayLitExprs(tokens, t, depth+1)
        for _, e := range es {
            exprs = append(exprs, e)
        }

        for tokens.Next().Type == token.Comma {
            if tokens.Next().Type == token.BraceR { break } // trailing comma
            es := prsArrayLitExprs(tokens, t, depth+1)
            for _, e := range es {
                exprs = append(exprs, e)
            }
            parsedLen++
        }
    }

    // check literal length and missing characters
    if parsedLen != t.Lens[depth] {
        if parsedLen > t.Lens[depth] {
            fmt.Fprintf(os.Stderr, "[ERROR] too big array literal (expected len %d, but got %d)\n", t.Lens[depth], parsedLen)
            fmt.Fprintf(os.Stderr, "\tarray type: %v\n", t)
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        } else {
            switch tokens.Cur().Type {
            case token.Comma, token.BraceR:
                fmt.Fprintf(os.Stderr, "[ERROR] too small array literal (expected len %d, but got %d)\n", t.Lens[depth], parsedLen)
                fmt.Fprintf(os.Stderr, "\tarray type: %v\n", t)
                fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            default:
                fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
                fmt.Fprintf(os.Stderr, "\tarray type: %v\n", t)
                fmt.Fprintln(os.Stderr, "\t" + tokens.Last().At())
            }
        }
        os.Exit(1)
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return
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
        if len(t.Lens) == 1 {
            res.Type = t.BaseType
        } else {
            res.Type = types.ArrType{ BaseType: t.BaseType, Lens: t.Lens[1:] }
        }
    case types.VecType:
        res.Type = t.BaseType
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] you cannot index %v", t)
        fmt.Fprintln(os.Stderr, "\t" + e.At())
        os.Exit(1)
    }

    return &res
}

func prsField(tokens *token.Tokens, obj ast.Expr) *ast.Field {
    dot := tokens.Next()
    if dot.Type != token.Dot {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \".\" but got %v\n", dot)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    fieldName := tokens.Next()
    if fieldName.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", fieldName)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    field := ast.Field{ Obj: obj, DotPos: dot.Pos, FieldName: fieldName }
    setFieldType(&field)

    return &field
}

func setFieldType(field *ast.Field) {
    switch t := field.Obj.GetType().(type) {
    case types.ArrType:
        field.Type = types.CreateUint(types.U64_Size)
    case types.VecType:
        field.Type = types.CreateUint(types.U64_Size)
    case types.StrType:
        field.Type = types.CreateUint(types.U32_Size)
    case types.StructType:
        field.StructType = t
        field.Type = field.StructType.GetType(field.FieldName.Str)

        if field.Type == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] struct \"%s\" has no field called \"%s\"\n", field.StructType.Name, field.FieldName.Str)
            fmt.Fprintf(os.Stderr, "\tfields: %v\n", t.GetFields())
            fmt.Fprintln(os.Stderr, "\t" + field.At())
            os.Exit(1)
        }
    // auto deref
    case types.PtrType:
        field.Obj = &ast.Unary{ Type: t.BaseType, Operator: token.Token{ Pos: field.FieldName.Pos, Type: token.Mul, Str: "*" }, Operand: field.Obj }
        setFieldType(field)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] type %s has no fields\n", t)
        fmt.Fprintln(os.Stderr, "\t" + field.Obj.At())
        os.Exit(1)
    }
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
            expr.Operand = prsField(tokens, expr.Operand)
            for tokens.Peek().Type == token.Dot {
                expr.Operand = prsField(tokens, expr.Operand)
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

    expr.Type = GetTypeUnary(&expr)
    return &expr
}

func prsCaseExpr(tokens *token.Tokens, condBase ast.Expr, placeholder *ast.Expr, lastCaseEnd token.Pos) (caseExpr ast.XCase) {
    if tokens.Cur().Type == token.Colon {
        if tokens.Last().Pos.Line == tokens.Cur().Pos.Line {
            fmt.Fprintln(os.Stderr, "[ERROR] missing case body(expr) for this case")
            fmt.Fprintln(os.Stderr, "\t" + lastCaseEnd.At())
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

    // parse case cond(s) ----------------
    expr := prsExpr(tokens)
    var conds ast.Expr = nil
    for tokens.Next().Type == token.Comma {
        conds = completeCond(placeholder, condBase, expr, conds)

        if tokens.Peek().Type == token.Colon || tokens.Peek().Type == token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: no expr after \",\"")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        tokens.Next()
        expr = prsExpr(tokens)
    }

    caseExpr.ColonPos = tokens.Cur().Pos
    caseExpr.Cond = completeCond(placeholder, condBase, expr, conds)

    if tokens.Cur().Type != token.Colon {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \":\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    if nextColon := tokens.FindNext(token.Colon); tokens.Cur().Pos.Line == nextColon.Line {
        nextSemiCol := tokens.FindNext(token.SemiCol)

        if nextSemiCol.Line == -1 || (nextSemiCol.Line == nextColon.Line && nextSemiCol.Col > nextColon.Col) {
            fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
            fmt.Fprintln(os.Stderr, "\t" + nextColon.At())
            os.Exit(1)
        }
    }


    // parse case body -------------------
    if tokens.Peek().Type == token.SemiCol {
        fmt.Fprintln(os.Stderr, "[ERROR] missing case body(expr) for this case")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Last().At())
        os.Exit(1)
    }

    tokens.Next()
    caseExpr.Expr = prsExpr(tokens)

    if tokens.Peek().Type == token.SemiCol { tokens.Next() }

    return
}

func prsXSwitch(tokens *token.Tokens) *ast.XSwitch {
    switchExpr := ast.XSwitch{ Pos: tokens.Cur().Pos }
    var condBase ast.Expr = nil
    var placeholder *ast.Expr = nil

    // set condBase -----------------------
    if tokens.Next().Type != token.BraceL {
        condBase = prsExpr(tokens)
        placeholder = getPlaceholder(condBase)
    }

    // parse cases ------------------------
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" at the beginning of the xswitch " +
            "but got \"%s\"(%v)\n", tokens.Cur().Str, tokens.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    switchExpr.BraceLPos = tokens.Cur().Pos

    lastCaseEnd := token.Pos{}
    for tokens.Next().Type != token.BraceR {
        expr := prsCaseExpr(tokens, condBase, placeholder, lastCaseEnd)
        lastCaseEnd = expr.ColonPos
        switchExpr.Cases = append(switchExpr.Cases, expr)
    }

    switchExpr.BraceRPos = tokens.Cur().Pos


    // catch some syntax errors -----------
    if len(switchExpr.Cases) == 0 {
        fmt.Fprintln(os.Stderr, "[ERROR] empty xswitch")
        fmt.Fprintln(os.Stderr, "\t" + switchExpr.BraceLPos.At())
        os.Exit(1)
    }
    for i,c := range switchExpr.Cases {
        if c.Cond == nil && i != len(switchExpr.Cases)-1 {
            i = len(switchExpr.Cases)-1 - i
            if i == 1 {
                fmt.Fprintln(os.Stderr, "[ERROR] one case after the default case (unreachable code)")
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %d cases after the default case (unreachable code)\n", i)
            }
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }
    }
    if switchExpr.Cases[len(switchExpr.Cases)-1].Cond != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] every xswitch requires a default case")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    switchExpr.Type = switchExpr.Cases[0].GetType()
    return &switchExpr
}

func prsBinary(tokens *token.Tokens, expr ast.Expr, min_precedence precedence) ast.Expr {
    for isBinaryExpr(tokens) && getPrecedence(tokens) >= min_precedence {
        var b ast.Binary
        b.OperandL = expr

        precedenceL := getPrecedence(tokens)
        b.Operator = tokens.Next()

        tokens.Next()
        precedenceR := getPrecedence(tokens)

        // switch/xswitch
        if tokens.Cur().Type == token.BraceL {
            b.Type = types.BoolType{}
            return &b
        }

        switch {
        case isParenExpr(tokens):
            b.OperandR = prsParenExpr(tokens)
        case isUnaryExpr(tokens):
            b.OperandR = prsUnaryExpr(tokens)
        case tokens.Cur().Type == token.Name:
            switch tokens.Peek().Type {
            case token.Dot:
                ident := prsIdentExpr(tokens)
                b.OperandR = prsField(tokens, ident)
                for tokens.Peek().Type == token.Dot {
                    b.OperandR = prsField(tokens, b.OperandR)
                }

            case token.BrackL:
                expr := prsIdentExpr(tokens)
                tokens.Next()
                b.OperandR = prsIndexExpr(tokens, expr)

            case token.ParenL:
                b.OperandR = prsCallFn(tokens)

            default:
                b.OperandR = prsIdentExpr(tokens)
            }
        case tokens.Cur().Type == token.Str:
            b.OperandR = prsStrLit(tokens)
        default:
            b.OperandR = prsBasicLit(tokens)
        }

        b.Type = GetTypeBinary(&b)

        if isBinaryExpr(tokens) {
            b.OperandR = prsBinary(tokens, b.OperandR, precedenceL+1)
            b.Type = GetTypeBinary(&b)
        }

        // left to right as correct order of operations
        if precedenceR > precedenceL {
            swap(&b)
        }

        expr = &b
    }

    return expr
}

func swap(expr *ast.Binary) {
    switch expr.Operator.Type {
    case token.Minus:
        expr.Operator.Type = token.Plus
        expr.Operator.Str = "+"

        t := token.Token{ Type: token.Minus, Str: "-" }
        expr.OperandR = &ast.Unary{ Operator: t, Operand: expr.OperandR, Type: expr.Type }

    // TODO: proper fix
    // only tmp
    case token.Div:
        return

    case token.Geq:
        expr.Operator.Type = token.Leq
        expr.Operator.Str = "<="

    case token.Leq:
        expr.Operator.Type = token.Geq
        expr.Operator.Str = ">="

    case token.Grt:
        expr.Operator.Type = token.Lss
        expr.Operator.Str = "<"

    case token.Lss:
        expr.Operator.Type = token.Grt
        expr.Operator.Str = ">"
    }

    tmp := expr.OperandR
    expr.OperandR = expr.OperandL
    expr.OperandL = tmp
}

func prsGenericUsedType(tokens *token.Tokens) types.Type {
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

        tokens.Next()
        return typ
    }

    return nil
}

func prsCallGenericFn(tokens *token.Tokens) *ast.FnCall {
    ident := prsIdentExpr(tokens)

    tokens.Next()
    usedType := prsGenericUsedType(tokens)

    posL := tokens.Cur().Pos
    vals := prsPassArgs(tokens)
    posR := tokens.Cur().Pos

    if obj := identObj.Get(ident.Name); obj != nil {
        if f,ok := obj.(*identObj.Func); ok {
            f.AddTypeToGeneric(usedType)
            return &ast.FnCall{ Ident: *ident, F: f, GenericUsedType: usedType, Values: vals, ParenLPos: posL, ParenRPos: posR }

        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only call a function (%s is not a function)\n", ident.Name)
            fmt.Fprintln(os.Stderr, "\t" + ident.At())
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", ident.Name)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    return nil
}

func prsCallFn(tokens *token.Tokens) *ast.FnCall {
    ident := prsIdentExpr(tokens)
    posL := tokens.Next().Pos
    vals := prsPassArgs(tokens)
    posR := tokens.Cur().Pos

    if obj := identObj.Get(ident.Name); obj != nil {
        if f,ok := obj.(*identObj.Func); ok {
            return &ast.FnCall{ Ident: *ident, F: f, Values: vals, ParenLPos: posL, ParenRPos: posR }

        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only call a function (%s is not a function)\n", ident.Name)
            fmt.Fprintln(os.Stderr, "\t" + ident.At())
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", ident.Name)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    return nil
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
