package wow

import (
	"crypto/sha1"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bacv/pow-wow/lib/hashcash"
	"github.com/bacv/pow-wow/lib/protocol"
	"github.com/bacv/pow-wow/svc"
)

var ComputeTimeout = 60 * time.Second

type clientSvc struct{}

func NewWowClientService() svc.WowService {
	return &clientSvc{}
}

func (s *clientSvc) Handle(w svc.ResponseWriter, r protocol.Message) {
	mt, m, err := r.Unmarshal()

	if err != nil {
		log.Print(err)
		return
	}

	switch mt {
	case protocol.MsgChallenge:
		s.handleMsgChallenge(w, m)
	case protocol.MsgWords:
		s.handleMsgWords(w, m)
	default:
		w.Close()
	}
}

func (s *clientSvc) handleMsgChallenge(w svc.ResponseWriter, m string) {
	values := strings.Split(m, ":")
	bits, err := strconv.ParseUint(values[1], 10, 8)
	if err != nil {
		log.Print(err)
		return
	}

	hashC := make(chan string)
	go func() {
		hashC <- hashcash.NewHashcash(values[3], uint(bits), sha1.New()).Compute()
	}()

	select {
	case <-time.After(ComputeTimeout):
		log.Print("Compute timeout")
		return
	case header := <-hashC:
		w.Write(protocol.NewProofMsg(header))
	}
}

func (s *clientSvc) handleMsgWords(w svc.ResponseWriter, m string) {
	log.Print(m)
}
