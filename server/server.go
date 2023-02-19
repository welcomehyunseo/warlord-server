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

	rndDist    int32 // render distance
	spawnX     float64
	spawnY     float64
	spawnZ     float64
	spawnYaw   float32
	spawnPitch float32
}

func NewServer(
	addr string,
	max int,
	favicon, text string,
	rndDist int32,
	spawnX, spawnY, spawnZ float64,
	spawnYaw, spawnPitch float32,
) *Server {

	return &Server{
		addr:   addr,
		max:    max,
		online: 0,
		last:   0,

		favicon: favicon,
		text:    text,

		rndDist:    rndDist,
		spawnX:     spawnX,
		spawnY:     spawnY,
		spawnZ:     spawnZ,
		spawnYaw:   spawnYaw,
		spawnPitch: spawnPitch,
	}
}

func (s *Server) countEID() EID {
	a := s.last
	s.last++
	return a
}

func (s *Server) handleAddPlayerEvent(
	chanForEvent ChanForAddPlayerEvent,
	player *Player,
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
		NewLgElement("player", player),
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

				uid, username := event.GetUUID(), event.GetUsername()
				if err := cnt.AddPlayer(lg, uid, username); err != nil {
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
				if err := cnt.UpdateLatency(lg, uid, latency); err != nil {
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
	player *Player,
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
		NewLgElement("player", player),
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
	player *Player,
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
		NewLgElement("player", player),
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
						player,
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
	cnt *Client,
	uid UID, username string,
	wg *sync.WaitGroup,
) (
	*Player,
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

	spawnX, spawnY, spawnZ :=
		s.spawnX, s.spawnY, s.spawnZ
	spawnYaw, spawnPitch :=
		s.spawnYaw, s.spawnPitch

	player := NewPlayer(
		eid,
		uid,
		username,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	)

	if err := cnt.JoinGame(
		lg, eid,
		spawnX, spawnY, spawnZ,
		spawnYaw, spawnPitch,
	); err != nil {
		return nil, nil, nil, err
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
		player,
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
		player,
		cnt,
		chanForError,
		wg,
	)

	if err := playerList.InitPlayer(
		lg,
		player,
		cnt,
		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
	); err != nil {
		return nil, nil, nil, err
	}

	chanForConfirmKeepAliveEvent := make(
		ChanForConfirmKeepAliveEvent,
		MaxNumForChannel,
	)
	go s.handleConfirmKeepAliveEvent(
		playerList,
		chanForConfirmKeepAliveEvent,
		player,
		cnt,
		chanForError,
		wg,
	)

	return player,
		chanForConfirmKeepAliveEvent,
		chanForError,
		nil
}

func (s *Server) closeClient(
	lg *Logger,
	playerList *PlayerList,
	player *Player,
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

	chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent :=
		playerList.ClosePlayer(
			lg,
			player,
		)
	close(chanForAddPlayerEvent)
	close(chanForUpdateLatencyEvent)
	close(chanForRemovePlayerEvent)

	close(chanForConfirmKeepAliveEvent)

	wg.Wait()
	close(chanForError)
}

func (s *Server) handleClient(
	playerList *PlayerList,
	cnt *Client,
) {

	addr := cnt.GetAddr()
	lg := NewLogger(
		"client-handler",
		NewLgElement("addr", addr),
	)
	lg.Debug("it is started to handle Client")

	defer func() {

		if err := recover(); err != nil {
			lg.Error(err)
		}

		lg.Debug("it is finished to handle Client")
		lg.Close()
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

	player,
		chanForConfirmKeepAliveEvent,
		chanForError,
		err := s.initClient(
		lg,
		playerList,
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
		player,
		chanForConfirmKeepAliveEvent,
		chanForError,
		wg,
	)

	for {
		select {
		case <-time.After(LoopDelayForPlayState):
			if err := cnt.LoopForPlayState(
				lg,
				player,
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
			cnt,
		)
	}

}
