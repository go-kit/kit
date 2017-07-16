package main

import "go/ast"

type (
	sourceContext struct {
		imports    []importPair
		interfaces []iface
	}

	importPair struct {
		alias *string
		path  string
	}

	iface struct {
		name    string
		methods []method
	}

	method struct {
		name    string
		params  []arg
		results []arg
	}

	arg struct {
		typ, name string
	}

	importSpecVisitor struct {
		src  *sourceContext
		node *ast.TypeSpec
	}

	typeSpecVisitor struct {
		src   *sourceContext
		node  *ast.TypeSpec
		iface *iface
	}

	interfaceTypeVisitor struct {
		src     *sourceContext
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []method
	}

	methodVisitor struct {
		src  *sourceContext
		node *ast.TypeSpec
		list *[]method
	}

	argVisitor struct {
		src             *sourceContext
		node            *ast.TypeSpec
		params, results *[]param
		isMethod        bool
	}
)

func (ip importPair) String() string {
	if ip.alias == nil {
		return ip.path
	}
	return fmt.Sprintf("%s %s", *ip.alias, ip.path)
}

func (sc *sourceContext) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return sc
	case *ast.ImportSpec:
		ip := importPair{path: rn.Path.Value}
		if rn.Name != nil {
			ip.alias = &(rn.Name.String())
		}
		sc.imports = append(sc.imports, ip)

	case *ast.TypeSpec:
		return &typeSpecVisitor{src: sc, node: rn}
	}
}

func (sc *sourceContext) validate() error {
	if len(sc.interfaces) != 1 {
		return fmt.Errorf("found %d interfaces, expecting exactly 1", sc.ifaceCount)
	}
	return nil
}

func (v *typeSpecVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Ident:
		if v.name == "" {
			v.name = rn.String()
		}
		return v
	case *ast.InterfaceType:
		return &interfaceTypeVisitor{src: v.src, ts: v, methods: []method{}}
	case nil:
		if v.iface != nil {
			v.iface.name = v.name
			v.interfaces = append(v.interfaces, *v.iface)
		}
		return nil
	}
}

func (v *interfaceTypeVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case ast.Field:
		return &methodVisitor{src: v.src, node: rn, list: &v.methods}
	case nil:
		v.ts.iface = &iface{methods: v.methods}
		return nil
	}
}

func (v *methodVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Ident:
		if rn.IsExported() {
			v.name = rn.String()
		}
		return v
	case *ast.FuncType:
		v.isMethod = true
		return v
	case *ast.Fieldlist:
		if v.params == nil {
			v.params = &[]Param{}
			return &argVisitor{src: v.src, list: v.params}
		}
		if v.results == nil {
			v.results = &[]param{}
		}
		return &argVisitor{src: v.src, list: v.results}
	case nil:
		if v.isMethod && v.name != "" {
			*v.list = append(*v.list, method{name: v.name, params: *v.params, results: *v.results})
		}
		return nil
	}
}

func (v *argVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.Ident:
		v.names = append(v.names, t.String())
		return nil
	case *ast.SelectorExpr:
		v.typ = t.String()
		return nil
	case nil:
		if v.typ == "" {
			v.typ, v.names = v.names[len(v.names)-1], v.names[:len(v.names)]
		}
		for _, n := range v.names {
			*v.list = append(*v.list, arg{typ: v.typ, name: n})
		}
		return nil
	}
}
