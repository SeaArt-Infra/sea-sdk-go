package kling

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

func (b *T2VBuilder) Mode(mode string) *T2VBuilder {
	b.req.Input[0].Params["mode"] = mode
	return b
}

func (b *T2VBuilder) NegativePrompt(text string) *T2VBuilder {
	b.req.Input[0].Params["negative_prompt"] = text
	return b
}

func (b *T2VBuilder) CfgScale(scale float64) *T2VBuilder {
	b.req.Input[0].Params["cfg_scale"] = scale
	return b
}

func (b *T2VBuilder) Sound(onOff string) *T2VBuilder {
	b.req.Input[0].Params["sound"] = onOff
	return b
}

func (b *T2VBuilder) MultiShot(enabled bool) *T2VBuilder {
	b.req.Input[0].Params["multi_shot"] = enabled
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
	b.req.Input[0].Params["image"] = url
	return b
}

func (b *I2VBuilder) ImageTailURL(url string) *I2VBuilder {
	b.req.Input[0].Params["image_tail"] = url
	return b
}

func (b *I2VBuilder) Duration(seconds int) *I2VBuilder {
	b.req.Input[0].Params["duration"] = seconds
	return b
}

func (b *I2VBuilder) AspectRatio(ratio string) *I2VBuilder {
	b.req.Input[0].Params["aspect_ratio"] = ratio
	return b
}

func (b *I2VBuilder) Mode(mode string) *I2VBuilder {
	b.req.Input[0].Params["mode"] = mode
	return b
}

func (b *I2VBuilder) NegativePrompt(text string) *I2VBuilder {
	b.req.Input[0].Params["negative_prompt"] = text
	return b
}

func (b *I2VBuilder) CfgScale(scale float64) *I2VBuilder {
	b.req.Input[0].Params["cfg_scale"] = scale
	return b
}

func (b *I2VBuilder) Sound(onOff string) *I2VBuilder {
	b.req.Input[0].Params["sound"] = onOff
	return b
}

func (b *I2VBuilder) MultiShot(enabled bool) *I2VBuilder {
	b.req.Input[0].Params["multi_shot"] = enabled
	return b
}

func (b *I2VBuilder) ShotType(shotType string) *I2VBuilder {
	b.req.Input[0].Params["shot_type"] = shotType
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

func (Factory) V3() *T2VBuilder    { return newT2VBuilder("kling_v3") }
func (Factory) V3I2V() *I2VBuilder { return newI2VBuilder("kling_v3_i2v") }
