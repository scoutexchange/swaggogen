package main

import (
	"fmt"
	"github.com/jackmanlabs/bucket/jlog"
	"github.com/jackmanlabs/errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func findEnumValues(referringPackage, typeName string) ([]interface{}, error) {

	pkgInfo := pkgInfos[referringPackage]
	typeName = strings.TrimPrefix(typeName, "*")

	importPaths := possibleImportPaths(pkgInfo, typeName)

	if len(importPaths) == 0 {
		log.Print("ERROR: Import paths not available for this enum type: ", typeName)
		jlog.Log(pkgInfo)
	}

	if len(importPaths) > 1 {
		log.Printf("WARNING: Multiple package candidates found for enum type (%s):", typeName)
		jlog.Log(importPaths)
	}

	if strings.Contains(typeName, ".") {
		index := strings.Index(typeName, ".") + 1
		typeName = typeName[index:]
	}

	for _, importPath := range importPaths {

		if importPath == "" {
			log.Print("Import path is blank!")
		}

		bpkg, err := build.Import(importPath, srcPath, 0)
		if err != nil {
			return nil, errors.Stack(err)
		}

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, bpkg.Dir, nil, parser.AllErrors)
		if err != nil {
			return nil, errors.Stack(err)
		}

		for _, pkg := range pkgs {
			enumVisitor := &EnumVisitor{
				Fset:     fset,
				TypeName: typeName,
				Values:   make([]interface{},0),
			}

			ast.Walk(enumVisitor, pkg)

			return enumVisitor.Values, nil
		}
	}

	return nil, nil
}

type EnumVisitor struct {
	Fset     *token.FileSet
	TypeName string

	// I would really like to store the names for the values, but the
	// JSON/OpenAPI spec only wants the values.
	Values []interface{}
}

func (this *EnumVisitor) Visit(node ast.Node) (w ast.Visitor) {

	if this.Fset == nil {
		log.Println("fset is nil.")
		return nil
	}

	switch t := node.(type) {

	case *ast.GenDecl:

		//ast.Fprint(os.Stderr, this.Fset, t, ast.NotNilFilter)

		// I've seen some folks that have many different kinds of constants in a
		// single const() declaration. Therefore, we need to assume that any of
		// the declarations could pertain to our type.

		var (
			//nam string = ""
			typ string = ""
			val interface{}
			iot int = 0
		)

		for _, spec := range t.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				//log.Printf("unexpected type %T\n", t) // %T prints whatever type t has
				continue
			}

			if valSpec.Type != nil {
				typ = resolveTypeExpression(valSpec.Type)
			}

			if typ != this.TypeName {
				continue
			}

			if len(valSpec.Names) != 1 {
				log.Print("WARNING: A possible constant declaration was found, but has more than one name: " + typ)
				continue
			} else {
				//nam = valSpec.Names[0].Name
			}

			if len(valSpec.Values) == 0 {
				// Assume iota is in play.
				iot++
				val = iot
			} else if len(valSpec.Values) > 1 {
				log.Print("WARNING: A possible constant declaration was found, but has more than one value: " + typ)
				continue
			} else {
				val = resolveValueExpression(valSpec.Values[0])
				if val == "iota" {
					// reset iota
					val = 0
					iot = 0
				}
			}

			this.Values = append(this.Values, val)

			//log.Printf("var %s\t%s = %v", nam, typ, val)
		}

		//case *ast.ValueSpec:
	//
	//	valueType := resolveTypeExpression(t.Type)
	//	if valueType != this.TypeName {
	//		return nil
	//	}
	//
	//	//ast.Fprint(os.Stderr, this.Fset, t, ast.NotNilFilter)
	//
	//	// Assume we have one name and one value.
	//	if len(t.Names) != 1 || len(t.Values) != 1 {
	//		log.Print("WARNING: A possible constant declaration was found, but has more than one name or value: " + valueType)
	//		return nil
	//	}
	//
	//	valueValue := resolveValueExpression(t.Values[0])
	//
	//	this.Values = append(this.Values, valueValue)

	case *ast.FuncDecl:
		// Ignore function declarations.
		return nil
	case *ast.ImportSpec:
		// Ignore import declarations.
		return nil
	case nil:
		//default:
		//	log.Printf("unexpected type %T\n", t) // %T prints whatever type t has
	}

	return this
}

func resolveValueExpression(expr ast.Expr) interface{} {
	switch t := expr.(type) {
	case *ast.BasicLit:
		return t.Value
	case *ast.UnaryExpr:
		return t.Op.String() + fmt.Sprint(resolveValueExpression(t.X))
	case *ast.Ident:
		return t.Name
	default:
		return fmt.Sprintf("Unknown<%T>", t)
	}
}
