package subcommands

import (
	"context"
	"flag"

	"github.com/bedrock-tool/bedrocktool/handlers"
	"github.com/bedrock-tool/bedrocktool/locale"
	"github.com/bedrock-tool/bedrocktool/utils/commands"
	"github.com/bedrock-tool/bedrocktool/utils/proxy"
)

type ChatLogCMD struct {
	ServerAddress string
	Verbose       bool
}

func (*ChatLogCMD) Name() string     { return "chat-log" }
func (*ChatLogCMD) Synopsis() string { return locale.Loc("chat_log_synopsis", nil) }
func (c *ChatLogCMD) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.ServerAddress, "address", "", "remote server address")
	f.BoolVar(&c.Verbose, "v", false, "verbose")
}

func (c *ChatLogCMD) Execute(ctx context.Context) error {
	proxy, err := proxy.New(true)
	if err != nil {
		return err
	}
	proxy.AddHandler(handlers.NewChatLogger())
	return proxy.Run(ctx, c.ServerAddress)
}

func init() {
	commands.RegisterCommand(&ChatLogCMD{})
}
