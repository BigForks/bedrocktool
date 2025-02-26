package behaviourpack

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type BlockBehaviour struct {
	FormatVersion  string         `json:"format_version"`
	MinecraftBlock MinecraftBlock `json:"minecraft:block"`
}

func (bp *Pack) AddBlock(block protocol.BlockEntry) {
	ns, _ := splitNamespace(block.Name)
	if ns == "minecraft" {
		return
	}

	minecraftBlock, version := parseBlock(block)

	bp.blocks[block.Name] = &BlockBehaviour{
		FormatVersion:  version,
		MinecraftBlock: minecraftBlock,
	}
}
