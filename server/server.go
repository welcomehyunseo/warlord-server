package server

import (
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

const Network = "tcp"   // network type of server
const McName = "1.12.2" // minecraft version name
const ProtVer = 340     // protocol version

const CompThold = 16 // threshold for compression

const DelayForCheckKeepAlive = time.Millisecond * 1000
const LoopDelayForPlayState = time.Millisecond * 1

const MaxNumForChannel = 16

type ChanForError chan any

type GreenRoomManager struct {
}

func NewGreenRoomManager() *GreenRoomManager {
	return &GreenRoomManager{}
}

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

func (s *Server) initClient(
	lg *Logger,
	world Overworld,
	cnt *Client,
	uid UID, username string,
	wg *sync.WaitGroup,
) (
	Player,
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
	go cnt.HandleAddPlayerEvent(
		chanForAddPlayerEvent,
		chanForError,
		wg,
	)

	chanForUpdateLatencyEvent := make(
		ChanForUpdateLatencyEvent,
		MaxNumForChannel,
	)
	go cnt.HandleUpdateLatencyEvent(
		chanForUpdateLatencyEvent,
		chanForError,
		wg,
	)

	chanForRemovePlayerEvent := make(
		ChanForRemovePlayerEvent,
		MaxNumForChannel,
	)
	go cnt.HandleRemovePlayerEvent(
		chanForRemovePlayerEvent,
		chanForError,
		wg,
	)

	chanForConfirmKeepAliveEvent := make(
		ChanForConfirmKeepAliveEvent,
		MaxNumForChannel,
	)
	go cnt.HandleConfirmKeepAliveEvent(
		world,
		chanForConfirmKeepAliveEvent,
		uid,
		chanForError,
		wg,
	)

	chanForSpawnPlayerEvent := make(
		ChanForSpawnPlayerEvent,
		MaxNumForChannel,
	)
	go cnt.HandleSpawnPlayerEvent(
		chanForSpawnPlayerEvent,
		chanForError,
		wg,
	)

	chanForDespawnEntityEvent := make(
		ChanForDespawnEntityEvent,
		MaxNumForChannel,
	)
	go cnt.HandleDespawnEntityEvent(
		chanForDespawnEntityEvent,
		chanForError,
		wg,
	)

	chanForSetEntityRelativePosEvent := make(
		ChanForSetEntityRelativePosEvent,
		MaxNumForChannel,
	)
	go cnt.HandleSetEntityRelativePosEvent(
		chanForSetEntityRelativePosEvent,
		chanForError,
		wg,
	)

	chanForSetEntityLookEvent := make(
		ChanForSetEntityLookEvent,
		MaxNumForChannel,
	)
	go cnt.HandleSetEntityLookEvent(
		chanForSetEntityLookEvent,
		chanForError,
		wg,
	)

	chanForSetEntityMetadataEvent := make(
		ChanForSetEntityMetadataEvent,
		MaxNumForChannel,
	)
	go cnt.HandleSetEntityMetadataEvent(
		chanForSetEntityMetadataEvent,
		chanForError,
		wg,
	)

	chanForLoadChunkEvent := make(
		ChanForLoadChunkEvent,
		MaxNumForChannel,
	)
	go cnt.HandleLoadChunkEvent(
		chanForLoadChunkEvent,
		chanForError,
		wg,
	)

	chanForUnloadChunkEvent := make(
		ChanForUnloadChunkEvent,
		MaxNumForChannel,
	)
	go cnt.HandleUnloadChunkEvent(
		chanForUnloadChunkEvent,
		chanForError,
		wg,
	)

	player := NewGuest(
		eid, uid, username,
	)
	if err := world.InitPlayer(
		player,

		cnt,
		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,

		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,
	); err != nil {
		return nil, nil, nil, err
	}

	return player,
		chanForConfirmKeepAliveEvent,
		chanForError,
		nil
}

func (s *Server) closeClient(
	lg *Logger,
	world Overworld,
	player Player,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
	chanForError ChanForError,
	wg *sync.WaitGroup,
) {
	s.Lock()
	defer s.Unlock()

	lg.Debug("it is started to close Client")
	defer func() {
		lg.Debug("it is finished to close Client")
	}()

	eid := player.GetEid()

	chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,
		chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent :=
		world.ClosePlayer(
			eid,
		)
	close(chanForSpawnPlayerEvent)
	close(chanForDespawnEntityEvent)
	close(chanForSetEntityRelativePosEvent)
	close(chanForSetEntityLookEvent)
	close(chanForSetEntityMetadataEvent)
	close(chanForLoadChunkEvent)
	close(chanForUnloadChunkEvent)

	close(chanForAddPlayerEvent)
	close(chanForUpdateLatencyEvent)
	close(chanForRemovePlayerEvent)

	close(chanForConfirmKeepAliveEvent)

	wg.Wait()
	close(chanForError)
}

func (s *Server) handleClient(
	world Overworld,
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
		"Client has successfully logged in",
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
		err :=
		s.initClient(
			lg,
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
		world,
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
				world,
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
	world Overworld,
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
			world,
			cnt,
		)
	}

}
