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
	*sync.RWMutex

	addr string // address

	max    int // maximum number of players
	online int // number of online players

	favicon string // base64 png image string
	text    string // description of server

}

func NewServer(
	addr string,
	max int,
	favicon, text string,
) *Server {

	return &Server{
		new(sync.RWMutex),

		addr,
		max,
		0,

		favicon,
		text,
	}
}

func (s *Server) initClient(
	lg *Logger,
	pl *PlayerList,
	world Overworld,
	uid uuid.UUID, username string,
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

	eid := GetEIDCounter().count()

	CHForError := make(
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

	CHForAPEvent := make(
		ChanForAddPlayerEvent,
		MaxNumForChannel,
	)
	CHForULEvent := make(
		ChanForUpdateLatencyEvent,
		MaxNumForChannel,
	)
	CHForRPEvent := make(
		ChanForRemovePlayerEvent,
		MaxNumForChannel,
	)
	CHForSPEvent := make(
		ChanForSpawnPlayerEvent,
		MaxNumForChannel,
	)
	CHForSERPEvent := make(
		ChanForSetEntityRelativeMoveEvent,
		MaxNumForChannel,
	)
	CHForSELEvent := make(
		ChanForSetEntityLookEvent,
		MaxNumForChannel,
	)
	CHForSEAEvent := make(
		ChanForSetEntityActionsEvent,
		MaxNumForChannel,
	)
	CHForDEEvent := make(
		ChanForDespawnEntityEvent,
		MaxNumForChannel,
	)
	CHForLCEvent := make(
		ChanForLoadChunkEvent,
		MaxNumForChannel,
	)
	CHForUCEvent := make(
		ChanForUnloadChunkEvent,
		MaxNumForChannel,
	)

	CHForCWEvent := make(
		ChanForClickWindowEvent,
		MaxNumForChannel,
	)

	dim, err := NewDimension(
		pl,
		world,
		eid,
		uid, username,
		CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
		CHForCWEvent,
		cnt,
	)
	if err != nil {
		return nil, nil, nil, cancel, nil, err
	}

	go cnt.HandleCommonEvents(
		CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		CHForSPEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEAEvent,
		CHForDEEvent,
		CHForLCEvent,
		CHForUCEvent,
		CHForError,
		ctx,
		wg,
	)

	CHForCKAEvent := make(
		ChanForConfirmKeepAliveEvent,
		MaxNumForChannel,
	)
	go cnt.HandleConfirmKeepAliveEvent(
		CHForCKAEvent,
		dim,
		CHForError,
		ctx,
		wg,
	)

	return dim,
		CHForCKAEvent,
		CHForError,
		cancel,
		wg,
		nil
}

func (s *Server) closeClient(
	lg *Logger,
	dim *Dimension,
	CHForCKAEvent ChanForConfirmKeepAliveEvent,
	CHForError ChanForError,
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
	wg.Wait()

	CHForAPEvent,
		CHForULEvent,
		CHForRPEvent,
		CHForSPEvent,
		CHForDEEvent,
		CHForSERPEvent,
		CHForSELEvent,
		CHForSEMEvent,
		CHForLCEvent,
		CHForUCEvent,
		CHForCWEvent :=
		dim.Close()
	close(CHForAPEvent)
	close(CHForULEvent)
	close(CHForRPEvent)
	close(CHForSPEvent)
	close(CHForDEEvent)
	close(CHForSERPEvent)
	close(CHForSELEvent)
	close(CHForSEMEvent)
	close(CHForLCEvent)
	close(CHForUCEvent)

	close(CHForCWEvent)

	close(CHForCKAEvent)

	close(CHForError)
}

func (s *Server) handleClient(
	pl *PlayerList,
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
		"client has successfully logged in",
		NewLgElement("uid", uid),
		NewLgElement("username", username),
	)

	dim,
		CHForCKAEvent,
		CHForError,
		cancel,
		wg,
		err :=
		s.initClient(
			lg,
			pl,
			world,
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
		CHForCKAEvent,
		CHForError,
		cancel,
		wg,
	)

	//item := item2.NewStickItem(
	//	10,
	//	&nbt.ItemNbt{
	//		Display: &nbt.DisplayOfItemNbt{
	//			Name: "text",
	//		},
	//	},
	//)
	//SSIWPacket := packet.NewOutPacketToSetSlotInWindow(
	//	0,
	//	38,
	//	item,
	//)
	//cnt.writeWithComp(SSIWPacket)

	for {
		select {
		case <-time.After(LoopDelayForPlayState):
			if err := cnt.LoopToPlaying(
				lg,
				dim,
				CHForCKAEvent,
			); err != nil {
				panic(err)
			}

			break
		case err := <-CHForError:
			panic(err)
		}

	}
}

func (s *Server) Render(
	pl *PlayerList,
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

		cnt := NewClient(
			conn,
		)
		go s.handleClient(
			pl,
			world,
			cnt,
		)
	}

}
