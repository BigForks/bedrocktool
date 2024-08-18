package render

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"math"
	"os"
	"path"
	"strings"

	"github.com/bedrock-tool/bedrocktool/subcommands/merge"
	"github.com/bedrock-tool/bedrocktool/utils"
	"github.com/bedrock-tool/bedrocktool/utils/behaviourpack"
	"github.com/bedrock-tool/bedrocktool/utils/commands"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sirupsen/logrus"
)

type RenderCMD struct {
	WorldPath string
	Out       string
	notTall   bool
}

func (*RenderCMD) Name() string     { return "render" }
func (*RenderCMD) Synopsis() string { return "render a world to png" }

func (c *RenderCMD) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.WorldPath, "world", "", "world path")
	f.StringVar(&c.Out, "out", "world.png", "out png path")
	f.BoolVar(&c.notTall, "height255", false, "if the world is a 255 height world, will error if wrong")
}

func (c *RenderCMD) Execute(ctx context.Context) error {
	blockReg := &merge.BlockRegistry{
		BlockRegistry: world.DefaultBlockRegistry,
		Rids:          make(map[uint32]merge.Block),
	}
	entityReg := &merge.EntityRegistry{}

	c.WorldPath = path.Clean(strings.ReplaceAll(c.WorldPath, "\\", "/"))
	c.Out = path.Clean(strings.ReplaceAll(c.Out, "\\", "/"))

	if c.WorldPath == "" {
		return fmt.Errorf("missing -world")
	}

	fmt.Printf("%s\n", c.WorldPath)

	db, err := mcdb.Config{
		Log:      logrus.StandardLogger(),
		Blocks:   blockReg,
		Entities: entityReg,
		ReadOnly: true,
	}.Open(c.WorldPath)
	if err != nil {
		return err
	}
	defer db.Close()

	var resourcePacks []resource.Pack
	resourcePacksFolder := path.Join(c.WorldPath, "resource_packs")
	resourcePackEntries, err := os.ReadDir(resourcePacksFolder)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, entry := range resourcePackEntries {
		pack, err := resource.ReadPath(path.Join(resourcePacksFolder, entry.Name()))
		if err != nil {
			return err
		}
		resourcePacks = append(resourcePacks, pack)
	}

	var behaviorPacks []resource.Pack
	behaviorPacksFolder := path.Join(c.WorldPath, "behavior_packs")
	behaviorPackEntries, err := os.ReadDir(behaviorPacksFolder)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, entry := range behaviorPackEntries {
		pack, err := resource.ReadPath(path.Join(behaviorPacksFolder, entry.Name()))
		if err != nil {
			return err
		}
		behaviorPacks = append(behaviorPacks, pack)
	}

	var entries []protocol.BlockEntry
	for _, pack := range behaviorPacks {
		blockEntries, err := fs.Glob(pack, "blocks/**/*.json")
		if err != nil {
			return err
		}
		for _, bff := range blockEntries {
			f, err := pack.Open(bff)
			if err != nil {
				return err
			}
			var block behaviourpack.MinecraftBlock
			err = json.NewDecoder(f).Decode(&block)
			f.Close()
			if err != nil {
				return err
			}

			ent := protocol.BlockEntry{
				Name: block.Description.Identifier,
				Properties: map[string]any{
					"components": block.Components,
				},
			}
			entries = append(entries, ent)
		}
	}

	var renderer utils.ChunkRenderer
	renderer.ResolveColors(entries, resourcePacks)

	var images = make(map[world.ChunkPos]*image.RGBA)

	boundsMin := world.ChunkPos{math.MaxInt32, math.MaxInt32}
	boundsMax := world.ChunkPos{math.MinInt32, math.MinInt32}

	it := db.NewColumnIterator(nil, c.notTall)
	defer it.Release()
	for it.Next() {
		col := it.Column()
		pos := it.Position()

		if err := it.Error(); err != nil {
			return err
		}

		boundsMin[0] = min(boundsMin[0], pos[0])
		boundsMin[1] = min(boundsMin[1], pos[1])
		boundsMax[0] = max(boundsMax[0], pos[0])
		boundsMax[1] = max(boundsMax[1], pos[1])

		images[pos] = renderer.Chunk2Img(col.Chunk)
	}
	if err := it.Error(); err != nil {
		return err
	}

	chunksX := int(boundsMax[0] - boundsMin[0] + 1)
	chunksY := int(boundsMax[1] - boundsMin[1] + 1)
	r := image.Rect(0, 0, chunksX*16, chunksY*16)

	fmt.Printf("%dx%d pixels\n", r.Dx(), r.Dy())

	img := image.NewRGBA(r)

	for pos, tile := range images {
		px := image.Pt(
			int((pos.X()-boundsMin.X())*16),
			int((pos.Z()-boundsMin.Z())*16),
		)
		draw.Draw(img, image.Rect(
			px.X, px.Y,
			px.X+16, px.Y+16,
		), tile, image.Point{}, draw.Src)
	}

	f, err := os.Create(c.Out)
	if err != nil {
		return err
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		return err
	}

	logrus.Infof("Wrote %s", c.Out)

	return nil
}

func init() {
	commands.RegisterCommand(&RenderCMD{})
}
