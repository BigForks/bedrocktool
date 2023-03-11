package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/bedrock-tool/bedrocktool/locale"
	"github.com/fatih/color"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

var pool = packet.NewPool()

var MutedPackets = []string{
	"packet.UpdateBlock",
	"packet.MoveActorAbsolute",
	"packet.SetActorMotion",
	"packet.SetTime",
	"packet.RemoveActor",
	"packet.AddActor",
	"packet.UpdateAttributes",
	"packet.Interact",
	"packet.LevelEvent",
	"packet.SetActorData",
	"packet.MoveActorDelta",
	"packet.MovePlayer",
	"packet.BlockActorData",
	"packet.PlayerAuthInput",
	"packet.LevelChunk",
	"packet.LevelSoundEvent",
	"packet.ActorEvent",
	"packet.NetworkChunkPublisherUpdate",
	"packet.UpdateSubChunkBlocks",
	"packet.SubChunk",
	"packet.SubChunkRequest",
	"packet.Animate",
	"packet.NetworkStackLatency",
	"packet.InventoryTransaction",
	"packet.PlaySound",
}

var (
	ExtraVerbose []string
	FLog         io.Writer
	dmpLock      sync.Mutex
)

func dmpStruct(level int, inputStruct any, withType bool, isInList bool) (s string) {
	tBase := strings.Repeat("\t", level)

	ii := reflect.Indirect(reflect.ValueOf(inputStruct))
	typeName := reflect.TypeOf(inputStruct).String()
	typeString := ""
	if withType {
		if slices.Contains([]string{"bool", "string"}, typeName) {
		} else {
			typeString = typeName + " "
		}
	}

	if ii.Kind() == reflect.Struct {
		if ii.NumField() == 0 {
			s += typeName + "{}"
		} else {
			s += typeName + "{\n"
			for i := 0; i < ii.NumField(); i++ {
				fieldType := ii.Type().Field(i)

				if fieldType.IsExported() {
					field := ii.Field(i).Interface()
					d := dmpStruct(level+1, field, true, false)
					s += tBase + fmt.Sprintf("\t%s = %s\n", fieldType.Name, d)
				} else {
					s += tBase + " " + fieldType.Name + " (unexported)"
				}
			}
			s += tBase + "}"
		}
	} else if ii.Kind() == reflect.Slice {
		var t reflect.Type
		is_elem_struct := false
		if ii.Len() > 0 {
			e := ii.Index(0)
			t = reflect.TypeOf(e.Interface())
			is_elem_struct = t.Kind() == reflect.Struct
		}

		if ii.Len() > 1000 {
			s += "[<slice too long>]"
		} else if ii.Len() == 0 {
			s += typeString + "[]"
		} else {
			s += typeString + "["
			if is_elem_struct {
				s += "\n"
			}
			for i := 0; i < ii.Len(); i++ {
				if is_elem_struct {
					s += tBase + "\t"
				}
				s += dmpStruct(level+1, ii.Index(i).Interface(), false, true)
				if is_elem_struct {
					s += "\n"
				} else {
					s += " "
				}
			}
			if is_elem_struct {
				s += tBase
			}
			s += "]"
		}
	} else if ii.Kind() == reflect.Map {
		j, err := json.MarshalIndent(ii.Interface(), tBase, "\t")
		if err != nil {
			s += err.Error()
		}
		s += string(j)
	} else {
		if !isInList {
			s += typeString
		}
		s += fmt.Sprintf("%#v", ii.Interface())
	}

	return s
}

func PacketLogger(header packet.Header, payload []byte, src, dst net.Addr) {
	var pk packet.Packet
	if pkFunc, ok := pool[header.PacketID]; ok {
		pk = pkFunc()
	} else {
		pk = &packet.Unknown{PacketID: header.PacketID}
	}

	if pk.ID() == packet.IDRequestNetworkSettings {
		ClientAddr = src
	}

	defer func() {
		if recoveredErr := recover(); recoveredErr != nil {
			logrus.Errorf("%T: %w", pk, recoveredErr)
		}
	}()

	pk.Unmarshal(protocol.NewReader(bytes.NewBuffer(payload), 0))

	if FLog != nil {
		dmpLock.Lock()
		defer dmpLock.Unlock()
		FLog.Write([]byte(dmpStruct(0, pk, true, false)))
		FLog.Write([]byte("\n\n\n"))
	}

	pkName := reflect.TypeOf(pk).String()[1:]
	if slices.Contains(MutedPackets, pkName) {
		return
	}

	switch pk := pk.(type) {
	case *packet.Disconnect:
		logrus.Infof(locale.Loc("disconnect", locale.Strmap{"Pk": pk}))
	}

	dirS2C := color.GreenString("S") + "->" + color.CyanString("C")
	dirC2S := color.CyanString("C") + "->" + color.GreenString("S")
	var dir string = dirS2C

	if ClientAddr != nil {
		if src == ClientAddr {
			dir = dirC2S
		}
	} else {
		srcAddr, _, _ := net.SplitHostPort(src.String())
		if IPPrivate(net.ParseIP(srcAddr)) {
			dir = dirS2C
		}
	}

	logrus.Debugf("%s 0x%02x, %s", dir, pk.ID(), pkName)

	if slices.Contains(ExtraVerbose, pkName) {
		logrus.Debugf("%+v", pk)
	}
}
