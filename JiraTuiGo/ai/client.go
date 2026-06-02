package ai

// AiClient is the interface for AI text generation backends.
type AiClient interface {
	Generate(system, user string) (string, error)
}
