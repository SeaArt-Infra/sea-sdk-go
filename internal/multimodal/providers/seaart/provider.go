package seaart

import mmtypes "github.com/seaart/sa-go/internal/multimodal/types"

const VerSDXL = "c9090ffbe5649de2f34cfe5b865d50fe"

var modelVersion = map[string]string{
	"sdxl":          VerSDXL,
	"z_image":       VerSDXL,
	"z_image_turbo": VerSDXL,
}

var denoiseDefaults = map[string]float64{
	"z_image":       0.7,
	"z_image_turbo": 0.7,
}

type Factory struct{}

func NewFactory() Factory { return Factory{} }

type Builder struct {
	req *mmtypes.GenerateRequest
}

func newBuilder(model string) *Builder {
	req := mmtypes.NewGenerateRequest(model)
	if ver, ok := modelVersion[model]; ok {
		req.Input[0].Params["model_ver_no"] = ver
	}
	if def, ok := denoiseDefaults[model]; ok {
		req.Input[0].Params["denoise"] = def
	}
	return &Builder{req: req}
}

func (b *Builder) Prompt(prompt string) *Builder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *Builder) NegativePrompt(neg string) *Builder {
	b.req.Input[0].Params["negative_prompt"] = neg
	return b
}

func (b *Builder) Size(width, height int) *Builder {
	b.req.Input[0].Params["width"] = width
	b.req.Input[0].Params["height"] = height
	return b
}

func (b *Builder) Steps(steps int) *Builder {
	b.req.Input[0].Params["steps"] = steps
	return b
}

func (b *Builder) Seed(seed int64) *Builder {
	b.req.Input[0].Params["seed"] = seed
	return b
}

func (b *Builder) CFGScale(scale float64) *Builder {
	b.req.Input[0].Params["cfg_scale"] = scale
	return b
}

func (b *Builder) N(n int) *Builder {
	b.req.Input[0].Params["n_iter"] = n
	return b
}

func (b *Builder) InputImage(url string) *Builder {
	b.req.Input[0].Content = append(b.req.Input[0].Content, mmtypes.ContentItem{
		Type: "image",
		URL:  url,
	})
	if b.req.Input[0].Params["action"] == nil {
		b.req.Input[0].Params["action"] = 1
	}
	return b
}

func (b *Builder) Action(action int) *Builder {
	b.req.Input[0].Params["action"] = action
	return b
}

func (b *Builder) Denoise(strength float64) *Builder {
	b.req.Input[0].Params["denoise"] = strength
	return b
}

func (b *Builder) ModelVerNo(verNo string) *Builder {
	b.req.Input[0].Params["model_ver_no"] = verNo
	return b
}

func (b *Builder) Param(key string, value any) *Builder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *Builder) Build() *mmtypes.GenerateRequest { return b.req }

func (Factory) SDXL() *Builder        { return newBuilder("sdxl") }
func (Factory) ZImage() *Builder      { return newBuilder("z_image") }
func (Factory) ZImageTurbo() *Builder { return newBuilder("z_image_turbo") }
