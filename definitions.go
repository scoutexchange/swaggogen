package main

import (
	"github.com/jackmanlabs/errors"
	"strings"
)

func deriveDefinitionsFromOperations(operationIntermediates []OperationIntermediate) (DefinitionStore, error ){

	var definitionStore DefinitionStore        = make(map[string]*DefinitionIntermediate)


	for _, operationIntermediate := range operationIntermediates {
		for _, responseIntermediate := range operationIntermediate.Responses {
			definitionIntermediate,err := getDefinition(operationIntermediate.PackagePath, responseIntermediate.Type.GoType())
			err := responseIntermediate.Type.DefineDefinitions(operationIntermediate.PackagePath)
			if err != nil {
				return errors.Stack(err)
			}
		}
		for _, parameterIntermediate := range operationIntermediate.Parameters {
			err := parameterIntermediate.Type.DefineDefinitions(operationIntermediate.PackagePath)
			if err != nil {
				return errors.Stack(err)
			}
		}
	}

	return nil
}


func  getDefinition(referringPackage, goType string) (*DefinitionIntermediate,error ){

	if referringPackage == "" {
		return nil,errors.New("Referencing Package Path is empty.")
	}

	var err error


	if goType == "nil" {
		return nil,nil
	}

	if isPrimitive, _, _ := IsPrimitive(goType); isPrimitive {
		return nil,nil
	}



	var definition *DefinitionIntermediate
	definition, ok := definitionStore.ExistsDefinition(referringPackage, goType)

	if !ok {
		definition, err = findDefinition(referringPackage, goType)
		if err != nil {
			return errors.Stack(err)
		} else if definition == nil {
			return errors.New("Failed to generate definition for type: " + goType)
		}

		definitionStore.Add(definition)
	}

	this.PackagePath = definition.PackagePath
	this.PackageName = definition.PackageName

	if !ok {
		// This triggers the definition of all the members of the discovered type associated with the present member.
		definition.DefineDefinitions()
	}

	return nil
}


// What packages could have possibly contained this type?
func possibleImportPaths(pkgInfo PackageInfo, goType string) []string {

	if !strings.Contains(goType, ".") {
		return []string{pkgInfo.ImportPath}
	}

	chunks := strings.Split(goType, ".")

	alias := chunks[0]

	importPaths := make([]string, 0)

	for importPath, aliases := range pkgInfo.Imports {
		for _, alias_ := range aliases {
			if alias_ == alias {
				// I'm pretty sure that there should never be duplicate importPaths here.
				// Otherwise, check for duplicates.
				importPaths = append(importPaths, importPath)
			}
		}
	}

	return importPaths
}
