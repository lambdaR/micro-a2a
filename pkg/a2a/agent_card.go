package a2a

// AgentCard conveys key information about an agent
type AgentCard struct {
	// Human readable name of the agent
	Name string `json:"name"`

	// Human-readable description of the agent
	Description string `json:"description"`

	// URL to the address the agent is hosted at
	URL string `json:"url"`

	// The service provider of the agent
	Provider *AgentProvider `json:"provider,omitempty"`

	// The version of the agent - format is up to the provider
	Version string `json:"version"`

	// URL to documentation for the agent
	DocumentationURL string `json:"documentationUrl,omitempty"`

	// Optional capabilities supported by the agent
	Capabilities *AgentCapabilities `json:"capabilities"`

	// Security scheme details used for authenticating with this agent.
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`

	// Security requirements for contacting the agent.
	Security []map[string][]string `json:"security,omitempty"`

	// Supported mime types for input
	DefaultInputModes []string `json:"defaultInputModes"`

	// Supported mime types for output
	DefaultOutputModes []string `json:"defaultOutputModes"`

	// Skills are a unit of capability that an agent can perform
	Skills []AgentSkill `json:"skills"`
}

// Provider represents the service provider information
type AgentProvider struct {
	Organization string `json:"organization"`
	URL          string `json:"url"`
}

// AgentCapabilities describes optional capabilities supported by the agent
type AgentCapabilities struct {
	// True if the agent supports SSE
	Streaming bool `json:"streaming,omitempty"`

	// True if the agent can notify updates to client
	PushNotifications bool `json:"pushNotifications,omitempty"`

	// True if the agent exposes status change history for tasks
	StateTransitionHistory bool `json:"stateTransitionHistory,omitempty"`
}

// AgentSkill represents a unit of capability that an agent can perform
type AgentSkill struct {
	// Unique identifier for the agent's skill
	ID string `json:"id"`

	// Human readable name of the skill
	Name string `json:"name"`

	// Description of the skill
	Description string `json:"description"`

	// Tagwords describing classes of capabilities
	Tags []string `json:"tags"`

	// Example scenarios that the skill can perform
	Examples []string `json:"examples,omitempty"`

	// Supported mime types for input (if different than default)
	InputModes []string `json:"inputModes,omitempty"`

	// Supported mime types for output (if different than default)
	OutputModes []string `json:"outputModes,omitempty"`
}

type SecuritySchemeType string

// supported Auth schemas
const (
	APIKeySecurity        SecuritySchemeType = "apiKey"
	HTTPAuthSecurity      SecuritySchemeType = "http"
	OAuth2Security        SecuritySchemeType = "oauth2"
	OpenIdConnectSecurity SecuritySchemeType = "openIdConnect"
)

// SecurityScheme is an interface that all SecurityScheme types must implement.
// It serves as a marker interface to group different SecurityScheme types
// that can be used in AgentCard
//
// Types that implement SecurityScheme are:
// - HTTPAuthSecurityScheme
// - OAuth2SecurityScheme
// - OpenIdConnectSecurityScheme
// - APIKeySecurityScheme
type SecurityScheme interface {
	// ssGlue is a marker method that doesn't do anything but
	// ensures type safety when working with different SecurityScheme types
	ssGlue()
}

// HTTPAuthSecurityScheme represents HTTP Authentication security scheme
type HTTPAuthSecurityScheme struct {
	// Type is always "http" for HTTP Authentication
	Type string `json:"type"`
	
	// The name of the HTTP Authentication scheme to be used in the Authorization header
	Scheme string `json:"scheme"`
	
	// A hint to the client to identify how the bearer token is formatted
	BearerFormat string `json:"bearerFormat,omitempty"`
	
	// Description of this security scheme
	Description string `json:"description,omitempty"`
}

func (ss HTTPAuthSecurityScheme) ssGlue() {}

// OAuth2SecurityScheme represents OAuth2 security scheme
type OAuth2SecurityScheme struct {
	// Type is always "oauth2" for OAuth2 Authentication
	Type string `json:"type"`
	
	// Description of this security scheme
	Description string `json:"description,omitempty"`
	
	// Configuration for the supported OAuth Flows
	Flows OAuth2Flows `json:"flows"`
}

func (ss OAuth2SecurityScheme) ssGlue() {}

// OAuth2Flows contains configuration details for supported OAuth flows
type OAuth2Flows struct {
	// Configuration for the OAuth Implicit flow
	Implicit *ImplicitOAuthFlow `json:"implicit,omitempty"`
	
	// Configuration for the OAuth Authorization Code flow
	AuthorizationCode *AuthorizationCodeOAuthFlow `json:"authorizationCode,omitempty"`
	
	// Configuration for the OAuth Client Credentials flow
	ClientCredentials *ClientCredentialsOAuthFlow `json:"clientCredentials,omitempty"`
	
	// Configuration for the OAuth Password flow
	Password *PasswordOAuthFlow `json:"password,omitempty"`
}

// ImplicitOAuthFlow contains configuration details for OAuth Implicit flow
type ImplicitOAuthFlow struct {
	// The authorization URL to be used for this flow
	AuthorizationURL string `json:"authorizationUrl"`
	
	// The URL to be used for obtaining refresh tokens
	RefreshURL string `json:"refreshUrl,omitempty"`
	
	// The available scopes for the OAuth2 security scheme
	Scopes map[string]string `json:"scopes"`
}

// AuthorizationCodeOAuthFlow contains configuration details for OAuth Authorization Code flow
type AuthorizationCodeOAuthFlow struct {
	// The authorization URL to be used for this flow
	AuthorizationURL string `json:"authorizationUrl"`
	
	// The token URL to be used for this flow
	TokenURL string `json:"tokenUrl"`
	
	// The URL to be used for obtaining refresh tokens
	RefreshURL string `json:"refreshUrl,omitempty"`
	
	// The available scopes for the OAuth2 security scheme
	Scopes map[string]string `json:"scopes"`
}

// ClientCredentialsOAuthFlow contains configuration details for OAuth Client Credentials flow
type ClientCredentialsOAuthFlow struct {
	// The token URL to be used for this flow
	TokenURL string `json:"tokenUrl"`
	
	// The URL to be used for obtaining refresh tokens
	RefreshURL string `json:"refreshUrl,omitempty"`
	
	// The available scopes for the OAuth2 security scheme
	Scopes map[string]string `json:"scopes"`
}

// PasswordOAuthFlow contains configuration details for OAuth Password flow
type PasswordOAuthFlow struct {
	// The token URL to be used for this flow
	TokenURL string `json:"tokenUrl"`
	
	// The URL to be used for obtaining refresh tokens
	RefreshURL string `json:"refreshUrl,omitempty"`
	
	// The available scopes for the OAuth2 security scheme
	Scopes map[string]string `json:"scopes"`
}

// OpenIdConnectSecurityScheme represents OpenID Connect security scheme
type OpenIdConnectSecurityScheme struct {
	// Type is always "openIdConnect" for OpenID Connect
	Type string `json:"type"`
	
	// OpenId Connect URL to discover OAuth2 configuration values
	OpenIdConnectURL string `json:"openIdConnectUrl"`
	
	// Description of this security scheme
	Description string `json:"description,omitempty"`
}

func (ss OpenIdConnectSecurityScheme) ssGlue() {}

// APIKeySecurityScheme represents API Key security scheme
type APIKeySecurityScheme struct {
	// Type is always "apiKey" for API Key Authentication
	Type string `json:"type"`
	
	// The name of the header, query or cookie parameter to be used
	Name string `json:"name"`
	
	// The location of the API key (query, header, cookie)
	In string `json:"in"`
	
	// Description of this security scheme
	Description string `json:"description,omitempty"`
}

func (ss APIKeySecurityScheme) ssGlue() {}
