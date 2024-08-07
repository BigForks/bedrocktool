package popups

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/bedrock-tool/bedrocktool/ui"
	"github.com/bedrock-tool/bedrocktool/ui/messages"
	"github.com/bedrock-tool/bedrocktool/utils"
)

type ConnectPopup struct {
	ui    ui.UI
	state string
	close widget.Clickable

	ListenIP   string
	ListenPort int

	connectButton widget.Clickable
}

func NewConnect(ui ui.UI, ListenIP string, ListenPort int) Popup {
	if ListenIP == "0.0.0.0" {
		ListenIP = "127.0.0.1"
	}
	return &ConnectPopup{ui: ui, ListenIP: ListenIP, ListenPort: ListenPort}
}

func (p *ConnectPopup) ID() string {
	return "connect"
}

func (p *ConnectPopup) Layout(gtx C, th *material.Theme) D {
	if p.connectButton.Clicked(gtx) {
		go utils.OpenUrl(fmt.Sprintf("minecraft://connect/?serverUrl=%s&serverPort=%d", p.ListenIP, p.ListenPort))
	}

	if p.close.Clicked(gtx) {
		messages.Router.Handle(&messages.Message{
			Source: p.ID(),
			Target: "ui",
			Data:   messages.ExitSubcommand{},
		})
		messages.Router.Handle(&messages.Message{
			Source: p.ID(),
			Target: "ui",
			Data:   messages.Close{Type: "popup", ID: p.ID()},
		})
	}

	return LayoutPopupBackground(gtx, th, "connect", func(gtx C) D {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return layout.Center.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							switch p.state {
							case "listening":
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(material.Label(th, 40, "Listening").Layout),
									layout.Rigid(material.Body1(th, fmt.Sprintf("connect to %s with port %d\nin the minecraft bedrock client to continue", p.ListenIP, p.ListenPort)).Layout),
								)
							case "connecting-server":
								return material.Label(th, 40, "Connecting to Server").Layout(gtx)
							case "established":
								return material.Label(th, 40, "Established").Layout(gtx)
							}
							return D{}
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Max.X /= 2

				return layout.Flex{
					Axis:      layout.Horizontal,
					Spacing:   layout.SpaceBetween,
					Alignment: layout.End,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						b := material.Button(th, &p.close, "Close")
						b.CornerRadius = 8
						return b.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						if p.state == "listening" {
							return layout.Flex{
								Axis:      layout.Horizontal,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx C) D {
									b := material.Button(th, &p.connectButton, "Open Minecraft")
									b.CornerRadius = 8
									return b.Layout(gtx)
								}),
							)
						}
						return D{}
					}),
				)
			}),
		)
	})
}

func (p *ConnectPopup) HandleMessage(msg *messages.Message) *messages.Message {
	switch m := msg.Data.(type) {
	case messages.ConnectStateUpdate:
		switch m.State {
		case messages.ConnectStateListening:
			p.state = "listening"
		case messages.ConnectStateServerConnecting:
			p.state = "connecting-server"
		case messages.ConnectStateEstablished:
			p.state = "established"
		case messages.ConnectStateDone:
			messages.Router.Handle(&messages.Message{
				Source: p.ID(),
				Target: "ui",
				Data:   messages.Close{Type: "popup", ID: p.ID()},
			})
		}
	}
	return nil
}
