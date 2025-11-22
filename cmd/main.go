package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
		agents []*graupel.CortexAgent
		resp   *http.Response
	)
	agents, resp, err = client.Agents.List(ctx, "snowflake_intelligence", "agents")
	if err != nil {
		log.Fatal(err)
	}

	for _, agent := range agents {
		fmt.Printf("Agent: %+v\n", agent)
		fmt.Println(agent)
	}

	fmt.Printf("Agent Count: %d\n", len(agents))
	fmt.Printf("HTTP Response: %+v\n", resp)

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
