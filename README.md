# Swaggogen

Swaggogen is a tool for extracting Go (golang) type information from an
application and combining it with code comments to generate a Swagger/OpenAPI
2.0 specification document.

## Usage

Swaggogen takes one parameter, `pkg`. This parameter should be the package path
of the application you want to document.

Example:

```
swaggogen -pkg github.com/foo/bar
```

The application will generate the Swagger/OpenAPI document as JSON and print it
to stdout.

It is acknowledged that there are some unavoidable warnings that are printed to
stderr, and it's not pretty. The author(s) know this, and it is preferred that
end users be aware of the limitations as they exist. Because these warnings are
printed to stderr (not stdout), they should not affect the output of any JSON.
Furthermore, in practical settings, these warnings have never indicated a
failure to generate a complete Swagger specification document. Please feel free
to submit a merge request as appropriate.

### Optional Flags

#### `profile` *string*

This is for programmers only.

This flag accepts the path to a file name where profiling details will be
stored. Profiling details can be reviewed using something like the following
command:

```
go tool pprof swaggogen swaggogen.prof
```

#### `ignore` *string*

This flag accepts a comma-separated list of packages that you want to ignore.
This is useful if, for example, you import a package that has annotations that 
shouldn't be in your final spec.

#### `naming` *string*

This flag accepts one of **full**, **partial**, or **simple**.

When using the value of **full**, the whole Go package path is used to generate
the Swagger model name from the Go type. To remain compatible with the Swagger
spec, the slashes are replaced by periods.

When using **partial**, the Swagger model names are comprised of the Go package
name and the Go type name, much in the same way that they would be referenced in
Go code. 

When using **simple**, the Swagger model name is simply the name of the
corresponding Go type. No package information is used.

As you may imagine, the likelihood of name collisions increases with each step
in this spectrum. There are no warnings in the code to protect you from
collisions.

## Recognized Comment Blocks

Swaggogen picks up on three kinds of comment blocks, **Tag Definitions**,
**Operation Definitions**, and **API Definitions**. There's no reason you can't
mix these comment blocks, but your results may vary. The only blocks it makes
sense to mix are **Tag Definition** and **API Definition** blocks. The only
reason these two blocks are recognized separately is for convenience.

Each comment block is processed as a series of sections. Each section begins
with a recognized tag, and then a body. For compatibility with Godoc, it is
recommended that the body be indented compared to the section tag.

An example section:

```
/*
OpenAPI Query String Parameters:
    world  string  required  World UUID
    user   string  optional  User UUID
    x      int     optional  X-coordinate for blind query
    y      int     optional  Y-coordinate for blind query
    w      int     optional  Width of query area
    h      int     optional  Height of query area
*/
```

The section ends when another tag is recognized.

If you want to add comments to a comment block that is likely to be processed by
Swaggogen, add those comments to the top of the block. Section processing only
begins when a section is actually detected.

Because of the complexity of the Swagger specification and the variety of
information that can be absorbed by the specification, I regret to inform you
that each section is processed in its own way. Hopefully the layouts of each
section makes enough practical sense to be practical. If you can suggest a
better layout for either the comment blocks or the sections, please put in a
merge request or contact me. 

those that specifically
contain these tags: "OpenAPI Tag:", "OpenAPI Path:", and "OpenAPI API Title:".
These tags are not case sensitive. If you combine the comment blocks, Swaggogen
will pick up the various tags in the comment blocks just the same.

###  

Swaggogen observes two kinds of code blocks, **API definitions** and **Route
Definitions**. The lines that are parsed for use in the Swagger document must
contain a keyword, which is a marker beginning with an '@'. The format of each
line depends on the keyword.

These annotations are intended to be compatible with Yuriy Vasiyarov's project,
found at http://github.com/yvasiyarov/swagger.

For the sake of simplicity, a **Route Definition**  combines the necessary
information to generate Paths and Operations in Swagger terminology. For this
reason, throughout the documentation, a *Route* will be in reference to a
Swagger *Operation* in combination with its respective *Path*.

### API Definitions

**API Definitions** support the following tags:

 * `OpenAPI API Title:`
 * `OpenAPI API Description:`
 * `OpenAPI API Version:`
 * `OpenAPI Base Path:`

