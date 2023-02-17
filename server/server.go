package server

import (
	"errors"
	"github.com/google/uuid"
	"net"
	"time"
)

const Network = "tcp"   // network type of server
const McName = "1.12.2" // minecraft version name
const ProtVer = 340     // protocol version

const CompThold = 16 // threshold for compression

const MinRndDist = 2  // minimum render distance
const MaxRndDist = 32 // maximum render distance

const CheckKeepAliveTime = time.Millisecond * 1000
const Loop3Time = time.Millisecond * 1

var DifferentKeepAlivePayloadError = errors.New("the payload of keep-alive must be same as the given")
var OutOfRndDistRangeError = errors.New("it is out of maximum and minimum value of render distance")

type ChanForError chan any

type Server struct {
	addr string // address

	max    int   // maximum number of players
	online int   // number of online players
	last   int32 // last entity ID

	favicon string // base64 png image string
	text    string // description of server

	world *Overworld
}

func NewServer(
	addr string,
	max int,
	favicon, text string,
	world *Overworld,
) *Server {

	return &Server{
		addr: addr,

		max:    max,
		online: 0,
		last:   0,

		favicon: favicon,
		text:    text,

		world: world,
	}
}

func (s *Server) countEID() int32 {
	eid := s.last
	s.last++
	return eid
}

