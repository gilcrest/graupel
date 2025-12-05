package graupel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// AgentObjectService handles communication with the Cortex Agents related
// methods of the Snowflake Cortex API.
type AgentObjectService service

type ListOptions struct {
	// like filters the output by resource name. Uses case-insensitive pattern
	// matching with support for SQL wildcard characters.
	like string `url:"like,omitempty"`

	// fromName enables fetching rows only following the first row whose object
	// name matches the specified string. Case-sensitive and does not have to be
	// the full name.
	fromName string `url:"fromName,omitempty"`

	// showLimit limits the maximum number of rows returned by the command.
	// Minimum: 1. Maximum: 10000.
	showLimit int `url:"showLimit,omitempty"`
}

// List returns a list of Cortex Agents in the specified database and schema.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#list-cortex-agents for more details.
func (s *AgentObjectService) List(ctx context.Context, database, schema string, opts *ListOptions) ([]*CortexAgent, *Response, error) {
	if database == "" {
		return nil, nil, fmt.Errorf("database cannot be empty")
	}
	if schema == "" {
		return nil, nil, fmt.Errorf("schema cannot be empty")
	}

	u := fmt.Sprintf("databases/%s/schemas/%s/agents",
		url.PathEscape(database),
		url.PathEscape(schema))

	var err error
	u, err = addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	var req *http.Request
	req, err = s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var (
		agents []*CortexAgent
		resp   *Response
	)
	resp, err = s.client.Do(ctx, req, &agents)
	if err != nil {
		return nil, resp, err
	}

	return agents, resp, nil
}

type DescribeAgentResponse struct {
	AgentSpec    string    `json:"agent_spec"`
	Name         string    `json:"name"`
	DatabaseName string    `json:"database_name"`
	SchemaName   string    `json:"schema_name"`
	Owner        string    `json:"owner"`
	CreatedOn    time.Time `json:"created_on"`
}

// Describe retrieves a specific Cortex Agent by name.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#describe-cortex-agent for more details.
func (s *AgentObjectService) Describe(ctx context.Context, database, schema, name string) (*CortexAgent, *Response, error) {
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

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var (
		resp *Response
		dar  *DescribeAgentResponse
	)
	resp, err = s.client.Do(ctx, req, &dar)
	if err != nil {
		return nil, resp, err
	}

	// The only CortexAgent fields that are returned as part of the Describe
	// response are:
	//  "agent_spec", "name", "database_name", "schema_name", "owner", "created_on"
	//
	// Populate the remaining fields by Unmarshalling the AgentSpec JSON string
	// into the AgentSpec struct
	var agentSpec AgentSpec
	err = json.Unmarshal([]byte(dar.AgentSpec), &agentSpec)
	if err != nil {
		log.Fatalf("failed to unmarshal agent_spec: %v", err)
	}

	agent := new(CortexAgent)

	// not returned by the Describe API call?
	agent.Comment = nil
	agent.Profile = nil

	agent.Models = agentSpec.Models
	agent.Instructions = agentSpec.Instructions
	agent.Orchestration = agentSpec.Orchestration
	agent.Tools = agentSpec.Tools

	agent.ToolResources, err = ParseToolResources(agentSpec.ToolResources, agent.Tools)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse tool resources: %v", err)
	}

	// get missing fields (Comment and Profile) from List response
	var la []*CortexAgent
	la, _, err = s.List(ctx, database, schema, &ListOptions{like: name})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get agent list for additional fields: %v", err)
	}

	for _, a := range la {
		if *a.Name != name {
			continue
		}
		agent.Comment = a.Comment
		agent.Profile = a.Profile
		break
	}

	return agent, resp, nil
}

const (
	CreateModeErrorIfExists = "errorIfExists"
	CreateModeOrReplace     = "orReplace"
	CreateModeIfNotExists   = "ifNotExists"
)

// ValidateCreateMode checks if the given create mode is either empty or one of the
// valid CreateMode constants. Returns true if valid, false otherwise.
func ValidateCreateMode(mode string) bool {
	if mode == "" {
		return true
	}
	return mode == CreateModeErrorIfExists ||
		mode == CreateModeOrReplace ||
		mode == CreateModeIfNotExists
}

// CreateOptions is used to specify the Resource Creation Mode when creating a Cortex Agent.
// Valid values: "errorIfExists", "orReplace", "ifNotExists"
type CreateOptions struct {
	CreateMode string `url:"createMode,omitempty"`
}

