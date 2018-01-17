package main

import (
	"github.com/go-openapi/spec"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// This is an intermediate representation of a path and/or operation as parsed
// in the comments. A collection of these can be combined and transformed to
// create the swagger hierarchy.
type OperationIntermediate struct {
	Accepts     []string
	Description string
	Method      string
	PackagePath string // Where this operation was found.
	Parameters  []ParameterIntermediate
	Path        string
	Responses   []*ResponseIntermediate
	Summary     string
	Tags        []string
}

type ParameterIntermediate struct {
	In          string
	Required    bool
	Description string
	Type        *MemberIntermediate
}

func (this *ParameterIntermediate) Schema() *spec.Schema {
	return this.Type.Schema()
}

type ResponseIntermediate struct {
	Success     bool
	StatusCode  int
	Description string
	Type        SchemerDefiner
}

func (this *ResponseIntermediate) Schema() *spec.Schema {

	schema := this.Type.Schema()

	if schema != nil {
		schema.Title = ""
	}

	return schema
}

// This function does not do type detection. It merely scrapes what information
// there is in the comment block.
func intermediatateOperation(commentBlock string) OperationIntermediate {

	/*
		OpenAPI Summary:
			List Villages

		OpenAPI Path:
			/api/villages

		OpenAPI Method:
			GET

		OpenAPI Query String Parameters:
			world  string  required  World UUID
			user   string  optional  User UUID
			x      int     optional  X-coordinate for blind query
			y      int     optional  Y-coordinate for blind query
			w      int     optional  Width of query area
			h      int     optional  Height of query area

		OpenAPI Request Body:
			nil

		OpenAPI Responses:
			200	[]types.Village	List of villages

		OpenAPI Description:
			This endpoint returns all of the villages that belong to the user and world
			specified by the query string parameter.

			The `world` parameter is required in all uses.

			The `user` parameter returns all villages owned by that user. If the calling
			player has permission to view all village information, then that information
			will be returned. Otherwise, only a subset of village information is
			returned. Use of this parameter is the recommended way to get the calling
			user's villages. Use of this parameter takes precedence over the use of the
			`x`, `y`, `w`, and `h` parameters.

			Similar to the tiles endpoint (`/api/tiles [GET]`), the `x`, `y`, `w`, and
			`h` parameters control the retrieval of all the villages in a specific area
			of the map. Unless specified, the values for these parameters are assumed to
			be zero. If `w` and `h` are zero, then only one village is returned (if it
			exists at the coordinates provided). The maximum values accepted for `w` and
			`h` will be 1000, and values exceeding 1000 will be quietly accepted as
			1000.

			In all circumstances, a set (array) is returned regardless of the quantity
			of villages returned.

		OpenAPI Tags:
			Villages

		OpenAPI Content Type:
			application/json
	*/

	var oi OperationIntermediate = OperationIntermediate{
		Accepts:    make([]string, 0),
		Parameters: make([]ParameterIntermediate, 0),
		Responses:  make([]*ResponseIntermediate, 0),
		Tags:       make([]string, 0),
	}

	sections := parseSections(commentBlock)

	//log.Print("\n",commentBlock)

	for _, section := range sections {
		title := strings.TrimSpace(section.Title)
		title = strings.ToLower(title)

		switch title {
		case "openapi summary":
			if l, ok := section.Line(0); ok {
				oi.Summary = l
			}
		case "openapi path":
			if l, ok := section.Line(0); ok {
				oi.Path = l
			}
		case "openapi method":
			if l, ok := section.Line(0); ok {
				oi.Method = l
			}
		case "openapi query string parameters":
			oi.Parameters = append(oi.Parameters, parseQueryStringParams(section)...)
		case "openapi path parameters":
			oi.Parameters = append(oi.Parameters, parsePathParams(section)...)
		case "openapi request body":
			l, ok := section.Line(0)
			if !ok {
				continue
			}

			bodyType := getFirstWord(l)
			if bodyType == "nil" {
				continue
			}

			bodyParam := ParameterIntermediate{
				In:          "body",
				Required:    true,
				Description: "",
				Type:        &MemberIntermediate{Type: bodyType},
			}

			oi.Parameters = append(oi.Parameters, bodyParam)

		case "openapi responses":
			oi.Responses = parseResponses(section)
		case "openapi description":
			oi.Description = section.Body
		case "openapi tags":
			for _, l := range section.Lines() {
				l = getFirstWord(l)

				if l == "" {
					continue
				} else {
					oi.Tags = append(oi.Tags, l)
				}
			}
		case "openapi content type":
			for _, l := range section.Lines() {
				l = getFirstWord(l)
				l = strings.ToLower(l)

				if l == "" {
					continue
				} else if strings.Contains(l, "json") {
					l = "application/json"
				} else if strings.Contains(l, "xml") {
					l = "application/xml"
				}
			}
		default:
			log.Print("Unrecognized section:\n", section)
		}
	}

	return oi
}

func parsePathParams(section Section) []ParameterIntermediate {
	var params []ParameterIntermediate = parseParams(section)

	for i := range params {
		params[i].In = "path"
	}

	return params

}

func parseQueryStringParams(section Section) []ParameterIntermediate {
	var params []ParameterIntermediate = parseParams(section)

	for i := range params {
		params[i].In = "query"
	}

	return params
}

func parseParams(section Section) []ParameterIntermediate {

	/*
		OpenAPI Query String Parameters:
		world  string  required  World UUID
		user   string  optional  User UUID
		x      int     optional  X-coordinate for blind query
		y      int     optional  Y-coordinate for blind query
		w      int     optional  Width of query area
		h      int     optional  Height of query area
	*/

	var (
		out []ParameterIntermediate = make([]ParameterIntermediate, 0)
		rx  *regexp.Regexp          = regexp.MustCompile(`(\S+)\s+(\w+)\s+(\w+)\s+(.+)`)
	)

	// This is probably the ugliest loop I have ever written in my life.
	for _, l := range section.Lines() {

		matches := rx.FindStringSubmatch(l)
		if matches == nil {
			// no match
			continue
		}

		paramType := &MemberIntermediate{
			Type:     matches[2],
			JsonName: matches[1],
		}

		desc := matches[4]
		desc = strings.TrimSpace(desc)
		if strings.HasPrefix(desc, "\"") {
			strings.Trim("desc", "\"")
		}

		var parameterIntermediate ParameterIntermediate = ParameterIntermediate{
			Description: matches[4],
			In:          "", // This should get set by the caller.
			Required:    strings.ToLower(matches[3]) == "required",
			Type:        paramType,
		}

		out = append(out, parameterIntermediate)
	}

	return out
}

func parseResponses(section Section) []*ResponseIntermediate {

	/*
		OpenAPI Responses:
			200	[]types.Village	List of villages
	*/

	var (
		out []*ResponseIntermediate = make([]*ResponseIntermediate, 0)
		rx  *regexp.Regexp          = regexp.MustCompile(`(\d+)\s+(\S+)\s+(.+)`)
	)

	// This is probably the ugliest loop I have ever written in my life.
	for _, l := range section.Lines() {

		matches := rx.FindStringSubmatch(l)
		if matches == nil {
			// no match
			log.Print("No match for response:", l)
			continue
		}

		goType := matches[2]

		var responseType SchemerDefiner

		if isMap, k, v := IsMap(goType); isMap {

			keyType := &MemberIntermediate{
				Type:        k,
				Validations: make(ValidationMap),
			}

			valueType := &MemberIntermediate{
				Type:        v,
				Validations: make(ValidationMap),
			}

			responseType = &MapIntermediate{
				Type:        goType,
				ValueType:   valueType,
				KeyType:     keyType,
				Validations: make(ValidationMap),
			}

		} else if isSlice, v := IsSlice(goType); isSlice {
			valueType := &MemberIntermediate{
				Type:        v,
				Validations: make(ValidationMap),
			}

			responseType = &SliceIntermediate{
				Type:        goType,
				ValueType:   valueType,
				Validations: make(ValidationMap),
			}
		} else {
			responseType = &MemberIntermediate{
				Type:        goType,
				Validations: make(ValidationMap),
			}
		}

		statusCode, _ := strconv.Atoi(matches[1])

		ri := &ResponseIntermediate{
			Success:     strings.ToLower(matches[1]) == "success",
			StatusCode:  statusCode,
			Type:        responseType,
			Description: matches[3],
		}

		out = append(out, ri)
	}

	return out
}
