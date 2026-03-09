package llm

// RegisterOpenAI registers the OpenAI provider with the client
func RegisterOpenAI(client *Client, config OpenAIConfig) {
	client.SetProvider(ProviderOpenAI, NewOpenAIProvider(config))
}

// RegisterAnthropic registers the Anthropic provider with the client
// Note: This is also available directly in anthropic.go, but provided here
// for consistency with the RegisterOpenAI pattern.
func RegisterAnthropic(client *Client, config AnthropicConfig) {
	client.SetProvider(ProviderAnthropic, NewAnthropicProvider(config))
}

// RegisterOllama registers the Ollama provider with the client
func RegisterOllama(client *Client, config OllamaConfig) {
	client.SetProvider(ProviderOllama, NewOllamaProvider(config))
}

// RegisterZAI registers the z.ai provider with the client
func RegisterZAI(client *Client, config ZAIConfig) {
	client.SetProvider(ProviderZAI, NewZAIProvider(config))
}