type CreateRequest struct {
	// The database name where the agent is located.
	DatabaseName *string `json:"database_name,omitempty"`

	// The schema name where the agent is located.
	SchemaName *string `json:"schema_name,omitempty"`

	// The name of the agent.
	Name *string `json:"name"`

	// The comment or description for the agent.
	Comment *string `json:"comment,omitempty"`

	// Profile information for the agent.
	Profile *AgentProfile `json:"profile,omitempty"`

	// Model to use for orchestration. If not provided, a model is automatically selected.
	Models *ModelConfig `json:"models,omitempty"`

	// AgentInstructions contains the various instruction types for the agent.
	Instructions *AgentInstructions `json:"instructions,omitempty"`

	// Orchestration configuration for the agent - currently only budget is supported.
	Orchestration *OrchestrationConfig `json:"orchestration,omitempty"`

	// List of tools available to the agent.
	Tools []*Tool `json:"tools,omitempty"`

	// Configuration for each tool used by the agent.
	ToolResources map[string]any `json:"tool_resources,omitempty"`
}

type CreateResponse struct {
	Status string `json:"status"`
}

// Create creates a new Cortex Agent in the specified database and schema.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#create-cortex-agent for more details.
func (s *AgentObjectService) Create(ctx context.Context, agent *CortexAgent, opts *CreateOptions) (*CreateResponse, *Response, error) {
	if agent == nil {
		return nil, nil, fmt.Errorf("agent cannot be nil")
	}
	if agent.DatabaseName == nil || *agent.DatabaseName == "" {
		return nil, nil, fmt.Errorf("database name cannot be empty")
	}
	if agent.SchemaName == nil || *agent.SchemaName == "" {
		return nil, nil, fmt.Errorf("schema name cannot be empty")
	}
	if agent.Name == nil || *agent.Name == "" {
		return nil, nil, fmt.Errorf("agent name cannot be empty")
	}

	// Validate the CreateMode option if provided
	if opts != nil && !ValidateCreateMode(opts.CreateMode) {
		return nil, nil, fmt.Errorf("invalid createMode: %q, must be one of: %q, %q, %q, or empty",
			opts.CreateMode, CreateModeErrorIfExists, CreateModeOrReplace, CreateModeIfNotExists)
	}

	// Construct the URL for creating the agent
	u := fmt.Sprintf("databases/%s/schemas/%s/agents",
		url.PathEscape(*agent.DatabaseName),
		url.PathEscape(*agent.SchemaName))

	var err error
	u, err = addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	var req *http.Request
	req, err = s.client.NewRequest(http.MethodPost, u, agent)
	if err != nil {
		return nil, nil, err
	}

	var (
		resp *Response
		cr   *CreateResponse
	)

	resp, err = s.client.Do(ctx, req, &cr)
	if err != nil {
		return nil, nil, err
	}

	return cr, resp, nil
}

// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#update-cortex-agent for more details.

// ValidateIfExists checks if the given ifExists value is either "true", or "false".
// Returns true if valid, false otherwise.
func ValidateIfExists(ifExists string) bool {
	return ifExists == "true" || ifExists == "false"
}

// DeleteOptions is used to specify the behavior when deleting a Cortex Agent.
// Valid values: "true", "false"
type DeleteOptions struct {
	IfExists string `url:"ifExists,omitempty"`
}

type DeleteResponse struct {
	Status string `json:"status"`
}

// Delete deletes a Cortex Agent with the specified name. If the ifExists parameter is set to true, the operation succeeds
// even if the agent does not exist. Otherwise, the operation fails if the agent cannot be deleted.
// See https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-agents-rest-api#delete-cortex-agent for more details.
func (s *AgentObjectService) Delete(ctx context.Context, database, schema, name string, opts *DeleteOptions) (*DeleteResponse, *Response, error) {
	if database == "" {
		return nil, nil, fmt.Errorf("database cannot be empty")
	}
	if schema == "" {
		return nil, nil, fmt.Errorf("schema cannot be empty")
	}
	if name == "" {
		return nil, nil, fmt.Errorf("agent name cannot be empty")
	}

	// Validate the IfExists option if provided
	if opts != nil && !ValidateIfExists(opts.IfExists) {
		return nil, nil, fmt.Errorf(`invalid ifExists: %q, must be "true" or "false"`, opts.IfExists)
	}

	// Construct the URL for deleting the agent
	u := fmt.Sprintf("databases/%s/schemas/%s/agents/%s",
		url.PathEscape(database),
		url.PathEscape(schema),
		url.PathEscape(name))

	var err error
	u, err = addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	var req *http.Request
	req, err = s.client.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var (
		resp *Response
		dr   *DeleteResponse
	)
	resp, err = s.client.Do(ctx, req, &dr)
	if err != nil {
		return nil, nil, err
	}

	return dr, resp, nil
}
