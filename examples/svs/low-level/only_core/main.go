package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/apex/log"
	svs "github.com/justincpresley/ndn-sync/pkg/svs"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
)

func passAll(enc.Name, enc.Wire, ndn.Signature) bool {
	return true
}

func main() {
	log.SetLevel(log.WarnLevel) // Change to "InfoLevel" to Look at Interests
	logger := log.WithField("module", "main")

	source := flag.String("source", "", "a string of the nodename")
	interval := flag.Uint("interval", 5000, "update frequency in # of milliseconds")
	flag.Parse()
	if *source == "" {
		logger.Errorf("A source is required to participate.")
		return
	}

	timer := eng.NewTimer()
	face := eng.NewStreamFace("unix", "/var/run/nfd/nfd.sock", true)
	app := eng.NewEngine(face, timer, sec.NewSha256IntSigner(timer), passAll)
	err := app.Start()
	if err != nil {
		logger.Errorf("Unable to start engine: %+v", err)
		return
	}
	defer app.Shutdown()

	syncPrefix, _ := enc.NameFromStr("/svs")
	nid, _ := enc.NameFromStr(*source)
	seqno := uint64(0)

	fmt.Println("Activating Core ...")
	config := &svs.TwoStateCoreConfig{
		SyncPrefix:     syncPrefix,
		FormalEncoding: false,
	}
	core := svs.NewCore(app, config, svs.GetDefaultConstants())
	core.Listen()
	core.Activate(true)
	defer core.Shutdown()
	fmt.Printf("Activated.\n\n")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	send := time.NewTimer(time.Duration(*interval) * time.Millisecond)
	recv := core.Subscribe()
	fmt.Println("Reporting all updates only while updating Core.")

loopCount:
	for {
		select {
		// Send updates periodically
		case <-send.C:
			seqno++
			core.Update(nid, seqno)
			send.Reset(time.Duration(*interval) * time.Millisecond)

		// Receive code when available
		case missing := <-recv:
			for _, m := range missing {
				for m.StartSeq <= m.EndSeq {
					fmt.Println(m.Dataset.String() + ": " + strconv.FormatUint(m.StartSeq, 10))
					m.StartSeq++
				}
			}

		// Close when Keyboard Interrupt
		case <-sigChannel:
			if !send.Stop() {
				<-send.C
			}
			logger.Infof("Received signal %+v - exiting.", sigChannel)
			break loopCount
		}
	}
}
