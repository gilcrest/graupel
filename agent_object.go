package graupel

import (
	"time"
)

type CortexAgent struct {
	Owner         *string                  `json:"owner,omitempty"`
	DatabaseName  *string                  `json:"database_name,omitempty"`
	SchemaName    *string                  `json:"schema_name,omitempty"`
	Name          *string                  `json:"name"`
	Comment       *string                  `json:"comment,omitempty"`
	Profile       *AgentProfile            `json:"profile,omitempty"`
	Models        *ModelConfig             `json:"models,omitempty"`
	Instructions  *AgentInstructions       `json:"instructions,omitempty"`
	Orchestration *OrchestrationConfig     `json:"orchestration,omitempty"`
	Tools         []*Tool                  `json:"tools,omitempty"`
	ToolResources map[string]*ToolResource `json:"tool_resources,omitempty"`
	CreatedOn     time.Time                `json:"created_on,omitempty"`
}

// AgentInstructions contains the various instruction types for the agent.
type AgentInstructions struct {
	// Instructions for response generation.
	Response *string `json:"response,omitempty"`
	// These custom instructions are used when the agent is planning which tools to use.
	Orchestration *string `json:"orchestration,omitempty"`
	// System instructions for the agent.
	System *string `json:"system,omitempty"`
}

// AgentProfile is profile information for a Data Cortex agent.
type AgentProfile struct {
	// Display name for the agent.
	DisplayName *string `json:"display_name,omitempty"`
}

// BudgetConfig specifies time and token budgets for agent operations.
type BudgetConfig struct {
	// Time budget in seconds.
	Seconds *int `json:"seconds,omitempty"`
	// Token budget.
	Tokens *int `json:"tokens,omitempty"`
}

// ExecutionEnvironment  is the configuration for server-executed tools.
type ExecutionEnvironment struct {
	// The type of execution environment, currently only warehouse is supported.
	Type *string `json:"type"`
	// The name of the warehouse. Case-sensitive, if it is an unquoted identifier, provide the name in all-caps.
	Warehouse *string `json:"warehouse"`
	// The query timeout in seconds
	QueryTimeout *int `json:"query_timeout,omitempty"` // Query timeout in seconds
}

// ModelConfig specifies the models used by the agent.
type ModelConfig struct {
	// Model to use for orchestration. If not provided, a model is automatically selected.
	Orchestration *string `json:"orchestration,omitempty"`
}

type OrchestrationConfig struct {
	// Budget configuration for orchestration.
	Budget *BudgetConfig `json:"budget,omitempty"`
}

// Tool defines a tool that can be used by the agent. Tools provide specific
// capabilities like data analysis, search, or generic functions.
type Tool struct {
	ToolSpec *ToolSpec `json:"tool_spec,omitempty"`
}

// ToolInputSchema represents the schema for tool inputs.
type ToolInputSchema struct {
	// The type of the input schema object, e.g., "object", "array", "string", etc.
	Type *string `json:"type"`
	// A description of what the input is.
	Description *string `json:"description,omitempty"`
	// If type is object, definitions of each input parameter
	Properties map[string]ToolInputSchema `json:"properties,omitempty"`
	// If type is array, the schema for the elements of the array
	Items *ToolInputSchema `json:"items,omitempty"`
	// If type is object, list of required input parameter names
	Required []string `json:"required,omitempty"`
}

type ToolResource interface {
	toolResourceName() string
}

// CortexAnalystTextToSQLToolResource represents configuration for text-to-SQL
// analysis tool. Provides parameters for SQL query generation and execution.
// Exactly one of semantic_model_file or semantic_view must be provided.
type CortexAnalystTextToSQLToolResource struct {
	SemanticModelFile    string                `json:"semantic_model_file,omitempty"`   // The path to a file stored in a Snowflake Stage holding the semantic model yaml.
	SemanticView         string                `json:"semantic_view,omitempty"`         // The name of the Snowflake native semantic model object.
	ExecutionEnvironment *ExecutionEnvironment `json:"execution_environment,omitempty"` // Configuration for how to execute the generated SQL query.
}

func (t *CortexAnalystTextToSQLToolResource) toolResourceName() string {
	return "cortex_analyst_text_to_sql"
}

// CortexSearchToolResource represents Configuration for search functionality.
// Defines how document search and retrieval should be performed.
type CortexSearchToolResource struct {
	// The fully qualified name of the search service.
	SearchService *string `json:"search_service"`
	// The title column of the document.
	TitleColumn *string `json:"title_column"`
	// The ID column of the document.
	IDColumn *string `json:"id_column"`
	// Filter query for search results.
	Filter map[string]map[string]string `json:"filter"`
}

func (t *CortexSearchToolResource) toolResourceName() string {
	return "cortex_search"
}

// GenericToolResource represents a custom function or stored procedure tool.
type GenericToolResource struct {
	Type                 *string               `json:"type"`
	ExecutionEnvironment *ExecutionEnvironment `json:"execution_environment,omitempty"`
	Identifier           *string               `json:"identifier,omitempty"`
}

func (t *GenericToolResource) toolResourceName() string {
	return "generic"
}

// ToolSpec is the specification of the tool's type, configuration, and input requirements.
type ToolSpec struct {
	// The type of tool capability. Can be specialized types like ‘cortex_analyst_text_to_sql’ or ‘generic’ for general-purpose tools.
	Type string `json:"type"`
	// Unique identifier for referencing this tool instance. Used to match with configuration in tool_resources.
	Name string `json:"name"`
	// Description of the tool to be considered for tool use.
	Description string `json:"description,omitempty"`
	// JSON Schema definition of the expected input parameters for this tool.
	// This will be fed to the agent so it knows the structure it should follow
	// for when generating the input for ToolUses. Required for generic tools
	// to specify their input parameters.
	InputSchema *ToolInputSchema `json:"input_schema"`
}
