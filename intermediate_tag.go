package main

import "strings"

type TagIntermediate struct {
	Name        string
	Description string
}

func intermediatateTags(commentBlock string) []TagIntermediate {

	tagIntermediates := make([]TagIntermediate, 0)
	sections := parseSections(commentBlock)

	for _, section := range sections {

		var tagIntermediate TagIntermediate

		title := strings.ToLower(section.Title)
		if title != "openapi tag" {
			continue
		}

		lines := strings.Split(section.Body, "\n")
		if len(lines) == 0 {
			// No body means no tag.
			continue
		}

		// The first line of the body is the actual tag.
		tagIntermediate.Name = strings.TrimSpace(lines[0])

		// The remaining lines are the tag description.
		if len(lines) > 1 {
			tagIntermediate.Description = strings.Join(lines[1:], "\n")
		}

		tagIntermediates = append(tagIntermediates, tagIntermediate)
	}

	return tagIntermediates

}
