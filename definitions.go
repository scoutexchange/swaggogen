package main

import (
	"github.com/jackmanlabs/errors"
	"strings"
)

func deriveDefinitionsFromOperations(operationIntermediates []OperationIntermediate) (DefinitionStore, error) {

	var defStore DefinitionStore = make(map[string]*DefinitionIntermediate)

	// This first loop gets all the top-level definitions.
	for _, operationIntermediate := range operationIntermediates {
		for _, responseIntermediate := range operationIntermediate.Responses {
			var typ SchemerDefiner = responseIntermediate.Type
			referringPackage := operationIntermediate.PackagePath
			goType := typ.GoType()

			defs, err := getDefinition(defStore, referringPackage, goType)
			if err != nil {
				return defStore, errors.Stack(err)
			}

			defStore.Add(defs...)
			if len(defs) > 0 {
				typ.SetPackageName(defs[0].PackageName)
				typ.SetPackagePath(defs[0].PackagePath)
			}
		}

		for _, parameterIntermediate := range operationIntermediate.Parameters {
			var typ SchemerDefiner = parameterIntermediate.Type
			referringPackage := operationIntermediate.PackagePath
			goType := typ.GoType()

			defs, err := getDefinition(defStore, referringPackage, goType)
			if err != nil {
				return defStore, errors.Stack(err)
			}

			if len(defs) > 0 {
				typ.SetPackageName(defs[0].PackageName)
				typ.SetPackagePath(defs[0].PackagePath)
				defStore.Add(defs...)
			}
		}
	}

	// This loop gets all the definitions of the sub-types of formerly defined
	// definitions.
	moreFound := true
	for moreFound {

		defs, err := findNextUnknownDefinition(defStore)
		if err != nil {
			return defStore, nil
		}

		if len(defs) > 0 {
			moreFound = true
			defStore.Add(defs...)
		}else{
			moreFound = false
		}
	}

	return defStore, nil
}

// This is used to allow incremental building of the definition store.
// Otherwise, we risk a lot of duplicate lookups.
func findNextUnknownDefinition(defStore DefinitionStore) ([]*DefinitionIntermediate, error) {


	for _, def := range defStore {

		goType := def.Name

		defs, err := getDefinition(defStore, def.PackagePath, goType)
		if err != nil {
			return nil, errors.Stack(err)
		}

		if len(defs) > 0 {
			return defs, nil
		}

		for _, member := range def.Members {
			defs, err := getDefinition(defStore, def.PackagePath, member.GoType())
			if err != nil {
				return nil, errors.Stack(err)
			}

			if len(defs) > 0 {
				return defs, nil
			}
		}
	}


	return nil, nil
}

// This is troublesome.
// We have the definition store available, but I want to maintain a fairly
// functional coding style. If we do the add here, we can do the add
// intentionally, i.e. only when the add is necessary.
// On the other hand, I doubt a duplicate addition would would cost much.
// Or even happen frequently.
func getDefinition(defStore DefinitionStore, referringPackage, goType string) ([]*DefinitionIntermediate, error) {

	if referringPackage == "" {
		return nil, errors.New("Referencing Package Path is empty.")
	}


	if goType == "nil" {
		return nil, nil
	}

	if isPrimitive, _, _ := IsPrimitive(goType); isPrimitive {
		return nil, nil
	}

	var types []string

	// There was a real temptation to make this recursive instead of a loop.
	if ok, t := IsSlice(goType); ok {
		types = []string{t}
	} else if ok, t, u := IsMap(goType); ok {
		types = []string{t, u}
	} else {
		types = []string{goType}
	}

	var defs []*DefinitionIntermediate = make([]*DefinitionIntermediate, 0)

	for _, typ := range types {
		def, ok := defStore.ExistsDefinition(referringPackage, typ)

		if ok {
			continue
		}

		_, err := findDefinition(referringPackage, goType)
		if err != nil {
			return defs, errors.Stack(err)
		} else if def == nil {
			return defs, errors.New("Failed to generate definition for type: " + goType)
		}

		// Embedded types require special treatment. we need the definitions
		// right now to construct the flattened struct. Also, we don't
		// necessarily want the embedded struct type to show up in the
		// definitions.
		// Suggestion for enhancement: get the embedded types first, possibly in
		// a separate store.
		for _, embeddedType := range def.EmbeddedTypes {
			def, ok := defStore.ExistsDefinition(def.PackagePath, embeddedType)
			if !ok {
				def, err = findDefinition(def.PackagePath, embeddedType)
				if err != nil {
					return nil, errors.Stack(err)
				} else if def == nil {
					return nil, errors.Newf("Failed to find definition for embedded member: %s:%s", goType, embeddedType)
				}
			}

			mergeDefinitions(def, def)
		}

		defs = append(defs, def)
	}

	return defs, nil
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
