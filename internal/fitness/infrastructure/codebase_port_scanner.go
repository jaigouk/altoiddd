package infrastructure

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// MethodSignature is a single method extracted from an interface.
type MethodSignature struct {
	name       string
	parameters map[string]string
}

// NewMethodSignature creates a MethodSignature.
func NewMethodSignature(name string, parameters map[string]string) MethodSignature {
	params := make(map[string]string, len(parameters))
	for k, v := range parameters {
		params[k] = v
	}
	return MethodSignature{name: name, parameters: params}
}

// Name returns the method name.
func (m MethodSignature) Name() string { return m.name }

// Parameters returns a defensive copy of parameter name -> type mappings.
func (m MethodSignature) Parameters() map[string]string {
	out := make(map[string]string, len(m.parameters))
	for k, v := range m.parameters {
		out[k] = v
	}
	return out
}

// PortDefinition is an interface found in a port file.
type PortDefinition struct {
	name     string
	filePath string
	methods  []MethodSignature
}

// NewPortDefinition creates a PortDefinition.
func NewPortDefinition(name, filePath string, methods []MethodSignature) PortDefinition {
	m := make([]MethodSignature, len(methods))
	copy(m, methods)
	return PortDefinition{name: name, filePath: filePath, methods: m}
}

// Name returns the interface name.
func (p PortDefinition) Name() string { return p.name }

// FilePath returns the source file path.
func (p PortDefinition) FilePath() string { return p.filePath }

// Methods returns a defensive copy of the methods.
func (p PortDefinition) Methods() []MethodSignature {
	out := make([]MethodSignature, len(p.methods))
	copy(out, p.methods)
	return out
}

// CodebasePortScanner scans Go source files for interface definitions via AST.
type CodebasePortScanner struct{}

// Scan scans a directory for interface definitions.
func (s CodebasePortScanner) Scan(portsDir string) map[string]PortDefinition {
	result := make(map[string]PortDefinition)

	info, err := os.Stat(portsDir)
	if err != nil || !info.IsDir() {
		return result
	}

	entries, err := os.ReadDir(portsDir)
	if err != nil {
		return result
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(portsDir, entry.Name())
		file, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			slog.Debug("Skipping malformed file", "path", filePath)
			continue
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				methods := extractMethods(fset, ifaceType)
				port := NewPortDefinition(typeSpec.Name.Name, filePath, methods)
				result[typeSpec.Name.Name] = port
			}
		}
	}

	return result
}

func extractMethods(fset *token.FileSet, iface *ast.InterfaceType) []MethodSignature {
	var methods []MethodSignature
	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue // embedded interface
		}
		name := method.Names[0].Name
		if strings.HasPrefix(name, "_") || !ast.IsExported(name) {
			continue
		}

		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		params := make(map[string]string)
		if funcType.Params != nil {
			for _, field := range funcType.Params.List {
				typeStr := typeToString(field.Type)
				for _, paramName := range field.Names {
					params[paramName.Name] = typeStr
				}
			}
		}

		methods = append(methods, NewMethodSignature(name, params))
	}
	return methods
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.InterfaceType:
		return "any"
	default:
		return ""
	}
}
