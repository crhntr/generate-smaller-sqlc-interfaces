package interfaces

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

func Generate(w io.Writer, pkg *packages.Package, sqlFilePaths ...string) error {
	var mappings []Mapping
	for _, filePath := range sqlFilePaths {
		buf, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		mappings, err = parseInterfaceMappings(mappings, buf)
		if err != nil {
			return err
		}
	}

	interfaces := make(map[string][]*ast.FuncDecl)

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) != 1 {
				continue
			}
			startIdentType, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			identType, ok := startIdentType.X.(*ast.Ident)
			if !ok || identType.Name != "Queries" {
				continue
			}
			i := slices.IndexFunc(mappings, func(mapping Mapping) bool {
				return mapping.Method == fd.Name.Name
			})
			if i < 0 {
				continue
			}
			mapping := mappings[i]

			for _, interfaceName := range mapping.Interfaces {
				methods := interfaces[interfaceName]

				methods = append(methods, fd)

				interfaces[interfaceName] = methods
			}
		}
	}

	interfaceNames := slices.Collect(maps.Keys(interfaces))
	slices.Sort(interfaceNames)

	out := bytes.NewBuffer(nil)

	for _, name := range interfaceNames {
		methods := interfaces[name]

		interfaceType := &ast.InterfaceType{
			Methods: new(ast.FieldList),
		}

		for _, method := range methods {
			interfaceType.Methods.List = append(interfaceType.Methods.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(method.Name.Name)},
				Type:  method.Type,
			})
		}

		out.WriteString("\n\n")
		if err := format.Node(out, token.NewFileSet(), &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(name),
					Type: interfaceType,
				},
			},
		}); err != nil {
			return err
		}
	}
	_, err := w.Write(out.Bytes())
	return err
}

type Mapping struct {
	Method     string
	Interfaces []string
}

func parseInterfaceMappings(mappings []Mapping, buf []byte) ([]Mapping, error) {
	s := bufio.NewReader(bytes.NewReader(buf))
	for i := 0; i < len(buf); i++ {
		line, err := s.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if suffix, ok := strings.CutPrefix(line, "-- name: "); ok {
			j := strings.LastIndex(suffix, ":")
			name := strings.TrimSpace(suffix[:min(j, len(suffix))])
			mappings = append(mappings, Mapping{
				Method: name,
			})
		}
		if suffix, ok := strings.CutPrefix(line, "-- interface: "); ok && len(mappings) > 0 {
			list := mappings[len(mappings)-1].Interfaces
			names := strings.Fields(suffix)
			list = append(list, names...)
			slices.Sort(list)
			list = slices.Compact(list)
			mappings[len(mappings)-1].Interfaces = list
		}
	}
	return mappings, nil
}
