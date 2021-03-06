package main

import (
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/bacv/pow-wow/svc"
	"github.com/bacv/pow-wow/svc/wow"
	"github.com/namsral/flag"
)

func main() {
	log.SetOutput(os.Stdout)
	rand.Seed(time.Now().UnixNano())
	port := flag.Int("port", 8080, "port to run words of wisdom server on")
	flag.Parse()

	wowSource := wow.NewWisdomSource()
	wowService := wow.NewWowServerService(wowSource)
	tcpServer := svc.NewTcpServer(wowService.Handle)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tcpServer.Serve(*port)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// TODO: gracefull shutdown
	wg.Wait()
}
