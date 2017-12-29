package main

import (
	"strings"
)

type ApiIntermediate struct {
	ApiVersion     string
	ApiTitle       string
	ApiDescription string
	BasePath       string
}

func intermediatateApi(commentBlocks []string) ApiIntermediate {

	/*
		OpenAPI API Title:
			Agame Public API
		OpenAPI API Description:
			The Agame Public API is the API that is exposed to the world to facilitate
			gameplay.
		OpenAPI API Version:
			1.0
		OpenAPI Base Path:
			/api
	*/

	var apiIntermediate ApiIntermediate = ApiIntermediate{}

	for _, commentBlock := range commentBlocks {

		sections := parseSections(commentBlock)

		for _, section := range sections {

			//log.Print(section)

			title := strings.TrimSpace(section.Title)
			title = strings.ToLower(title)

			switch title {
			case "openapi api title":
				if l, ok := section.Line(0); ok {
					apiIntermediate.ApiTitle = l
				}
			case "openapi api description":
				apiIntermediate.ApiDescription = section.Body
			case "openapi api version":
				if l, ok := section.Line(0); ok {
					apiIntermediate.ApiVersion = l
				}
			case "openapi base path":
				if l, ok := section.Line(0); ok {
					apiIntermediate.BasePath = l
				}
			}
		}
	}

	return apiIntermediate
}

func getFirstWord(s string) string {
	words := strings.Split(s, " ")
	for _, word := range words {
		word = strings.Trim(word, "\t")
		if word != "" {
			return word
		}
	}

	return ""
}
