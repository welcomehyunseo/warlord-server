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
	space *Space,
	uid UID, username string,
	cnt *Client,
) (
	*Dimension,
	ChanForConfirmKeepAliveEvent,
	ChanForError,
	context.CancelFunc,
	*sync.WaitGroup,
	error,
) {
	s.Lock()
	defer s.Unlock()

	lg.Debug("it is started to init Client")
	defer func() {
		lg.Debug("it is finished to init Client")
	}()

	eid := s.countEID()

	chanForError := make(
		ChanForError,
		MaxNumForChannel,
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	wg := new(sync.WaitGroup)

	if err := cnt.JoinGame(
		lg,
		eid,
	); err != nil {
		return nil, nil, nil, cancel, nil, err
	}

	chanForAPEvent := make(
		ChanForAddPlayerEvent,
		MaxNumForChannel,
	)
	chanForULEvent := make(
		ChanForUpdateLatencyEvent,
		MaxNumForChannel,
	)
	chanForRPEvent := make(
		ChanForRemovePlayerEvent,
		MaxNumForChannel,
	)
	chanForSPEvent := make(
		ChanForSpawnPlayerEvent,
		MaxNumForChannel,
	)
	chanForDEEvent := make(
		ChanForDespawnEntityEvent,
		MaxNumForChannel,
	)
	chanForSERPEvent := make(
		ChanForSetEntityRelativePosEvent,
		MaxNumForChannel,
	)
	chanForSELEvent := make(
		ChanForSetEntityLookEvent,
		MaxNumForChannel,
	)
	chanForSEMEvent := make(
		ChanForSetEntityMetadataEvent,
		MaxNumForChannel,
	)
	chanForLCEvent := make(
		ChanForLoadChunkEvent,
		MaxNumForChannel,
	)
	chanForUnCEvent := make(
		ChanForUnloadChunkEvent,
		MaxNumForChannel,
	)
	chanForUpCEvent := make(
		ChanForUpdateChunkEvent,
		MaxNumForChannel,
	)

	dim, err := NewDimension(
		space,
		eid,
		uid, username,
		0,
		chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent,
		cnt,
	)
	if err != nil {
		return nil, nil, nil, cancel, nil, err
	}

	go cnt.HandleCommonEvents(
		chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForError,
		ctx,
		wg,
	)

	go cnt.HandleUpdateChunkEvent(
		chanForUpCEvent,
		dim,
		chanForError,
		ctx,
		wg,
	)

	chanForCKAEvent := make(
		ChanForConfirmKeepAliveEvent,
		MaxNumForChannel,
	)
	go cnt.HandleConfirmKeepAliveEvent(
		chanForCKAEvent,
		dim,
		chanForError,
		ctx,
		wg,
	)

	return dim,
		chanForCKAEvent,
		chanForError,
		cancel,
		wg,
		nil
}

func (s *Server) closeClient(
	lg *Logger,
	dim *Dimension,
	chanForCKAEvent ChanForConfirmKeepAliveEvent,
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

	chanForAPEvent,
		chanForULEvent,
		chanForRPEvent,
		chanForSPEvent,
		chanForDEEvent,
		chanForSERPEvent,
		chanForSELEvent,
		chanForSEMEvent,
		chanForLCEvent,
		chanForUnCEvent,
		chanForUpCEvent :=
		dim.Close()
	close(chanForAPEvent)
	close(chanForULEvent)
	close(chanForRPEvent)
	close(chanForSPEvent)
	close(chanForDEEvent)
	close(chanForSERPEvent)
	close(chanForSELEvent)
	close(chanForSEMEvent)
	close(chanForLCEvent)
	close(chanForUnCEvent)
	close(chanForUpCEvent)

	close(chanForCKAEvent)

	cancel()
	wg.Wait()

	close(chanForError)
}

func (s *Server) handleClient(
	headCmdMgr *HeadCmdMgr,
	worldCmdMgr *WorldCmdMgr,
	space *Space,
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

	dim,
		chanForCKAEvent,
		chanForError,
		cancel,
		wg,
		err :=
		s.initClient(
			lg,
			space,
			uid, username,
			cnt,
		)
	if err != nil {
		cancel()
		panic(err)
	}
	defer s.closeClient(
		lg,
		dim,
		chanForCKAEvent,
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
				headCmdMgr,
				worldCmdMgr,
				dim,
				chanForCKAEvent,
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
	headCmdMgr *HeadCmdMgr,
	worldCmdMgr *WorldCmdMgr,
	space *Space,
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
			headCmdMgr,
			worldCmdMgr,
			space,
			cnt,
		)
	}

}
