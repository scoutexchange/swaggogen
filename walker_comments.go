package main

import (
	"github.com/jackmanlabs/errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func getCommentBlocks(pkgPath string) ([]string, error) {

	bpkg, err := build.Import(pkgPath, srcPath, 0)
	if err != nil {
		return []string{}, nil
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, bpkg.Dir, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, errors.Stack(err)
	}

	commentVisitor := &CommentVisitor{Fset: fset}
	for _, pkg := range pkgs {
		ast.Walk(commentVisitor, pkg)
	}

	return commentVisitor.Comments, nil
}

/*
We're using a map for the imports so we don't have to worry about duplicates.
*/
type CommentVisitor struct {
	Fset     *token.FileSet
	Comments []string
}

func (this *CommentVisitor) Visit(node ast.Node) (w ast.Visitor) {

	if this.Fset == nil {
		log.Println("fset is nil.")
		return nil
	}

	switch t := node.(type) {

	case *ast.CommentGroup:
		if this.Comments == nil {
			this.Comments = make([]string, 0)
		}

		this.Comments = append(this.Comments, t.Text())

		return nil

		//case nil:
		//default:
		//	fmt.Printf("unexpected type %T\n", t) // %T prints whatever type t has
	}

	return this
}

// This is used to detect blocks with 'OpenAPI Path:'. A comment block that describes a path/operation is useless if it
// fails to describe the path. Therefore, this is a good indicator.
func hasPathComments(commentBlocks []string) []string {
	return detectComments(commentBlocks, "OpenAPI Path:")
}

// This detects comments blocks with 'OpenAPI API Title:'. The API Title is a required member of the Swagger definition,
// so it must be present.
func detectApiCommentBlocks(commentBlocks []string) []string {
	return detectComments(commentBlocks, "OpenAPI API Title:")
}

// This detects comment blocks with 'OpenAPI Tag:'. There is no garantee that these tags declarations will be a part of
// any other comment block.
func detectTagComments(commentBlocks []string) []string {
	return detectComments(commentBlocks, "OpenAPI Tag:")
}


// Comment detection is case-insensitive.
// Any comment blocks that prove to have the test string will be returned.
func detectComments(commentBlocks []string, keyword string) []string {

	keyword = strings.ToLower(keyword)
	detectedBlocks := make([]string, 0)

	for _, comment := range commentBlocks {
		comment_ := strings.ToLower(comment)
		if strings.Contains(comment_, keyword) {
			detectedBlocks = append(detectedBlocks, comment)
		}
	}

	return detectedBlocks
}
