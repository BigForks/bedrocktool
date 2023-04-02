package handlers

import (
	"fmt"
	"os"
	"time"

	"github.com/bedrock-tool/bedrocktool/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sirupsen/logrus"
)

type ChatLogger struct {
	Verbose bool
	fio     *os.File
}

func (c *ChatLogger) AddressAndName(address, hostname string) error {
	filename := fmt.Sprintf("%s_%s_chat.log", hostname, time.Now().Format("2006-01-02_15-04-05_Z07"))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	c.fio = f
	return nil
}

func (c *ChatLogger) PacketCB(pk packet.Packet, toServer bool, t time.Time) (packet.Packet, error) {
	if text, ok := pk.(*packet.Text); ok {
		logLine := text.Message
		if c.Verbose {
			logLine += fmt.Sprintf("   (TextType: %d | XUID: %s | PlatformChatID: %s)", text.TextType, text.XUID, text.PlatformChatID)
		}
		c.fio.WriteString(fmt.Sprintf("[%s] ", t.Format(time.RFC3339)))
		logrus.Info(logLine)
		if toServer {
			c.fio.WriteString("SENT: ")
		}
		c.fio.WriteString(logLine + "\n")
	}
	return pk, nil
}

func NewChatLogger() *utils.ProxyHandler {
	p := &ChatLogger{}
	return &utils.ProxyHandler{
		Name:           "Packet Capturer",
		PacketCB:       p.PacketCB,
		AddressAndName: p.AddressAndName,
		OnEnd: func() {
			p.fio.Close()
		},
	}
}
