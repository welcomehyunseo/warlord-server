package packet

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/welcomehyunseo/warlord-server/server/data"
)

const OutPacketIDToResponse = 0x00
const OutPacketIDToPong = 0x01

const OutPacketIDToRejectLogin = 0x00
const OutPacketIDToCompleteLogin = 0x02
const OutPacketIDToEnableComp = 0x03

type OutPacketToResponse struct {
	*packet
	max     int    // maximum number of players
	online  int    // current number of players
	text    string // string for description
	favicon string // a png image string that is base64 encoded
}

func NewOutPacketToResponse(
	max, online int,
	text, favicon string,
) *OutPacketToResponse {
	return &OutPacketToResponse{
		newPacket(
			Outbound,
			StatusState,
			OutPacketIDToResponse,
		),
		max, online,
		text, favicon,
	}
}

func (p *OutPacketToResponse) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	jsonString := fmt.Sprintf(
		"{"+
			"\"version\":{\"name\":\"%s\",\"protocol\":%d},"+
			"\"players\":{\"max\":%d,\"online\":%d,\"sample\":[]},"+
			"\"description\":{\"text\":\"%s\"},"+
			"\"favicon\":\"%s\","+
			"\"previewsChat\":%v,"+
			"\"enforcesSecureChat\":%v"+
			"}",
		"1.12.2", 340,
		p.max, p.online,
		p.text, p.favicon,
		true, true,
	)
	if err := dt.WriteString(jsonString); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToResponse) GetMax() int {
	return p.max
}

func (p *OutPacketToResponse) GetOnline() int {
	return p.online
}

func (p *OutPacketToResponse) GetText() string {
	return p.text
}

func (p *OutPacketToResponse) GetFavicon() string {
	return p.favicon
}

func (p *OutPacketToResponse) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, max: %d, online: %d, Text: %s, favicon: %s }",
		p.packet, p.max, p.online, p.text, p.favicon,
	)
}

type OutPacketToPong struct {
	*packet
	payload int64
}

func NewOutPacketToPong(
	payload int64,
) *OutPacketToPong {
	return &OutPacketToPong{
		newPacket(
			Outbound,
			StatusState,
			OutPacketIDToPong,
		),
		payload,
	}
}

func (p *OutPacketToPong) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteInt64(p.payload); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToPong) GetPayload() int64 {
	return p.payload
}

func (p *OutPacketToPong) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, payload: %d }",
		p.packet, p.payload,
	)
}

type OutPacketToCompleteLogin struct {
	*packet
	uid      uuid.UUID
	username string
}

func NewOutPacketToCompleteLogin(
	uid uuid.UUID,
	username string,
) *OutPacketToCompleteLogin {
	return &OutPacketToCompleteLogin{
		packet: newPacket(
			Outbound,
			LoginState,
			OutPacketIDToCompleteLogin,
		),
		uid:      uid,
		username: username,
	}
}

func (p *OutPacketToCompleteLogin) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteString(p.uid.String()); err != nil {
		return nil, err
	}
	if err := dt.WriteString(p.username); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToCompleteLogin) GetUID() uuid.UUID {
	return p.uid
}

func (p *OutPacketToCompleteLogin) GetUsername() string {
	return p.username
}

func (p *OutPacketToCompleteLogin) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, uid: %s, username: %s }",
		p.packet, p.uid, p.username,
	)
}

type OutPacketToEnableComp struct {
	*packet
	threshold int32
}

func NewOutPacketToEnableComp(
	threshold int32,
) *OutPacketToEnableComp {
	return &OutPacketToEnableComp{
		packet: newPacket(
			Outbound,
			LoginState,
			OutPacketIDToEnableComp,
		),
		threshold: threshold,
	}
}

func (p *OutPacketToEnableComp) Pack() (
	[]byte,
	error,
) {
	dt := data.NewData()
	if err := dt.WriteVarInt(p.threshold); err != nil {
		return nil, err
	}

	return dt.GetBytes(), nil
}

func (p *OutPacketToEnableComp) GetThreshold() int32 {
	return p.threshold
}

func (p *OutPacketToEnableComp) String() string {
	return fmt.Sprintf(
		"{ packet: %+v, threshold: %d }",
		p.packet, p.threshold,
	)
}
