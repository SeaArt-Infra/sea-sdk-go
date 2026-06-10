package nano

import mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"

type Factory struct{}

func NewFactory() Factory { return Factory{} }

type BananaBuilder struct {
	req  *mmtypes.GenerateRequest
	urls []string
}

func newBananaBuilder(model string) *BananaBuilder {
	return &BananaBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *BananaBuilder) Prompt(prompt string) *BananaBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *BananaBuilder) AspectRatio(ratio string) *BananaBuilder {
	b.req.Input[0].Params["aspect_ratio"] = ratio
	return b
}

func (b *BananaBuilder) Resolution(res string) *BananaBuilder {
	b.req.Input[0].Params["resolution"] = res
	return b
}

func (b *BananaBuilder) ImageURLs(urls ...string) *BananaBuilder {
	b.urls = append(b.urls, urls...)
	return b
}

func (b *BananaBuilder) Param(key string, value any) *BananaBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *BananaBuilder) Build() *mmtypes.GenerateRequest {
	if len(b.urls) > 0 {
		b.req.Input[0].Params["image_urls"] = b.urls
	}
	return b.req
}

func (Factory) Banana2() *BananaBuilder { return newBananaBuilder("nano_banana_2") }
