package main

import (
	"log"
	"os"

	"github.com/danomagnum/gologix"
)

// IOInput is sent to the PLC (4 bytes).
type IOInput struct {
	Data [4]byte
}

// IOOutput is received from the PLC (496 bytes of process data).
type IOOutput struct {
	Data [496]byte
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Printf("config: load error: %v (using defaults)", err)
	}

	// ioCh bridges the gologix channel provider to the historian.
	ioCh := make(chan [496]byte, 2048)

	provider := &gologix.IOChannelProvider[IOInput, IOOutput]{}

	// Bridge: receive from gologix -> buffered ioCh.
	go func() {
		dataCh := provider.GetOutputDataChannel()
		for data := range dataCh {
			select {
			case ioCh <- data.Data:
			default:
				log.Println("io: buffer full, dropping sample")
			}
		}
	}()

	go runHistorian(ioCh)

	// Start the gologix EIP server.
	// Configure a Generic Ethernet Module in the PLC IO tree pointing at this
	// machine's IP: 496 bytes output (from PLC), 4 bytes input (to PLC).
	go func() {
		r := gologix.PathRouter{}
		path := []byte{52, 4}
		r.Handle(path, provider)
		s := gologix.NewServer(&r)
		log.Printf("io: starting gologix server (TCP 44818 / UDP 2222)")
		if err := s.Serve(); err != nil {
			log.Printf("io: server exited: %v", err)
			os.Exit(1)
		}
	}()

	if err := serveWeb(":8000"); err != nil {
		log.Fatalf("web: %v", err)
	}
}
