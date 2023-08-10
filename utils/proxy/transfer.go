package proxy

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type transferingErr struct {
	To string
}

func (transferingErr) Error() string {
	return "transferingErr"
}

type transferHandler struct {
	p *Context
}

func NewTransferHandler() *Handler {
	t := &transferHandler{}
	return &Handler{
		Name: "transfer",
		ProxyRef: func(pc *Context) {
			t.p = pc
		},
		PacketCB: t.packetCB,
	}
}

func (t *transferHandler) packetCB(pk packet.Packet, toServer bool, timeReceived time.Time, preLogin bool) (packet.Packet, error) {
	switch pk := pk.(type) {
	case *packet.Transfer:
		var pk2 packet.Packet = nil
		if t.p.Client != nil {
			host, port, err := net.SplitHostPort(t.p.Client.ClientData().ServerAddress)
			if err != nil {
				return nil, err
			}
			_port, _ := strconv.Atoi(port)
			pk2 = &packet.Transfer{Address: host, Port: uint16(_port)}
		}

		return pk2, &transferingErr{
			To: fmt.Sprintf("%s:%d", pk.Address, pk.Port),
		}
	}
	return pk, nil
}
