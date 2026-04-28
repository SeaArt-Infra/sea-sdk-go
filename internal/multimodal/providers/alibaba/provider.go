package alibaba

import mmtypes "github.com/seaart/sa-go/internal/multimodal/types"

type Factory struct{}

func NewFactory() Factory { return Factory{} }

type ImageBuilder struct {
	req      *mmtypes.GenerateRequest
	messages []any
}

func newImageBuilder(model string) *ImageBuilder {
	return &ImageBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *ImageBuilder) Prompt(prompt string) *ImageBuilder {
	b.messages = append(b.messages, prompt)
	return b
}

func (b *ImageBuilder) InputImage(url string) *ImageBuilder {
	b.messages = append(b.messages, url)
	return b
}

func (b *ImageBuilder) Size(size string) *ImageBuilder {
	b.req.Input[0].Params["size"] = size
	return b
}

func (b *ImageBuilder) N(n int) *ImageBuilder {
	b.req.Input[0].Params["n"] = n
	return b
}

func (b *ImageBuilder) Seed(seed int) *ImageBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *ImageBuilder) ThinkingMode(enabled bool) *ImageBuilder {
	b.req.Input[0].Params["thinking_mode"] = enabled
	return b
}

func (b *ImageBuilder) Watermark(enabled bool) *ImageBuilder {
	b.req.Input[0].Params["watermark"] = enabled
	return b
}

func (b *ImageBuilder) Param(key string, value any) *ImageBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *ImageBuilder) Build() *mmtypes.GenerateRequest {
	b.req.Input[0].Params["messages"] = []map[string]any{
		{"role": "user", "content": b.messages},
	}
	return b.req
}

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

func (b *T2VBuilder) AudioURL(url string) *T2VBuilder {
	b.req.Input[0].Params["audio_url"] = url
	return b
}

func (b *T2VBuilder) Duration(seconds int) *T2VBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *T2VBuilder) Size(size string) *T2VBuilder {
	b.req.Input[0].Params["size"] = size
	return b
}

func (b *T2VBuilder) Audio(enabled bool) *T2VBuilder {
	b.req.Input[0].Params["audio"] = enabled
	return b
}

func (b *T2VBuilder) ShotType(shotType string) *T2VBuilder {
	b.req.Input[0].Params["shot_type"] = shotType
	return b
}

func (b *T2VBuilder) Seed(seed int) *T2VBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *T2VBuilder) NegativePrompt(text string) *T2VBuilder {
	b.req.Input[0].Params["negative_prompt"] = text
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

func (b *I2VBuilder) ImgURL(url string) *I2VBuilder {
	b.req.Input[0].Params["img_url"] = url
	return b
}

func (b *I2VBuilder) AudioURL(url string) *I2VBuilder {
	b.req.Input[0].Params["audio_url"] = url
	return b
}

func (b *I2VBuilder) Resolution(res string) *I2VBuilder {
	b.req.Input[0].Params["resolution"] = res
	return b
}

func (b *I2VBuilder) Duration(seconds int) *I2VBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *I2VBuilder) Seed(seed int) *I2VBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *I2VBuilder) Audio(enabled bool) *I2VBuilder {
	b.req.Input[0].Params["audio"] = enabled
	return b
}

func (b *I2VBuilder) NegativePrompt(text string) *I2VBuilder {
	b.req.Input[0].Params["negative_prompt"] = text
	return b
}

func (b *I2VBuilder) ShotType(shotType string) *I2VBuilder {
	b.req.Input[0].Params["shot_type"] = shotType
	return b
}

func (b *I2VBuilder) Param(key string, value any) *I2VBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *I2VBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type ReferenceBuilder struct {
	req  *mmtypes.GenerateRequest
	refs []string
}

func newReferenceBuilder(model string) *ReferenceBuilder {
	return &ReferenceBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *ReferenceBuilder) Prompt(prompt string) *ReferenceBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *ReferenceBuilder) ReferenceURLs(urls ...string) *ReferenceBuilder {
	b.refs = append(b.refs, urls...)
	return b
}

func (b *ReferenceBuilder) AudioURL(url string) *ReferenceBuilder {
	b.req.Input[0].Params["audio_url"] = url
	return b
}

func (b *ReferenceBuilder) Duration(seconds int) *ReferenceBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *ReferenceBuilder) Size(size string) *ReferenceBuilder {
	b.req.Input[0].Params["size"] = size
	return b
}

func (b *ReferenceBuilder) Seed(seed int) *ReferenceBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *ReferenceBuilder) Audio(enabled bool) *ReferenceBuilder {
	b.req.Input[0].Params["audio"] = enabled
	return b
}

func (b *ReferenceBuilder) ShotType(shotType string) *ReferenceBuilder {
	b.req.Input[0].Params["shot_type"] = shotType
	return b
}

func (b *ReferenceBuilder) NegativePrompt(text string) *ReferenceBuilder {
	b.req.Input[0].Params["negative_prompt"] = text
	return b
}

func (b *ReferenceBuilder) Param(key string, value any) *ReferenceBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *ReferenceBuilder) Build() *mmtypes.GenerateRequest {
	if len(b.refs) > 0 {
		b.req.Input[0].Params["reference_urls"] = b.refs
	}
	return b.req
}

func (Factory) Wan27ImagePro() *ImageBuilder { return newImageBuilder("alibaba_wan27_image_pro") }
func (Factory) Wanx26T2V() *T2VBuilder       { return newT2VBuilder("alibaba_wanx26_t2v") }
func (Factory) Wanx26I2V() *I2VBuilder       { return newI2VBuilder("alibaba_wanx26_i2v") }
func (Factory) Wanx26Reference() *ReferenceBuilder {
	return newReferenceBuilder("alibaba_wanx26_reference")
}
