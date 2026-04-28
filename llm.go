package sa

import "github.com/seaart/sa-go/internal/transport"

// LLMService provides text-generation, reranking, embeddings, and model listing APIs.
type LLMService struct {
	client *transport.Client
}
