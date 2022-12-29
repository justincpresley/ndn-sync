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

func updateCallback(missing []svs.MissingData) {
	var temp uint64
	for _, m := range missing {
		temp = m.LowSeqno()
		for temp <= m.HighSeqno() {
			fmt.Println(m.Source() + ": " + strconv.FormatUint(uint64(temp), 10))
			temp++
		}
	}
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
	face := eng.NewStreamFace("unix", "/var/run/nfd.sock", true)
	app := eng.NewEngine(face, timer, sec.NewSha256IntSigner(timer), passAll)
	err := app.Start()
	if err != nil {
		logger.Errorf("Unable to start engine: %+v", err)
		return
	}
	defer app.Shutdown()

	syncPrefix, _ := enc.NameFromStr("/svs")
	nid, _ := enc.NameFromStr(*source)

	fmt.Println("Activating Core ...")
	config := &svs.CoreConfig{
		Source:         nid,
		SyncPrefix:     syncPrefix,
		UpdateCallback: updateCallback,
	}
	core := svs.NewCore(app, config, svs.GetDefaultConstants())
	core.Listen()
	core.Activate(true)
	defer core.Shutdown()
	fmt.Printf("Activated.\n\n")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	clock := time.NewTimer(time.Duration(*interval) * time.Millisecond)
	fmt.Println("Reporting all updates only while updating Core.")

loopCount:
	for {
		select {
		case <-clock.C:
			core.SetSeqno(core.GetSeqno() + 1)
			clock.Reset(time.Duration(*interval) * time.Millisecond)
		case <-sigChannel:
			if !clock.Stop() {
				<-clock.C
			}
			logger.Infof("Received signal %+v - exiting.", sigChannel)
			break loopCount
		}
	}
}
