package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
	"github.com/kelseyhightower/envconfig"

	_ "embed"

	"github.com/gilcrest/graupel"
)

type Config struct {
	PAT              string `required:"true"`
	SnowflakeBaseURL string `required:"true"`
}

const applicationName = "graupel"

// Embed our schema in a Go string variable.
//
//go:embed schema.cue
var cueSource string

func main() {
	var c Config
	err := envconfig.Process(applicationName, &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create a Snowflake Cortex client
	var client *graupel.Client
	client, err = graupel.NewClient(nil).
		WithProgrammaticAccessToken(c.PAT).
		WithSnowflakeBaseURL(c.SnowflakeBaseURL)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	var agent *graupel.CortexAgent
	agent, err = makeAgentViaJSON()
	if err != nil {
		log.Fatal(err)
	}

	// List all Cortex Agents
	var resp *graupel.Response
	resp, err = client.Agents.Create(ctx, agent)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response Status: %+v\n", resp.Response.Status)
}

func makeAgent() *graupel.CortexAgent {
	return &graupel.CortexAgent{
		Owner:         nil,
		DatabaseName:  nil,
		SchemaName:    nil,
		Name:          nil,
		Comment:       nil,
		Profile:       nil,
		Models:        nil,
		Instructions:  nil,
		Orchestration: nil,
		Tools:         nil,
		ToolResources: nil,
		AgentSpecStr:  nil,
		CreatedOn:     time.Time{},
	}
}

func makeAgentViaJSON() (*graupel.CortexAgent, error) {
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
	return new(graupel.CortexAgent), nil
}

func makeAgentViaYAML() *graupel.CortexAgent {
	return new(graupel.CortexAgent)
}

//type authTransport struct {
//	Transport http.RoundTripper
//	Token     string
//}
//
//func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
//	req = req.Clone(req.Context())
//	req.Header.Set("Authorization", "Bearer "+t.Token)
//	return t.Transport.RoundTrip(req)
//}