func (s *Server) handleSpawnPlayerEvent(
	chanForEvent ChanForSpawnPlayerEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"spawn-player-event-handler",
		NewLgElement("cnt", cnt),
	)
	lg.Debug("it is started to handle SpawnPlayerEvent")
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}

		lg.Debug("it is finished to handle SpawnPlayerEvent")
	}()

	var stop bool
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"it is started to process event",
				NewLgElement("event", event),
			)

			eid, uid :=
				event.GetEID(), event.GetUUID()
			x, y, z :=
				event.GetX(), event.GetY(), event.GetZ()
			yaw, pitch :=
				event.GetYaw(), event.GetPitch()
			packet := NewSpawnPlayerPacket(
				eid, uid,
				x, y, z,
				yaw, pitch,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}

			lg.Debug(
				"it is finished to process event",
			)
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleDespawnEntityEvent(
	chanForEvent ChanForDespawnEntityEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"despawn-entity-event-handler",
		NewLgElement("cnt", cnt),
	)
	lg.Debug("it is started to handle DespawnEntityEvent")
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}

		lg.Debug("it is finished to handle DespawnEntityEvent")
	}()

	var stop bool
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"it is started to process event",
				NewLgElement("event", event),
			)

			eid := event.GetEID()
			packet := NewDespawnEntityPacket(
				eid,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}

			lg.Debug(
				"it is finished to process event",
			)
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleLoadChunkEvent(
	chanForEvent ChanForLoadChunkEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"load-chunk-event-handler",
		NewLgElement("cnt", cnt),
	)
	lg.Debug("it is started to handle LoadChunkEvent")
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}

		lg.Debug("it is finished to handle LoadChunkEvent")
	}()

	var stop bool
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"it is started to process event",
				NewLgElement("event", event),
			)

			overworld, init := event.IsOverworld(), event.IsInit()
			cx, cz := event.GetCX(), event.GetCZ()
			chunk := event.GetChunk()
			bitmask, data := chunk.GenerateData(init, overworld)
			packet := NewSendChunkDataPacket(
				cx, cz,
				init,
				bitmask,
				data,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}

			lg.Debug(
				"it is finished to process event",
			)
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleUnloadChunkEvent(
	chanForEvent ChanForUnloadChunkEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"upload-chunk-event-handler",
		NewLgElement("cnt", cnt),
	)
	lg.Debug("it is started to handle UnloadChunkEvent")
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}

		lg.Debug("it is finished to handle UnloadChunkEvent")
	}()

	var stop bool
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"it is started to process event",
				NewLgElement("event", event),
			)

			cx, cz := event.GetCX(), event.GetCZ()
			packet := NewUnloadChunkPacket(
				cx, cz,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}

			lg.Debug(
				"it is finished to process event",
			)
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleSetEntityLookEvent(
	chanForEvent ChanForSetEntityLookEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"set-entity-look-event-handler",
		NewLgElement("cnt", cnt),
	)
	lg.Debug("it is started to handle SetEntityLookEvent")
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}

		lg.Debug("it is finished to handle SetEntityLookEvent")
	}()

	var stop bool
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"it is started to process event",
				NewLgElement("event", event),
			)

			eid := event.GetEID()
			yaw, pitch := event.GetYaw(), event.GetPitch()
			ground := event.GetGround()
			packet := NewSetEntityLookPacket(
				eid,
				yaw, pitch,
				ground,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}

			lg.Debug(
				"it is finished to process event",
			)
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleSetEntityRelativePosEvent(
	chanForEvent ChanForSetEntityRelativePosEvent,
	cnt *Client,
	chanForError ChanForError,
) {
	lg := NewLogger(
		"set-entity-relative-pos-event-handler",
		NewLgElement("cnt", cnt),
	)
	lg.Debug("it is started to handle SetEntityRelativePosEvent")
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}

		lg.Debug("it is finished to handle SetEntityRelativePosEvent")
	}()

	var stop bool
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			lg.Debug(
				"it is started to process event",
				NewLgElement("event", event),
			)

			eid := event.GetEID()
			deltaX, deltaY, deltaZ :=
				event.GetDeltaX(), event.GetDeltaY(), event.GetDeltaZ()
			ground := event.GetGround()
			packet := NewSetEntityRelativePosPacket(
				eid,
				deltaX, deltaY, deltaZ,
				ground,
			)
			if err := cnt.WriteWithComp(packet); err != nil {
				panic(err)
			}

			lg.Debug(
				"it is finished to process event",
			)
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleClient(
	cnt *Client,
) {
	cnt.Init()
	defer cnt.Close()

	addr := cnt.GetAddr()
	lg := NewLogger(
		"client-handler",
		NewLgElement("addr", addr),
	)
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
		}
	}()
	lg.Debug("it is started to handle Client")
	defer func() {
		lg.Debug("it is finished to handle Client")
	}()

	max, online :=
		s.max, s.online
	text, favicon :=
		s.text, s.favicon
	stop, err :=
		cnt.HandleNonLoginState(
			lg,
			max, online,
			text, favicon,
		)
	if err != nil {
		panic(err)
	}
	if stop == true {
		return
	}

	uid, username, err :=
		cnt.HandleLoginState(lg)
	if err != nil {
		panic(err)
	}

	lg.Debug(
		"client has successfully logged in",
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)

	eid := s.countEID()
	world := s.world
	spawnX, spawnY, spawnZ :=
		world.GetSpawnX(), world.GetSpawnY(), world.GetSpawnZ()
	spawnYaw, spawnPitch :=
		world.GetSpawnYaw(), world.GetSpawnPitch()
	if err := cnt.JoinGame(
		eid,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		panic(err)
	}

	chanForError := make(ChanForError, 1)

	// TODO: set chan size in best practices
	chanForSpawnPlayerEvent := make(ChanForSpawnPlayerEvent, 4)
	go s.handleSpawnPlayerEvent(
		chanForSpawnPlayerEvent,
		cnt,
		chanForError,
	)

	chanForDespawnEntityEvent := make(ChanForDespawnEntityEvent, 4)
	go s.handleDespawnEntityEvent(
		chanForDespawnEntityEvent,
		cnt,
		chanForError,
	)

	chanForLoadChunkEvent := make(ChanForLoadChunkEvent, 4)
	go s.handleLoadChunkEvent(
		chanForLoadChunkEvent,
		cnt,
		chanForError,
	)

	chanForUnloadChunkEvent := make(ChanForUnloadChunkEvent, 4)
	go s.handleUnloadChunkEvent(
		chanForUnloadChunkEvent,
		cnt,
		chanForError,
	)

	chanForSetEntityLookEvent :=
		make(ChanForSetEntityLookEvent, 4)
	go s.handleSetEntityLookEvent(
		chanForSetEntityLookEvent,
		cnt,
		chanForError,
	)

	chanForSetEntityRelativePosEvent :=
		make(ChanForSetEntityRelativePosEvent, 4)
	go s.handleSetEntityRelativePosEvent(
		chanForSetEntityRelativePosEvent,
		cnt,
		chanForError,
	)

	if err := world.InitPlayer(
		eid,
		uid, username,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityRelativePosEvent,
	); err != nil {
		panic(err)
	}
	defer func() {
		world.ClosePlayer(eid)

		close(chanForSpawnPlayerEvent)
		close(chanForDespawnEntityEvent)
		close(chanForLoadChunkEvent)
		close(chanForUnloadChunkEvent)
		close(chanForSetEntityLookEvent)
		close(chanForSetEntityRelativePosEvent)

		close(chanForError)
	}()

	go cnt.HandlePlayState(
		eid,
		world,
		chanForError,
	)

	for {
		var stop bool

		select {
		case <-chanForError:
			stop = true
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) Render() {
	lg := NewLogger(
		"server-renderer",
	)
	defer lg.Close()
	addr := s.addr
	network := Network

	lg.Info(
		"It is started to render.",
		NewLgElement("addr", addr),
		NewLgElement("network", network),
	)

	ln, err := net.Listen(network, addr)

	if err != nil {
		panic(err)
	}
	defer func() {
		lg.Info(
			"It is finished to render.",
		)
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		lg.Info(
			"The server accepted a new connection.",
			NewLgElement("addr", conn.RemoteAddr()),
		)

		cid, err := uuid.NewRandom()
		if err != nil {
			panic(err)
		}

		cnt := NewClient(
			cid,
			conn,
		)

		go s.handleClient(cnt)
	}

}
