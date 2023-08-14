package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/bedrock-tool/bedrocktool/utils/proxy"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (p *packetCapturer) dumpPacket(toServer bool, payload []byte) {
	p.dumpLock.Lock()
	defer p.dumpLock.Unlock()
	p.fio.Write([]byte{0xAA, 0xAA, 0xAA, 0xAA})
	packetSize := uint32(len(payload))
	binary.Write(p.fio, binary.LittleEndian, packetSize)
	binary.Write(p.fio, binary.LittleEndian, toServer)
	binary.Write(p.fio, binary.LittleEndian, time.Now().UnixMilli())
	p.fio.Write(payload)
	p.fio.Write([]byte{0xBB, 0xBB, 0xBB, 0xBB})
}

type packetCapturer struct {
	proxy    *proxy.Context
	fio      *os.File
	dumpLock sync.Mutex
}

func (p *packetCapturer) AddressAndName(address, hostname string) error {
	os.Mkdir("captures", 0o775)
	fio, err := os.Create(fmt.Sprintf("captures/%s-%s.pcap2", hostname, time.Now().Format("2006-01-02_15-04-05")))
	if err != nil {
		return err
	}
	proxy.WriteReplayHeader(fio)
	p.fio = fio
	return nil
}

func (p *packetCapturer) PacketFunc(header packet.Header, payload []byte, src, dst net.Addr) {
	buf := bytes.NewBuffer(nil)
	header.Write(buf)
	buf.Write(payload)
	p.dumpPacket(p.proxy.IsClient(src), buf.Bytes())
}

func NewPacketCapturer() *proxy.Handler {
	p := &packetCapturer{}
	return &proxy.Handler{
		Name: "Packet Capturer",
		ProxyRef: func(pc *proxy.Context) {
			p.proxy = pc
		},
		AddressAndName: p.AddressAndName,
		PacketRaw:      p.PacketFunc,
		OnEnd: func() {
			p.dumpLock.Lock()
			defer p.dumpLock.Unlock()
			p.fio.Close()
		},
	}
}

func init() {
	proxy.NewPacketCapturer = NewPacketCapturer
}