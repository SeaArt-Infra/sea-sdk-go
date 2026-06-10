package vidu

import mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"

type Factory struct{}

func NewFactory() Factory { return Factory{} }

type T2VBuilder struct {
	req *mmtypes.GenerateRequest
}

func newT2VBuilder(model string) *T2VBuilder {
	return &T2VBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *T2VBuilder) Prompt(prompt string) *T2VBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *T2VBuilder) Duration(seconds int) *T2VBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *T2VBuilder) AspectRatio(ratio string) *T2VBuilder {
	b.req.Input[0].Params["aspect_ratio"] = ratio
	return b
}

func (b *T2VBuilder) Resolution(res string) *T2VBuilder {
	b.req.Input[0].Params["resolution"] = res
	return b
}

func (b *T2VBuilder) Seed(seed int) *T2VBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *T2VBuilder) Param(key string, value any) *T2VBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *T2VBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type I2VBuilder struct {
	req *mmtypes.GenerateRequest
}

func newI2VBuilder(model string) *I2VBuilder {
	return &I2VBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *I2VBuilder) Prompt(prompt string) *I2VBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *I2VBuilder) ImageURL(url string) *I2VBuilder {
	b.req.Input[0].Params["image_url"] = url
	return b
}

func (b *I2VBuilder) Duration(seconds int) *I2VBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *I2VBuilder) Resolution(res string) *I2VBuilder {
	b.req.Input[0].Params["resolution"] = res
	return b
}

func (b *I2VBuilder) Seed(seed int) *I2VBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *I2VBuilder) Param(key string, value any) *I2VBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *I2VBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type ReferenceBuilder struct {
	req  *mmtypes.GenerateRequest
	urls []string
}

func newReferenceBuilder(model string) *ReferenceBuilder {
	return &ReferenceBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *ReferenceBuilder) Prompt(prompt string) *ReferenceBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *ReferenceBuilder) ImageURLs(urls ...string) *ReferenceBuilder {
	b.urls = append(b.urls, urls...)
	return b
}

func (b *ReferenceBuilder) Duration(seconds int) *ReferenceBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *ReferenceBuilder) Resolution(res string) *ReferenceBuilder {
	b.req.Input[0].Params["resolution"] = res
	return b
}

func (b *ReferenceBuilder) Seed(seed int) *ReferenceBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *ReferenceBuilder) Param(key string, value any) *ReferenceBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *ReferenceBuilder) Build() *mmtypes.GenerateRequest {
	if len(b.urls) > 0 {
		b.req.Input[0].Params["image_urls"] = b.urls
	}
	return b.req
}

func (Factory) Q3Pro() *T2VBuilder             { return newT2VBuilder("vidu_q3_pro") }
func (Factory) Q3ProI2V() *I2VBuilder          { return newI2VBuilder("vidu_q3_pro_i2v") }
func (Factory) Q3Reference() *ReferenceBuilder { return newReferenceBuilder("vidu_q3_reference") }