The `OpenAPI API Title:` is required by the Swagger specification, and is used
as a trigger for detecting **API Definition** comment blocks. So, make sure you use
that tag to make things work. Multiple **API definitions** are allowed, but they
will be combined without any guarantees of precedence.

#### `OpenAPI API Title:`

The `OpenAPI API Title:` tag defines the title of your API. The first line in
the section body is accepted as your title.

Example:

```
OpenAPI API Title:
    My REST API
```

#### `OpenAPI API Description:`

The `OpenAPI API Description:` tag defines the description of your API. Any text
in the body of the section is accepted as your description.

Example:

```
OpenAPI API Description:
    My API is awesome!
```

#### `OpenAPI API Version:`

The `OpenAPI API Version:` tag defines the API version of your application. The
body of the section is used for the version number. Any combination of
contiguous printable characters is acceptable.

Example:

```
OpenAPI API Version:
    1.0.0-test
```

#### `OpenAPI Base Path:`

The `OpenAPI Base Path:` tag defines the base path of your API. Per the Swagger
specification, this path is prepended to all paths defined in the paths of your
specification. An acceptable path (URL component) should begin with a forward
slash and contain letters, numbers, periods, slashes, hyphens, and underscores.

Example:

```
OpenAPI Base Path:
    /api/v1
```

### Route Definitions

**Route Definitions** are comprised of lines beginning with the following
keywords:

* `OpenAPI Content Type:`
* `OpenAPI Description:`
* `OpenAPI Method:`
* `OpenAPI Path:`
* `OpenAPI Query String Parameters:`
* `OpenAPI Request Body:`
* `OpenAPI Responses:`
* `OpenAPI Summary:`
* `OpenAPI Tags:`

Any comment block containing the `@Router` tag is considered an **Route
Definition**. Multiple API route definitions are allowed.

In various **Route Definition** sections, a type must be specified. A type can
be a Swagger-defined primitive type (int, string, boolean, etc.) or a Go type.
A specification of `nil` is understood to mean 'no type'.

If the type specification should be a Go type, it must be specified exactly how
it would be referenced in code. For example, if the struct type `Foo` is in the
local package, then the argument can be referenced simply with `Foo`. If it is
defined in another package that is imported with an alias
(`import f "/github.com/jackmanlabs/fooness"`), then the type argument should be
referenced with the alias, `f.Foo`. 


A complete example:

```go
/*
OpenAPI Path:
	/api/foos

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
	[]types.Foo

OpenAPI Description:
	This endpoint returns all of the foos that belong to the user and world
	specified by the query string parameter.

	The `world` parameter is required in all uses.

	The `user` parameter returns all foos owned by that user. If the calling
	player has permission to view all foo information, then that information
	will be returned. Otherwise, only a subset of foo information is
	returned. Use of this parameter is the recommended way to get the calling
	user's foos. Use of this parameter takes precedence over the use of the
	`x`, `y`, `w`, and `h` parameters.

	Similar to the tiles endpoint (`/api/tiles [GET]`), the `x`, `y`, `w`, and
	`h` parameters control the retrieval of all the foos in a specific area
	of the map. Unless specified, the values for these parameters are assumed to
	be zero. If `w` and `h` are zero, then only one foo is returned (if it
	exists at the coordinates provided). The maximum values accepted for `w` and
	`h` will be 1000, and values exceeding 1000 will be quietly accepted as
	1000.

	In all circumstances, a set (array) is returned regardless of the quantity
	of foos returned.
*/
func HandleFoosGet(w http.ResponseWriter, r *http.Request, userId string) error {}

```

#### `OpenAPI Content Type:`

The `OpenAPI Content Type:` tag defines the set of MIME types that this Route
consumes and produces; symmetry in this regard is assumed.

The body of the section should contain a list of content types, one per line.

Admittedly, this tool takes some liberties and simply checks for the presence
of `json` or `xml`. Accordingly, the **Produces** and **Consumes** properties
of the corresponding Swagger Operation object is populated with standard
`application/json` and `application/xml` strings. This behavior will likely
change as greater sophistication is required. Feel free to submit a merge
request with more sophisticated behavior.

Example:

```
OpenAPI Content Type:
    json
    xml
```

#### `OpenAPI Description:`

The `OpenAPI Description:` tag defines a human readable description for the
Swagger Operation. Everything in the section body is assumed to be the
description.

Example:

```
OpenAPI Description:
    This route is a good one.
```

