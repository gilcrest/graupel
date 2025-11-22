package graupel

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AgentObjectService handles communication with the Cortex Agents related
// methods of the Snowflake Cortex API.
type AgentObjectService service

// List returns a list of Cortex Agents in the specified database and schema.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#list-cortex-agents for more details.
func (s *AgentObjectService) List(ctx context.Context, database, schema string) ([]*CortexAgent, *http.Response, error) {
	if database == "" {
		return nil, nil, fmt.Errorf("database cannot be empty")
	}
	if schema == "" {
		return nil, nil, fmt.Errorf("schema cannot be empty")
	}

	u := fmt.Sprintf("databases/%s/schemas/%s/agents",
		url.PathEscape(database),
		url.PathEscape(schema))

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var (
		agents []*CortexAgent
		resp   *Response
	)
	resp, err = s.client.Do(ctx, req, &agents)
	if err != nil {
		return nil, nil, err
	}

	return agents, resp.Response, nil
}

// Describe retrieves a specific Cortex Agent by name.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#describe-cortex-agent for more details.
func (s *AgentObjectService) Describe(ctx context.Context, database, schema, name string) (*CortexAgent, *http.Response, error) {
	if database == "" {
		return nil, nil, fmt.Errorf("database cannot be empty")
	}
	if schema == "" {
		return nil, nil, fmt.Errorf("schema cannot be empty")
	}
	if name == "" {
		return nil, nil, fmt.Errorf("agent name cannot be empty")
	}

	u := fmt.Sprintf("databases/%s/schemas/%s/agents/%s",
		url.PathEscape(database),
		url.PathEscape(schema),
		url.PathEscape(name))

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// CortexAgent instance to hold the response
	agent := new(CortexAgent)

	var resp *Response
	resp, err = s.client.Do(ctx, req, agent)
	if err != nil {
		return nil, nil, err
	}

	return agent, resp.Response, nil
}

// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#create-cortex-agent for more details.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#update-cortex-agent for more details.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#delete-cortex-agent for more details.
