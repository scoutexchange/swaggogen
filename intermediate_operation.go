package main

import (
	"bufio"
	"bytes"
	"github.com/go-openapi/spec"
	"regexp"
	"strconv"
	"strings"
)

// This is an intermediate representation of a path and/or operation as parsed
// in the comments. A collection of these can be combined and transformed to
// create the swagger hierarchy.
type OperationIntermediate struct {
	Title       string
	Description string
	Accepts     []string
	Parameters  []ParameterIntermediate
	Responses   []*ResponseIntermediate
	Path        string
	Method      string
	PackagePath string // Where this operation was found.
	Tag         string
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
	schema.Title = ""

	return schema
}

// This function does not do type detection. It merely scrapes what information
// there is in the comment block.
func intermediatateOperation(commentBlock string) OperationIntermediate {

	/*
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

		OpenAPI Response Body:
			[]types.Village

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
	*/

	// @Title Get TimeZone
	// @Description Return a TimeZone, given its id
	// @Accept  json
	//
	// @Param   x-ems-consumer	header	string  true	"Defines the consumer of the API. MobileApp, etc."
	// @Param   x-ems-api-token	header	string	true	"Auth token, from /authenticate request"
	// @Param   id				path	int     true	"TimeZone ID"
	// @Param   timestamp      	query   string	true    "dateTime in timeZone local time, for which to get timezone info with offsets adjusted for DST, RFC3339"
	//
	// @Success 200 {object} model.TimeZoneModel "Success"
	// @Failure 400 {object} apicommon.ErrorResponse "Bad Request"
	// @Failure 401 {object} apicommon.ErrorResponse "Invalid or missing consumer credentials"
	// @Router /timezones/{id} [get]

	var (
		// At the time of writing, IntelliJ erroneously warns on unnecessary
		// escape sequences. Do not trust IntelliJ.
		rxAccept      *regexp.Regexp = regexp.MustCompile(`@Accept\s+(.+)`)
		rxDescription *regexp.Regexp = regexp.MustCompile(`@Description\s+(.+)`)
		rxParameter   *regexp.Regexp = regexp.MustCompile(`@Param\s+([\w-]+)\s+(\w+)\s+([\w\.]+)\s+(\w+)\s+\"(.+)\"`)
		rxResponse    *regexp.Regexp = regexp.MustCompile(`@(Success|Failure)\s+(\d+)\s+\{([\w]+)\}\s+([\w\.]+)\s+\"(.+)\"`)
		rxRouter      *regexp.Regexp = regexp.MustCompile(`@Router\s+([/\w\d-{}]+)\s+\[(\w+)\]`)
		rxTitle       *regexp.Regexp = regexp.MustCompile(`@Title\s+(.+)`)
	)

	var operationIntermediate OperationIntermediate = OperationIntermediate{
		Accepts:    make([]string, 0),
		Parameters: make([]ParameterIntermediate, 0),
		Responses:  make([]*ResponseIntermediate, 0),
	}

	b := bytes.NewBufferString(commentBlock)
	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		line := scanner.Text()

		switch {

		case rxAccept.MatchString(line):

			raw := rxAccept.FindStringSubmatch(line)[1]
			accepts := strings.Split(raw, ",")
			for _, accept := range accepts {
				accept = strings.TrimSpace(accept)
				accept = strings.ToLower(accept)

				if accept == "" {
					continue
				} else if accept == "json" {
					accept = "application/json"
				} else if accept == "xml" {
					accept = "application/xml"
				}

				operationIntermediate.Accepts = append(operationIntermediate.Accepts, accept)
			}

		case rxDescription.MatchString(line):
			operationIntermediate.Description = rxDescription.FindStringSubmatch(line)[1]
		case rxParameter.MatchString(line):

			matches := rxParameter.FindStringSubmatch(line)

			parameterType := &MemberIntermediate{
				Type:     matches[3],
				JsonName: matches[1],
			}

			parameterIntermediate := ParameterIntermediate{
				In:          matches[2],
				Type:        parameterType,
				Required:    strings.ToLower(matches[4]) == "true",
				Description: matches[5],
			}

			operationIntermediate.Parameters = append(operationIntermediate.Parameters, parameterIntermediate)

		case rxResponse.MatchString(line):

			matches := rxResponse.FindStringSubmatch(line)
			statusCode, _ := strconv.Atoi(matches[2])

			goType := matches[4]
			goTypeMeta := matches[3]
			if strings.ToLower(goTypeMeta) == "array" && !strings.HasPrefix(goType, "[]") {
				goType = "[]" + goType
			}

			var responseType SchemerDefiner

			if isSlice, v := IsSlice(goType); isSlice {
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

			responseIntermediate := &ResponseIntermediate{
				Success:     strings.ToLower(matches[1]) == "success",
				StatusCode:  statusCode,
				Type:        responseType,
				Description: matches[5],
			}

			operationIntermediate.Responses = append(operationIntermediate.Responses, responseIntermediate)

		case rxRouter.MatchString(line):
			matches := rxRouter.FindStringSubmatch(line)
			operationIntermediate.Path = matches[1]
			operationIntermediate.Method = matches[2]

		case rxTitle.MatchString(line):
			operationIntermediate.Title = rxTitle.FindStringSubmatch(line)[1]

		default:

			//log.Print(line)

		}
	}

	return operationIntermediate
}
