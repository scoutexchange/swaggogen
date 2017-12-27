package main

import (
	"regexp"
	"bytes"
	"bufio"
)


type ApiIntermediate struct {
	ApiVersion     string
	ApiTitle       string
	ApiDescription string
	BasePath       string
	SubApis        []SubApiIntermediate
}


func intermediatateApi(commentBlocks []string) ApiIntermediate {

	// @APIVersion 1.0.0
	// @APITitle REST API
	// @APIDescription EMS Rest API
	// @BasePath /api/v1
	// @SubApi HealthCheck [/health]

	var (
		// At the time of writing, IntelliJ erroneously warns on unnecessary
		// escape sequences. Do not trust IntelliJ.
		rxApiVersion     *regexp.Regexp = regexp.MustCompile(`@APIVersion\s+([\d\.]+)`)
		rxApiTitle       *regexp.Regexp = regexp.MustCompile(`@APITitle\s+(.+)`)
		rxApiDescription *regexp.Regexp = regexp.MustCompile(`@APIDescription\s+(.+)`)
		rxBasePath       *regexp.Regexp = regexp.MustCompile(`@BasePath\s+([/a-zA-Z0-9-]+)`)
		rxSubApi         *regexp.Regexp = regexp.MustCompile(`@SubApi\s+([0-9a-zA-Z]+)\s+\[([/a-zA-Z0-9-]+)\]`)
	)

	var apiIntermediate ApiIntermediate = ApiIntermediate{
		SubApis: make([]SubApiIntermediate, 0),
	}

	for _, commentBlock := range commentBlocks {

		b := bytes.NewBufferString(commentBlock)
		scanner := bufio.NewScanner(b)
		for scanner.Scan() {
			line := scanner.Text()

			switch {

			case rxApiDescription.MatchString(line):
				apiIntermediate.ApiDescription = rxApiDescription.FindStringSubmatch(line)[1]
			case rxApiTitle.MatchString(line):
				apiIntermediate.ApiTitle = rxApiTitle.FindStringSubmatch(line)[1]
			case rxApiVersion.MatchString(line):
				apiIntermediate.ApiVersion = rxApiVersion.FindStringSubmatch(line)[1]
			case rxBasePath.MatchString(line):
				apiIntermediate.BasePath = rxBasePath.FindStringSubmatch(line)[1]

			case rxSubApi.MatchString(line):
				matches := rxSubApi.FindStringSubmatch(line)
				subApi := SubApiIntermediate{
					Name: matches[1],
					Path: matches[2],
				}
				apiIntermediate.SubApis = append(apiIntermediate.SubApis, subApi)
			}
		}
	}

	return apiIntermediate
}
