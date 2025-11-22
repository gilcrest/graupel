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
