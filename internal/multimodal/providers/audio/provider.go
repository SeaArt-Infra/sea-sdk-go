package audio

import mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"

type Factory struct{}

func NewFactory() Factory { return Factory{} }

type LyriaBuilder struct {
	req *mmtypes.GenerateRequest
}

func newLyriaBuilder(model string) *LyriaBuilder {
	return &LyriaBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *LyriaBuilder) Prompt(prompt string) *LyriaBuilder {
	b.req.Input[0].Params["input"] = prompt
	return b
}

func (b *LyriaBuilder) Param(key string, value any) *LyriaBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *LyriaBuilder) Build() *mmtypes.GenerateRequest { return b.req }

type MurekaBuilder struct {
	req *mmtypes.GenerateRequest
}

func newMurekaBuilder(model string) *MurekaBuilder {
	return &MurekaBuilder{req: mmtypes.NewGenerateRequest(model)}
}

func (b *MurekaBuilder) Lyrics(lyrics string) *MurekaBuilder {
	b.req.Input[0].Params["lyrics"] = lyrics
	return b
}

func (b *MurekaBuilder) Prompt(prompt string) *MurekaBuilder {
	b.req.Input[0].Params["prompt"] = prompt
	return b
}

func (b *MurekaBuilder) Model(model string) *MurekaBuilder {
	b.req.Input[0].Params["model"] = model
	return b
}

func (b *MurekaBuilder) N(n int) *MurekaBuilder {
	b.req.Input[0].Params["n"] = n
	return b
}

func (b *MurekaBuilder) ReferenceID(id string) *MurekaBuilder {
	b.req.Input[0].Params["reference_id"] = id
	return b
}

func (b *MurekaBuilder) VocalID(id string) *MurekaBuilder {
	b.req.Input[0].Params["vocal_id"] = id
	return b
}

func (b *MurekaBuilder) MelodyID(id string) *MurekaBuilder {
	b.req.Input[0].Params["melody_id"] = id
	return b
}

func (b *MurekaBuilder) Param(key string, value any) *MurekaBuilder {
	b.req.Input[0].Params[key] = value
	return b
}

func (b *MurekaBuilder) Build() *mmtypes.GenerateRequest { return b.req }

func (Factory) Lyria3ProPreview() *LyriaBuilder     { return newLyriaBuilder("lyria_3_pro_preview") }
func (Factory) MurekaSongGenerator() *MurekaBuilder { return newMurekaBuilder("mureka_song_generator") }
