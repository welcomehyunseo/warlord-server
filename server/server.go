package server

import (
	"context"
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

const Network = "tcp"   // network type of server
const McName = "1.12.2" // minecraft version name
const ProtVer = 340     // protocol version

const CompThold = 1 // threshold for compression

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

func (s *Server) initClient(
	lg *Logger,
	gameManager *GameManager,
	world Overworld,
	uid UID, username string,
	cnt *Client,
	wg *sync.WaitGroup,
) (
	Player,

	*DimContext,
	ChanForError,
	context.CancelFunc,

	ChanForConfirmKeepAliveEvent,
	ChanForChangeDimEvent,

	error,
) {
	s.Lock()
	defer s.Unlock()

	lg.Debug("it is started to init Client")
	defer func() {
		lg.Debug("it is finished to init Client")
	}()

	eid := s.countEID()

	player := NewGuest(
		eid, uid, username,
	)
	dimContext := NewDimContext(
		world,
		player,
	)

	chanForError := make(
		ChanForError,
		MaxNumForChannel,
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	if err := cnt.JoinGame(
		lg,
		eid,
	); err != nil {
		return nil, nil, nil, cancel, nil, nil, err
	}

	chanForUpdateChunkEvent := make(
		ChanForUpdateChunkEvent,
		MaxNumForChannel,
	)
	go cnt.HandleUpdateChunkEvent(
		chanForUpdateChunkEvent,
		dimContext,
		chanForError,
		ctx,
		wg,
	)

	chanForAddPlayerEvent := make(
		ChanForAddPlayerEvent,
		MaxNumForChannel,
	)
	chanForUpdateLatencyEvent := make(
		ChanForUpdateLatencyEvent,
		MaxNumForChannel,
	)
	chanForRemovePlayerEvent := make(
		ChanForRemovePlayerEvent,
		MaxNumForChannel,
	)
	chanForSpawnPlayerEvent := make(
		ChanForSpawnPlayerEvent,
		MaxNumForChannel,
	)
	chanForDespawnEntityEvent := make(
		ChanForDespawnEntityEvent,
		MaxNumForChannel,
	)
	chanForSetEntityRelativePosEvent := make(
		ChanForSetEntityRelativePosEvent,
		MaxNumForChannel,
	)
	chanForSetEntityLookEvent := make(
		ChanForSetEntityLookEvent,
		MaxNumForChannel,
	)
	chanForSetEntityMetadataEvent := make(
		ChanForSetEntityMetadataEvent,
		MaxNumForChannel,
	)
	chanForLoadChunkEvent := make(
		ChanForLoadChunkEvent,
		MaxNumForChannel,
	)
	chanForUnloadChunkEvent := make(
		ChanForUnloadChunkEvent,
		MaxNumForChannel,
	)
	go cnt.HandleCommonEvents(
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

		chanForError,
		ctx,
		wg,
	)

	if err := world.InitPlayer(
		player,

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

		chanForUpdateChunkEvent,

		cnt,
	); err != nil {
		return nil, nil, nil, cancel, nil, nil, err
	}

	chanForConfirmKeepAliveEvent := make(
		ChanForConfirmKeepAliveEvent,
		MaxNumForChannel,
	)
	go cnt.HandleConfirmKeepAliveEvent(
		chanForConfirmKeepAliveEvent,
		dimContext,
		chanForError,
		ctx,
		wg,
	)

	chanForChangeDimEvent := make(
		ChanForChangeDimEvent,
		MaxNumForChannel,
	)
	if err := gameManager.InitPlayer(
		eid,
		chanForChangeDimEvent,
	); err != nil {
		return nil, nil, nil, cancel, nil, nil, err
	}

	return player,

		dimContext,
		chanForError,
		cancel,

		chanForConfirmKeepAliveEvent,
		chanForChangeDimEvent,

		nil
}

func (s *Server) closeClient(
	lg *Logger,
	gameManager *GameManager,
	world Overworld,
	player Player,
	chanForConfirmKeepAliveEvent ChanForConfirmKeepAliveEvent,
	chanForError ChanForError,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
) {
	s.Lock()
	defer s.Unlock()

	lg.Debug("it is started to close Client")
	defer func() {
		lg.Debug("it is finished to close Client")
	}()

	cancel()

	eid := player.GetEID()

	chanForChangeDimEvent :=
		gameManager.ClosePlayer(eid)
	close(chanForChangeDimEvent)

	close(chanForConfirmKeepAliveEvent)

	chanForAddPlayerEvent,
		chanForUpdateLatencyEvent,
		chanForRemovePlayerEvent,
		chanForSpawnPlayerEvent,
		chanForDespawnEntityEvent,
		chanForSetEntityLookEvent,
		chanForSetEntityRelativePosEvent,
		chanForSetEntityMetadataEvent,
		chanForLoadChunkEvent,
		chanForUnloadChunkEvent,

		chanForUpdateChunkEvent :=
		world.ClosePlayer(
			player,
		)
	close(chanForAddPlayerEvent)
	close(chanForUpdateLatencyEvent)
	close(chanForRemovePlayerEvent)
	close(chanForSpawnPlayerEvent)
	close(chanForDespawnEntityEvent)
	close(chanForSetEntityRelativePosEvent)
	close(chanForSetEntityLookEvent)
	close(chanForSetEntityMetadataEvent)
	close(chanForLoadChunkEvent)
	close(chanForUnloadChunkEvent)

	close(chanForUpdateChunkEvent)

	wg.Wait()

	close(chanForError)
}

func (s *Server) handleClient(
	gameManager *GameManager,
	lobby *Lobby,
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

	var world Overworld
	world = lobby

	wg := new(sync.WaitGroup)

	player,
		dimContext,
		chanForError,
		cancel,
		chanForConfirmKeepAliveEvent,
		chanForChangeDimEvent,
		err :=
		s.initClient(
			lg,
			gameManager,
			world,
			uid, username,
			cnt,
			wg,
		)
	if err != nil {
		panic(err)
	}
	defer s.closeClient(
		lg,
		gameManager,
		world,
		player,
		chanForConfirmKeepAliveEvent,
		chanForError,
		cancel,
		wg,
	)

	//eid := player.GetEID()

	for {
		select {
		case <-time.After(LoopDelayForPlayState):
			if err := cnt.LoopForPlayState(
				lg,
				gameManager,
				world,
				player,
				chanForConfirmKeepAliveEvent,
			); err != nil {
				panic(err)
			}

			break
		case event := <-chanForChangeDimEvent:
			world, player :=
				event.GetWorld(),
				event.GetPlayer()
			if err := dimContext.Change(
				world,
				player,
				cnt,
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
	gameManager *GameManager,
	lobby *Lobby,
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
			gameManager,
			lobby,
			cnt,
		)
	}

}
