// Package sa provides a unified SDK entrypoint for SeaArt APIs.
//
// The public package stays intentionally small:
//
//   - Client and shared options live in the root package.
//   - Modal APIs are exposed through Client.Modal.
//   - LLM APIs are exposed through Client.LLM.
//
// Internal implementation is organized for long-term growth:
//
//   - internal/transport: HTTP transport and request plumbing
//   - internal/shared: shared SDK primitives
//   - internal/multimodal: multimodal services, types, and providers
//   - internal/llm: LLM services and shared request/response types
package sa
