package graupel

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CortexAgent represents a Cortex Agent in Snowflake.
type CortexAgent struct {
	// The owner of the agent.
	Owner *string `json:"owner,omitempty"`

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
	ToolResources map[string]ToolResource `json:"tool_resources,omitempty"`

	// String representation of the agent specification.
	AgentSpecStr *string `json:"agent_spec,omitempty"`

	// Timestamp when the agent was created.
	CreatedOn time.Time `json:"created_on,omitempty"`
}

func (ca CortexAgent) String() string {
	return Stringify(ca)
}

type AgentSpec struct {
	// Model to use for orchestration. If not provided, a model is automatically selected.
	Models *ModelConfig `json:"models,omitempty"`

	// Orchestration configuration for the agent - currently only budget is supported.
	Orchestration *OrchestrationConfig `json:"orchestration,omitempty"`

	// AgentInstructions contains the various instruction types for the agent.
	Instructions *AgentInstructions `json:"instructions,omitempty"`

	// List of tools available to the agent.
	Tools []*Tool `json:"tools,omitempty"`

	// Message to return when a tool is unable to answer the user's query.
	ToolUnableToAnswer *string `json:"tool_unable_to_answer,omitempty"`

	// Configuration for each tool used by the agent.
	ToolResources map[string]any `json:"tool_resources,omitempty"`
	//ToolResources map[string]*ToolResource `json:"tool_resources,omitempty"`

	// Who knows what else might be added in the future
	Experimental map[string]any `json:"experimental,omitempty"`
}

func (as AgentSpec) String() string {
	return Stringify(as)
}

// SampleQuestion is a sample question to guide the agent's behavior.
type SampleQuestion struct {
	Question *string `json:"question,omitempty"`
}

func (sq SampleQuestion) String() string {
	return Stringify(sq)
}

// AgentInstructions contains the various instruction types for the agent.
type AgentInstructions struct {
	// Instructions for response generation.
	Response *string `json:"response,omitempty"`
	// These custom instructions are used when the agent is planning which tools to use.
	Orchestration *string `json:"orchestration,omitempty"`
	// System instructions for the agent.
	System *string `json:"system,omitempty"`
	// Sample questions to guide the agent's behavior.
	SampleQuestions []*SampleQuestion `json:"sample_questions,omitempty"`
}

func (ai AgentInstructions) String() string {
	return Stringify(ai)
}

// AgentProfile is profile information for a Data Cortex agent.
type AgentProfile struct {
	// Display name for the agent.
	DisplayName *string `json:"display_name,omitempty"`
}

func (ap AgentProfile) String() string {
	return Stringify(ap)
}

// BudgetConfig specifies time and token budgets for agent operations.
type BudgetConfig struct {
	// Time budget in seconds.
	Seconds *int `json:"seconds,omitempty"`
	// Token budget.
	Tokens *int `json:"tokens,omitempty"`
}

