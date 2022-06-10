package svc

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bacv/pow-wow/lib"
	"github.com/stretchr/testify/assert"
)

func TestTransportWriteToClosed(t *testing.T) {
	conn, _ := net.Pipe()

	transport := NewTransport(conn, func(w ResponseWriter, r lib.Message) {})
	go func() {
		transport.Spawn()
	}()
	transport.Close()
	conn.Close()

	err := transport.Write(lib.Message{})
	assert.ErrorIs(t, err, ErrorWriteToClosed, "it should not be possible to write to a clossed transport")
	assert.True(t, transport.IsClosed())
}

func TestTransportHandler(t *testing.T) {
	expected := "of wisdom"
	connA, connB := net.Pipe()

	var sErr error
	serverHandler := func(w ResponseWriter, r lib.Message) {
		mt, _, err := r.Unmarshal()
		if err != nil {
			sErr = err
			return
		}

		switch mt {
		case lib.MsgRequest:
			w.Write(lib.NewChallengeMsg("test:1"))
		case lib.MsgProof:
			w.Write(lib.NewWordsMsg(expected))
		}
	}

	var cErr error
	var result string
	clientHandler := func(w ResponseWriter, r lib.Message) {
		mt, m, err := r.Unmarshal()
		if err != nil {
			cErr = err
			return
		}

		switch mt {
		case lib.MsgChallenge:
			w.Write(lib.NewProofMsg("1:1::test:::"))
		case lib.MsgWords:
			result = m
			w.Close()
		}

	}

	tA := NewTransport(connA, serverHandler)
	tB := NewTransport(connB, clientHandler)

	done := make(chan struct{})
	go func() {
		defer close(done)
		tA.Spawn()
	}()

	go func() {
		tB.Spawn()
	}()

	go func() {
		tB.Write(lib.NewRequestMsg())

	}()

	var err error
	select {
	case <-time.After(1 * time.Second):
		err = errors.New("timeout")
	case <-done:
		break
	}

	assert.NoError(t, err)
	assert.NoError(t, sErr)
	assert.NoError(t, cErr)
	assert.Equal(t, expected, result, fmt.Sprintf("result should be %s, got %s", expected, result))
}
