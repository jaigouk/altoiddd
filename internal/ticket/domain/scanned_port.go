package domain

import "maps"

// ScannedMethod represents a method signature extracted from a Go interface via AST scanning.
type ScannedMethod struct {
	name       string
	parameters map[string]string // param name -> param type
}

// NewScannedMethod creates a ScannedMethod value object.
func NewScannedMethod(name string, parameters map[string]string) ScannedMethod {
	params := make(map[string]string, len(parameters))
	maps.Copy(params, parameters)
	return ScannedMethod{name: name, parameters: params}
}

// Name returns the method name.
func (m ScannedMethod) Name() string { return m.name }

// Parameters returns a defensive copy of parameter name -> type mappings.
func (m ScannedMethod) Parameters() map[string]string {
	out := make(map[string]string, len(m.parameters))
	maps.Copy(out, m.parameters)
	return out
}

// ScannedPort represents an interface definition scanned from Go source code.
type ScannedPort struct {
	name     string
	filePath string
	methods  []ScannedMethod
}

// NewScannedPort creates a ScannedPort value object.
func NewScannedPort(name, filePath string, methods []ScannedMethod) ScannedPort {
	m := make([]ScannedMethod, len(methods))
	copy(m, methods)
	return ScannedPort{name: name, filePath: filePath, methods: m}
}

// Name returns the interface name.
func (p ScannedPort) Name() string { return p.name }

// FilePath returns the source file path.
func (p ScannedPort) FilePath() string { return p.filePath }

// Methods returns a defensive copy of the methods.
func (p ScannedPort) Methods() []ScannedMethod {
	out := make([]ScannedMethod, len(p.methods))
	copy(out, p.methods)
	return out
}
