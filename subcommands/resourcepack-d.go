//go:build packs

package subcommands

import (
	"context"
	"flag"

	"github.com/bedrock-tool/bedrocktool/locale"
	resourcepackd "github.com/bedrock-tool/bedrocktool/subcommands/resourcepack-d"
	"github.com/bedrock-tool/bedrocktool/utils/commands"
)

type ResourcePackCMD struct {
	ServerAddress string
	SaveEncrypted bool
	OnlyKeys      bool
	writeFolders  bool
	f             *flag.FlagSet
}

func (*ResourcePackCMD) Name() string     { return "packs" }
func (*ResourcePackCMD) Synopsis() string { return locale.Loc("pack_synopsis", nil) }

func (c *ResourcePackCMD) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.ServerAddress, "address", "", locale.Loc("remote_address", nil))
	f.BoolVar(&c.SaveEncrypted, "save-encrypted", false, locale.Loc("save_encrypted", nil))
	f.BoolVar(&c.OnlyKeys, "only-keys", false, locale.Loc("only_keys", nil))
	f.BoolVar(&c.writeFolders, "folders", false, "folders instead of zips")
	c.f = f
}

func (c *ResourcePackCMD) Execute(ctx context.Context) error {
	return resourcepackd.Execute_cmd(ctx, c.ServerAddress, c.OnlyKeys, c.SaveEncrypted, c.writeFolders, c.f)
}

func init() {
	commands.RegisterCommand(&ResourcePackCMD{})
}
