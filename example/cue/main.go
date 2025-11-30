package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gilcrest/graupel"

	_ "embed"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
	"cuelang.org/go/encoding/yaml"

	"github.com/google/jsonschema-go/jsonschema"
)

// Embed our schema in a Go string variable.
//
//go:embed schema.cue
var cueSource string

func main() {

	ctx := cuecontext.New()

	// Build the schema
	schema := ctx.CompileString(cueSource).LookupPath(cue.ParsePath("#AgentSpec"))

	// Load the JSON file specified (the program's sole argument) as a CUE value
	dataFilename := os.Args[1]
	dataFile, err := os.ReadFile(dataFilename)
	if err != nil {
		log.Fatal(err)
	}

	var dataExpr ast.Expr
	dataExpr, err = json.Extract(dataFilename, dataFile)
	if err != nil {
		log.Fatal(err)
	}

	// Build the CUE value from the JSON data
	v := ctx.BuildExpr(dataExpr)

	// Validate the JSON data using the schema
	uv := schema.Unify(v)
	if err := uv.Validate(); err != nil {
		log.Fatal(err)
	}

	var y []byte
	y, err = yaml.Encode(uv)
	if err != nil {
		log.Fatal(err)
	}

	// Build output filename from input filename
	base := filepath.Base(dataFilename)
	ext := filepath.Ext(base)
	rootName := strings.TrimSuffix(base, ext)
	outfile := rootName + ".yaml"

	err = os.WriteFile(outfile, y, 0644)
	if err != nil {
		log.Fatal(err)
	}

	var js *jsonschema.Schema
	js, err = jsonschema.For[graupel.AgentSpec](nil)
	if err != nil {
		log.Fatal(err)
	}

	var mjs []byte
	mjs, err = js.MarshalJSON()

	schemaOutfile := rootName + "_schema.json"
	err = os.WriteFile(schemaOutfile, mjs, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wrote YAML file: %s\n", outfile)
	fmt.Printf("Wrote JSON Schema file: %s\n", schemaOutfile)
}
