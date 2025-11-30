package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gilcrest/graupel"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PAT              string `required:"true"`
	SnowflakeBaseURL string `required:"true"`
}

const applicationName = "graupel"

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

	// List all Cortex Agents
	var (
		agent *graupel.CortexAgent
		//resp  *http.Response
	)
	agent, _, err = client.Agents.Describe(ctx, "snowflake_intelligence", "agents", "CLAIMNOTEAGENT")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Cortex Agent: %+v\n", agent)
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
