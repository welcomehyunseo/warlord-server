package server

import (
	"errors"
	"github.com/google/uuid"
	"math/rand"
	"net"
	"sync"
	"time"
)

const Network = "tcp"   // network type of server
const McName = "1.12.2" // minecraft version name
const ProtVer = 340     // protocol version

const CompThold = 16 // threshold for compression

const MinRndDist = 2  // minimum render distance
const MaxRndDist = 32 // maximum render distance

const DelayForCheckKeepAlive = time.Millisecond * 1000
const LoopDelayForPlayState = time.Millisecond * 1

const MaxNumForChannel = 16

type ChanForError chan any

type Server struct {
	sync.RWMutex

	addr string // address

	max    int // maximum number of players
	online int // number of online players
	last   EID // last entity ID

	favicon string // base64 png image string
	text    string // description of server

}

func NewServer(
	addr string,
	max int,
	favicon, text string,
) *Server {

	return &Server{
		addr:   addr,
		max:    max,
		online: 0,
		last:   0,

		favicon: favicon,
		text:    text,
	}
}

func (s *Server) countEID() EID {
	a := s.last
	s.last++
	return a
}

func (s *Server) handleSpawnPlayerEvent(
	chanForEvent ChanForSpawnPlayerEvent,
	cnt *Client,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"spawn-player-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SpawnPlayerEvent")
	defer func() {
		lg.Debug("it is finished to handle SpawnPlayerEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid, uid :=
					event.GetEID(), event.GetUUID()
				x, y, z :=
					event.GetX(), event.GetY(), event.GetZ()
				yaw, pitch :=
					event.GetYaw(), event.GetPitch()
				sneaking, sprinting :=
					event.IsSneaking(), event.IsSprinting()
				if err := cnt.SpawnPlayer(
					lg,
					eid, uid,
					x, y, z,
					yaw, pitch,
					sneaking, sprinting,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
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
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"despawn-entity-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle DespawnEntityEvent")
	defer func() {
		lg.Debug("it is finished to handle DespawnEntityEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				if err := cnt.DespawnEntity(
					lg,
					eid,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
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
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"set-entity-relative-pos-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SetEntityRelativePosEvent")
	defer func() {
		lg.Debug("it is finished to handle SetEntityRelativePosEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				deltaX, deltaY, deltaZ :=
					event.GetDeltaX(),
					event.GetDeltaY(),
					event.GetDeltaZ()
				ground := event.GetGround()
				if err := cnt.SetEntityRelativePos(
					lg,
					eid,
					deltaX, deltaY, deltaZ,
					ground,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
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
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"set-entity-look-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SetEntityLookEvent")
	defer func() {
		lg.Debug("it is finished to handle SetEntityLookEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				yaw, pitch :=
					event.GetYaw(), event.GetPitch()
				ground := event.GetGround()
				if err := cnt.SetEntityLook(
					lg,
					eid,
					yaw, pitch,
					ground,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (s *Server) handleSetEntityMetadataEvent(
	chanForEvent ChanForSetEntityMetadataEvent,
	cnt *Client,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"set-entity-look-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle SetEntityMetadataEvent")
	defer func() {
		lg.Debug("it is finished to handle SetEntityMetadataEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				eid := event.GetEID()
				metadata := event.GetMetadata()
				if err := cnt.SetEntityMetadata(
					lg,
					eid,
					metadata,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (s *Server) handleAddPlayerEvent(
	chanForEvent ChanForAddPlayerEvent,
	cnt *Client,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"add-player-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle AddPlayerEvent")
	defer func() {
		lg.Debug("it is finished to handle AddPlayerEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				uid, username :=
					event.GetUUID(), event.GetUsername()
				if err := cnt.AddPlayer(
					lg,
					uid, username,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				event.Fail()
				panic(err)
			}

			event.Done()
		}

		if stop == true {
			break
		}
	}
}

func (s *Server) handleUpdateLatencyEvent(
	chanForEvent ChanForUpdateLatencyEvent,
	cnt *Client,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"update-latency-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle UpdateLatencyEvent")
	defer func() {
		lg.Debug("it is finished to handle UpdateLatencyEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}

			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				uid, latency := event.GetUUID(), event.GetLatency()
				if err := cnt.UpdateLatency(
					lg,
					uid, latency,
				); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) handleRemovePlayerEvent(
	chanForEvent ChanForRemovePlayerEvent,
	cnt *Client,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"remove-player-event-handler",
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle RemovePlayerEvent")
	defer func() {
		lg.Debug("it is finished to handle RemovePlayerEvent")
	}()

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	stop := false
	for {
		select {
		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				uid := event.GetUUID()
				if err := cnt.RemovePlayer(lg, uid); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				panic(err)
			}
		}

		if stop == true {
			break
		}
	}
}

func (s *Server) handleConfirmKeepAliveEvent(
	playerList *PlayerList,
	chanForEvent ChanForConfirmKeepAliveEvent,
	uid UID,
	cnt *Client,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	lg := NewLogger(
		"confirm-keep-alive-event-handler",
		NewLgElement("uid", uid),
		NewLgElement("client", cnt),
	)
	defer lg.Close()

	lg.Debug("it is started to handle ConfirmKeepAliveEvent")
	defer func() {
		lg.Debug("it is finished to handle ConfirmKeepAliveEvent")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
			chanForError <- err
		}
	}()

	start := time.Time{}
	var payload0 int64

	stop := false
	for {
		select {
		case <-time.After(DelayForCheckKeepAlive):
			if start.IsZero() == false {
				break
			}
			payload0 = rand.Int63()
			if err := cnt.CheckKeepAlive(lg, payload0); err != nil {
				panic(err)
			}
			start = time.Now()

		case event, ok := <-chanForEvent:
			if ok == false {
				stop = true
				break
			}
			if err := func() error {
				lg.Debug(
					"it is started to process event",
					NewLgElement("event", event),
				)
				defer func() {
					lg.Debug("it is finished to process event")
				}()

				if err := func() error {
					payload1 := event.GetPayload()
					if payload1 != payload0 {
						return errors.New("payload for keep-alive must be same as given")
					}
					end := time.Now()
					latency := int32(end.Sub(start).Milliseconds())

					if err := playerList.UpdateLatency(
						lg,
						uid,
						latency,
						cnt,
					); err != nil {
						return err
					}

					return nil
				}(); err != nil {
					return err
				}

				start = time.Time{}

				return nil
			}(); err != nil {
				panic(err)
			}
			break
		}

		if stop == true {
			break
		}
	}

}

func (s *Server) initClient(
	lg *Logger,
	playerList *PlayerList,
	world *Overworld,
	cnt *Client,
	uid UID, username string,
	wg *sync.WaitGroup,
) (
	EID,
	ChanForConfirmKeepAliveEvent,
	ChanForError,
	error,
) {
	s.Lock()
	defer s.Unlock()

	lg.Debug("it is started to init Connection")
	defer func() {
		lg.Debug("it is finished to init Connection")
	}()

	eid := s.countEID()

	if err := cnt.JoinGame(
		lg,
		eid,
	); err != nil {
		return 0, nil, nil, err
	}

	chanForError := make(
		ChanForError,
		MaxNumForChannel,
	)

	chanForAddPlayerEvent := make(
		ChanForAddPlayerEvent,
		MaxNumForChannel,
	)
	go s.handleAddPlayerEvent(
		chanForAddPlayerEvent,
		cnt,
		chanForError,
		wg,
	)

	chanForUpdateLatencyEvent := make(
		ChanForUpdateLatencyEvent,
		MaxNumForChannel,
	)
	go s.handleUpdateLatencyEvent(
		chanForUpdateLatencyEvent,
		cnt,
		chanForError,
		wg,
	)

	chanForRemovePlayerEvent := make(
		ChanForRemovePlayerEvent,
		MaxNumForChannel,
	)
	go s.handleRemovePlayerEvent(
		chanForRemovePlayerEvent,
		cnt,
		chanForError,
		wg,
	)

	if err := playerList.InitPlayer(
		lg,
		uid, username,
		cnt,
		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
	); err != nil {
		return 0, nil, nil, err
	}

	chanForConfirmKeepAliveEvent := make(
		ChanForConfirmKeepAliveEvent,
		MaxNumForChannel,
	)
	go s.handleConfirmKeepAliveEvent(
		playerList,
		chanForConfirmKeepAliveEvent,
		uid,
		cnt,
		chanForError,
		wg,
	)

	chanForSpawnPlayerEvent := make(
		ChanForSpawnPlayerEvent,
		MaxNumForChannel,
	)
	go s.handleSpawnPlayerEvent(
		chanForSpawnPlayerEvent,
		cnt,
		chanForError,
		wg,
	)

	chanForDespawnEntityEvent := make(
		ChanForDespawnEntityEvent,
		MaxNumForChannel,
	)
	go s.handleDespawnEntityEvent(
		chanForDespawnEntityEvent,
		cnt,
		chanForError,
		wg,
	)

	chanForSetEntityRelativePosEvent := make(
		ChanForSetEntityRelativePosEvent,
		MaxNumForChannel,
	)
	go s.handleSetEntityRelativePosEvent(
		chanForSetEntityRelativePosEvent,
		cnt,
		chanForError,
		wg,
	)

	chanForSetEntityLookEvent := make(
		ChanForSetEntityLookEvent,
		MaxNumForChannel,
	)
	go s.handleSetEntityLookEvent(
		chanForSetEntityLookEvent,
		cnt,
		chanForError,
		wg,
	)

	chanForSetEntityMetadataEvent := make(
		ChanForSetEntityMetadataEvent,
		MaxNumForChannel,
	)
	go s.handleSetEntityMetadataEvent(
		chanForSetEntityMetadataEvent,
		cnt,
		chanForError,
		wg,
	)

	if err := world.InitPlayer(
		lg,
		eid, uid, username,
		cnt,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent,
	); err != nil {
		return 0, nil, nil, err
	}

	return eid,
		chanForConfirmKeepAliveEvent,
		chanForError,
		nil
}

func (s *Server) closeClient(
	lg *Logger,
	playerList *PlayerList,
	world *Overworld,
	eid EID, uid UID,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	s.Lock()
	defer s.Unlock()

	lg.Debug("it is started to close Connection")
	defer func() {
		lg.Debug("it is finished to close Connection")
	}()

	chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityMetadataEvent :=
		world.ClosePlayer(
			lg,
			eid,
		)
	close(chanForSpawnPlayerEvent)
	close(chanForDespawnEntityEvent)
	close(chanForSetEntityRelativePosEvent)
	close(chanForSetEntityLookEvent)
	close(chanForSetEntityMetadataEvent)

	close(chanForConfirmKeepAliveEvent)

	chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent :=
		playerList.ClosePlayer(
			lg,
			uid,
		)
	close(chanForAddPlayerEvent)
	close(chanForUpdateLatencyEvent)
	close(chanForRemovePlayerEvent)

	wg.Wait()
	close(chanForError)
}

func (s *Server) handleClient(
	playerList *PlayerList,
	world *Overworld,
	cnt *Client,
) {

	addr := cnt.GetAddr()
	lg := NewLogger(
		"client-handler",
		NewLgElement("addr", addr),
	)
	defer lg.Close()

	lg.Debug("it is started to handle Client")
	defer func() {
		lg.Debug("it is finished to handle Client")
	}()

	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
		}
	}()

	cnt.Init(lg)
	defer cnt.Close(lg)

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

	//ctx := context.Background()
	//ctx, cancel := context.WithCancel(ctx)
	//defer func() {
	//	cancel()
	//}()

	wg := new(sync.WaitGroup)

	eid,
		chanForConfirmKeepAliveEvent,
		chanForError,
		err :=
		s.initClient(
			lg,
			playerList,
			world,
			cnt,
			uid, username,
			wg,
		)
	if err != nil {
		panic(err)
	}
	defer s.closeClient(
		lg,
		playerList,
		world,
		eid, uid,
		chanForConfirmKeepAliveEvent,
		chanForError,
		wg,
	)

	for {
		select {
		case <-time.After(LoopDelayForPlayState):
			if err := cnt.LoopForPlayState(
				lg,
				world,
				eid,
				chanForConfirmKeepAliveEvent,
			); err != nil {
				panic(err)
			}
			break
		case err := <-chanForError:
			panic(err)
		}

	}
}

func (s *Server) Render(
	playerList *PlayerList,
	world *Overworld,
) {
	addr := s.addr
	network := Network
	lg := NewLogger(
		"server-renderer",
		NewLgElement("addr", addr),
		NewLgElement("network", network),
	)
	defer lg.Close()

	lg.Info("it is started to render")
	defer func() {
		lg.Info("it is finished to render")
	}()

	ln, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		lg.Info(
			"server accepted a new connection",
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
		go s.handleClient(
			playerList,
			world,
			cnt,
		)
	}

}