func (bc BudgetConfig) String() string {
	return Stringify(bc)
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

func (ee ExecutionEnvironment) String() string {
	return Stringify(ee)
}

// ModelConfig specifies the models used by the agent.
type ModelConfig struct {
	// Model to use for orchestration. If not provided, a model is automatically selected.
	Orchestration *string `json:"orchestration,omitempty"`
}

func (mc ModelConfig) String() string {
	return Stringify(mc)
}

type OrchestrationConfig struct {
	// Budget configuration for orchestration.
	Budget *BudgetConfig `json:"budget,omitempty"`
}

func (oc OrchestrationConfig) String() string {
	return Stringify(oc)
}

// Tool defines a tool that can be used by the agent. Tools provide specific
// capabilities like data analysis, search, or generic functions.
type Tool struct {
	ToolSpec *ToolSpec `json:"tool_spec,omitempty"`
}

func (t Tool) String() string {
	return Stringify(t)
}

type ObjectInputParameter struct {
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (oip ObjectInputParameter) String() string {
	return Stringify(oip)
}

type ArrayInputParameter struct {
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (aip ArrayInputParameter) String() string {
	return Stringify(aip)
}

// ToolInputSchema represents the schema for tool inputs.
type ToolInputSchema struct {
	// The type of the input schema object, e.g., "object", "array", "string", etc.
	Type *string `json:"type"`
	// A description of what the input is.
	Description *string `json:"description,omitempty"`
	// If type is object, definitions of each input parameter
	Properties map[string]*ObjectInputParameter `json:"properties,omitempty"`
	// If type is array, the schema for the elements of the array
	Items *ArrayInputParameter `json:"items,omitempty"`
	// If type is object, list of required input parameter names
	Required []string `json:"required,omitempty"`
}

func (tis ToolInputSchema) String() string {
	return Stringify(tis)
}

type ToolResource interface {
	toolResourceName() string
}

// CortexAnalystTextToSQLToolResource represents configuration for text-to-SQL
// analysis tool. Provides parameters for SQL query generation and execution.
// Exactly one of semantic_model_file or semantic_view must be provided.
type CortexAnalystTextToSQLToolResource struct {
	SemanticModelFile    *string               `json:"semantic_model_file,omitempty"`   // The path to a file stored in a Snowflake Stage holding the semantic model yaml.
	SemanticView         *string               `json:"semantic_view,omitempty"`         // The name of the Snowflake native semantic model object.
	ExecutionEnvironment *ExecutionEnvironment `json:"execution_environment,omitempty"` // Configuration for how to execute the generated SQL query.
}

func (cattstr CortexAnalystTextToSQLToolResource) String() string {
	return Stringify(cattstr)
}

func (cattstr CortexAnalystTextToSQLToolResource) toolResourceName() string {
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

func (cstr CortexSearchToolResource) String() string {
	return Stringify(cstr)
}

func (cstr CortexSearchToolResource) toolResourceName() string {
	return "cortex_search"
}

// GenericToolResource represents a custom function or stored procedure tool.
type GenericToolResource struct {
	Type                 *string               `json:"type"`
	ExecutionEnvironment *ExecutionEnvironment `json:"execution_environment,omitempty"`
	Identifier           *string               `json:"identifier,omitempty"`
}

func (gtr GenericToolResource) String() string {
	return Stringify(gtr)
}

func (gtr GenericToolResource) toolResourceName() string {
	return "generic"
}

// ParseToolResources parses the raw tool resources map into a typed map of ToolResources
// based on the tool types defined in the provided tools slice.
func ParseToolResources(toolResources map[string]any, tools []*Tool) (map[string]ToolResource, error) {
	// Build a case-insensitive map of tool name -> type
	toolTypeMap := make(map[string]string)

	for _, tool := range tools {
		if tool.ToolSpec != nil && tool.ToolSpec.Name != nil {
			toolTypeMap[strings.ToLower(*tool.ToolSpec.Name)] = *tool.ToolSpec.Type
		}
	}

	// Prepare result map
	result := make(map[string]ToolResource)

	for name, rawResource := range toolResources {
		// Look up the tool type using the name of the Tool Resource (using case-insensitive matching)
		toolType, found := toolTypeMap[strings.ToLower(name)]
		if !found {
			return nil, fmt.Errorf("no matching tool spec found for tool_resource %q", name)
		}

		// Marshal back to JSON to unmarshal into specific type
		rawJSON, err := json.Marshal(rawResource)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool resource %s: %w", name, err)
		}

		// Based on the matched tool type, unmarshal into specific struct
		var resource ToolResource
		switch toolType {
		case "cortex_analyst_text_to_sql":
			var cattstr CortexAnalystTextToSQLToolResource
			if err := json.Unmarshal(rawJSON, &cattstr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal CortexAnalystTextToSQLToolResource for %s: %w", name, err)
			}
			resource = cattstr

		case "cortex_search":
			var cstr CortexSearchToolResource
			if err := json.Unmarshal(rawJSON, &cstr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal CortexSearchToolResource for %s: %w", name, err)
			}
			resource = cstr

		case "generic":
			var gtr GenericToolResource
			if err := json.Unmarshal(rawJSON, &gtr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal GenericToolResource for %s: %w", name, err)
			}
			resource = gtr

		default:
			return nil, fmt.Errorf("unknown tool type %q for resource %s", toolType, name)
		}

		result[name] = resource
	}

	return result, nil
}

// ToolSpec is the specification of the tool's type, configuration, and input requirements.
type ToolSpec struct {
	// The type of tool capability. Can be specialized types like ‘cortex_analyst_text_to_sql’ or ‘generic’ for general-purpose tools.
	Type *string `json:"type"`
	// Unique identifier for referencing this tool instance. Used to match with configuration in tool_resources.
	Name *string `json:"name"`
	// Description of the tool to be considered for tool use.
	Description *string `json:"description,omitempty"`
	// JSON Schema definition of the expected input parameters for this tool.
	// This will be fed to the agent so it knows the structure it should follow
	// for when generating the input for ToolUses. Required for generic tools
	// to specify their input parameters.
	InputSchema *ToolInputSchema `json:"input_schema"`
}

func (ts ToolSpec) String() string {
	return Stringify(ts)
}
