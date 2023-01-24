package server

import (
	"encoding/json"
	"github.com/google/uuid"
)

const (
	ResponsePacketID = 0x00
	PongPacketID     = 0x01

	DisconnectLoginPacketID = 0x00
	CompleteLoginPacketID   = 0x02
)

type OutPacket interface {
	Write() *Data

	GetBoundTo() int
	GetState() int
	GetID() int32
}

type Version struct {
	Name     string `json:"name"`
	Protocol int32  `json:"protocol"`
}

type Sample struct {
	Name string    `json:"name"`
	Id   uuid.UUID `json:"playerID"`
}

type Players struct {
	Max    int       `json:"max"`
	Online int       `json:"online"`
	Sample []*Sample `json:"sample"`
}

type Description struct {
	Text string `json:"text"`
}

type JsonResponse struct {
	Version            *Version     `json:"version"`
	Players            *Players     `json:"players"`
	Description        *Description `json:"description"`
	Favicon            string       `json:"favicon"`
	PreviewsChat       bool         `json:"previewsChat"`
	EnforcesSecureChat bool         `json:"enforcesSecureChat"`
}

type ResponsePacket struct {
	*packet
	jsonResponse *JsonResponse
}

func NewResponsePacket(
	jsonResponse *JsonResponse,
) *ResponsePacket {
	return &ResponsePacket{
		packet: newPacket(
			Outbound,
			StatusState,
			ResponsePacketID,
		),
		jsonResponse: jsonResponse,
	}
}

func (p *ResponsePacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	buf, _ := json.Marshal(p.jsonResponse)
	d0.WriteString(string(buf))

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *ResponsePacket) GetJsonResponse() *JsonResponse {
	return p.jsonResponse
}

type PongPacket struct {
	*packet
	payload int64
}

func NewPongPacket(
	payload int64,
) *PongPacket {
	return &PongPacket{
		packet: newPacket(
			Outbound,
			StatusState,
			PongPacketID,
		),
		payload: payload,
	}
}

func (p *PongPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WriteInt64(p.payload)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *PongPacket) GetPayload() int64 {
	return p.payload
}

type CompleteLoginPacket struct {
	*packet
	playerID uuid.UUID
	username string
}

func NewCompleteLoginPacket(
	playerID uuid.UUID,
	username string,
) *CompleteLoginPacket {
	return &CompleteLoginPacket{
		packet: newPacket(
			Outbound,
			LoginState,
			CompleteLoginPacketID,
		),
		playerID: playerID,
		username: username,
	}
}

func (p *CompleteLoginPacket) Write() *Data {
	d0 := NewData()
	d0.WriteVarInt(p.GetID())
	d0.WriteString(p.playerID.String())
	d0.WriteString(p.username)

	length := d0.GetLength()
	d1 := NewData()
	d1.WriteVarInt(int32(length))
	d1.Write(d0)
	return d1
}

func (p *CompleteLoginPacket) GetPlayerID() uuid.UUID {
	return p.playerID
}

func (p *CompleteLoginPacket) GetUsername() string {
	return p.username
}