#### `OpenAPI Method:`

The `OpenAPI Method:` tag specifies the HTTP method of this endpoint. For most
people, this will be one of `GET`, `PUT`, `POST`, or `DELETE`. As of right now,
the method is not validated, and can be pretty much any single word. Also, the
method is not case sensitive, but will be made all-caps in the output for the
sake of common convention.

Example:

```
OpenAPI Method:
    GET
```

#### `OpenAPI Path:`

The `OpenAPI Path:` tag defines the path of the operation. Swaggogen expects
this to be a single line with the path of the operation, not including the base
path.

#### `OpenAPI Query String Parameters:`

The `OpenAPI Query String Parameters:` tag allows you to specify the query
string parameters of the operation.

Each line of the body may contain a single
parameter. Each line must contain four fields, in order: name, type, necessity,
and description. Each of the first three fields must not contain any spaces. Any
text after the first three fields is considered to be the description.
Quotations (with double quote marks, `"`) around the description are ignored.

The necessity of the parameter (required vs. optional) is specified in the third
field. A required field must have `required` in this field. Virtually any other
word is understood to mean 'not required', but I encourage `optional` because it
makes for some pretty columnar formatting.

Example:

```
OpenAPI Query String Parameters:
    world  string  required  World UUID
    user   string  optional  User UUID
    x      int     optional  X-coordinate for blind query
    y      int     optional  Y-coordinate for blind query
    w      int     optional  Width of query area
    h      int     optional  Height of query area
```

#### `OpenAPI Request Body:`

This tag specifies the type of the request body. This tag is optional, but if
you want to explicitly specify 'no body', use `nil`. No other information is
required.

Example:

```
OpenAPI Request Body:
    []foo.Bar
```

#### `OpenAPI Responses:`

The `OpenAPI Responses:` tag defines any number of responses that may be
generated by the endpoint.

One response should be given per line of the section body. Each line must
contain three fields: HTTP status code, body type, and a description.

The status code must be numeric. The body type must conform to a type as
described in the section titled *Route Definitions*. Neither of the first two
fields may contain spaces, and the description is any text following the first
two fields.

```
OpenAPI Responses:
    200 []foo.Bar   Normal response, a collection of foo.Bar objects.
    401 errs.Error  The user is not authenticated.
```

#### `OpenAPI Summary:`

The `OpenAPI Summary:` tag defines the summary of the operation. In previous
implementations of Swaggogen, this was called the title. To conform to the
Swagger/OpenAPI specification, it is now the summary.

The first line of the section body will be used as the summary.

Example:

```
OpenAPI Summary:
    Search for things.
```

#### `OpenAPI Tags:`

The `OpenAPI Tags:` tag defines the list of tags (as described in the section
*Tag Definitions*) that should apply to the operation. Correlation between tags
in the operations and tags defined by **Tag Definitions** is not validated.

Tags should be listed in the section body, one per line.

Example:

```
OpenAPI Tags:
    Villages
    Towns
```

# Code Structure

This tool operates, at least conceptually, in three phases: detection,
extraction, and generation.

During the parsing phase, the code project is scanned for comments blocks.

In the extraction phase, the comment blocks are transformed into intermediate
representations, called Intermediates. Some of these Intermediates are very
basic, such as the ApiIntermediate. It is comprised of parsed information and
nothing else. Other Intermediates, such as the DefinitionIntermediate, require
further analysis of the code, pulling struct definitions, enums, and so on from
the code to create Swagger type definitions in the generation stage.

The DefinitionIntermediates are special because they're store in a global
dictionary, called DefinitionStore, and can be referenced by any of the
extraction processes. This is primarily to improve performance and make certain
sections of code simpler.

Finally, in the generation phase, the Intermediates have been fully populated,
and the swagger object tree is generated. Many Intermediates have a Schema()
method so that they can generate their own schema by way of interface.

# Creditation

This tool was written blind with respect to other similar tools.

While this tool is intended to utilize the same annotations as the
[yvasiyarov project](http://github.com/yvasiyarov/swagger), the original parsing
algorithms were not copied (or even used as reference). Therefore, exact parsing
behavior is not expected to be the same.

Also, the availability of the
[OpenAPI specification models](http://github.com/go-openapi/spec) from the
[OpenAPI Initiative golang toolkit](https://github.com/go-openapi) library is
greatly appreciated.