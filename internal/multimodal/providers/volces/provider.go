package volces

import (
	"fmt"

	mmtypes "github.com/seaart/sa-go/internal/multimodal/types"
)

type Factory struct{}

func NewFactory() Factory { return Factory{} }

type SeedreamBuilder struct {
	req *mmtypes.GenerateRequest
}

func newSeedreamBuilder(model string) *SeedreamBuilder {
	return &SeedreamBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *SeedreamBuilder) Prompt(prompt string) *SeedreamBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *SeedreamBuilder) Size(size string) *SeedreamBuilder {
	b.req.Input[0].Params["size"] = size
	return b
}

func (b *SeedreamBuilder) SizeWH(width, height int) *SeedreamBuilder {
	return b.Size(fmt.Sprintf("%dx%d", width, height))
}

func (b *SeedreamBuilder) ReferenceImage(url string) *SeedreamBuilder {
	b.req.Input[0].Params["image"] = url
	return b
}

func (b *SeedreamBuilder) Seed(seed int64) *SeedreamBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *SeedreamBuilder) Param(key string, value any) *SeedreamBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *SeedreamBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type JimengBuilder struct {
	req *mmtypes.GenerateRequest
}

func newJimengBuilder(model string) *JimengBuilder {
	return &JimengBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *JimengBuilder) Prompt(prompt string) *JimengBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *JimengBuilder) Size(width, height int) *JimengBuilder {
	b.req.Input[0].Params["width"] = width
	b.req.Input[0].Params["height"] = height
	return b
}

func (b *JimengBuilder) Seed(seed int64) *JimengBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *JimengBuilder) Param(key string, value any) *JimengBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *JimengBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type SeedEditBuilder struct {
	req *mmtypes.GenerateRequest
}

func newSeedEditBuilder(model string) *SeedEditBuilder {
	return &SeedEditBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *SeedEditBuilder) Prompt(prompt string) *SeedEditBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *SeedEditBuilder) ImageURLs(urls ...string) *SeedEditBuilder {
	b.req.Input[0].Params["image_urls"] = urls
	return b
}

func (b *SeedEditBuilder) Size(width, height int) *SeedEditBuilder {
	b.req.Input[0].Params["width"] = width
	b.req.Input[0].Params["height"] = height
	return b
}

func (b *SeedEditBuilder) Scale(scale float64) *SeedEditBuilder {
	b.req.Input[0].Params["scale"] = scale
	return b
}

func (b *SeedEditBuilder) Seed(seed int64) *SeedEditBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *SeedEditBuilder) Param(key string, value any) *SeedEditBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *SeedEditBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type SeedEditI2IBuilder struct {
	req *mmtypes.GenerateRequest
}

func newSeedEditI2IBuilder(model string) *SeedEditI2IBuilder {
	return &SeedEditI2IBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *SeedEditI2IBuilder) Prompt(prompt string) *SeedEditI2IBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *SeedEditI2IBuilder) ImageURL(url string) *SeedEditI2IBuilder {
	b.req.Input[0].Params["image"] = url
	return b
}

func (b *SeedEditI2IBuilder) Size(size string) *SeedEditI2IBuilder {
	b.req.Input[0].Params["size"] = size
	return b
}

func (b *SeedEditI2IBuilder) Scale(scale float64) *SeedEditI2IBuilder {
	b.req.Input[0].Params["scale"] = scale
	return b
}

func (b *SeedEditI2IBuilder) Seed(seed int64) *SeedEditI2IBuilder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *SeedEditI2IBuilder) Param(key string, value any) *SeedEditI2IBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *SeedEditI2IBuilder) Build() *mmtypes.GenerateRequest { return b.req }

func (Factory) Seedream5() *SeedreamBuilder  { return newSeedreamBuilder("volces_seedream_5") }
func (Factory) Seedream45() *SeedreamBuilder { return newSeedreamBuilder("volces_seedream_4_5") }
func (Factory) Jimeng31() *JimengBuilder     { return newJimengBuilder("volces_jimeng_3_1") }
func (Factory) Jimeng30() *JimengBuilder     { return newJimengBuilder("volces_jimeng_3_0") }
func (Factory) SeedEdit3() *SeedEditBuilder  { return newSeedEditBuilder("volces_seededit_3_0") }
func (Factory) JimengI2I30() *SeedEditBuilder {
	return newSeedEditBuilder("volces_jimeng_i2i_3_0")
}
func (Factory) SeedEdit3I2I() *SeedEditI2IBuilder {
	return newSeedEditI2IBuilder("volces_seededit_3_0_i2i")
}
