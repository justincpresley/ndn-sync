package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	sourceName, _ := enc.NameFromStr(*source)
	config := &svs.HealthConfig{
		Source:         sourceName,
		GroupPrefix:    syncPrefix,
		FormalEncoding: false,
	}
	sync := svs.NewHealthSync(app, config, svs.GetDefaultConstants())
	sync.Listen()
	sync.Activate(true)
	defer sync.Shutdown()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	recv := sync.Tracker().Chan()

loopCount:
	for {
		select {
		case change := <-recv:
			if change.OldStatus == svs.Unseen {
				fmt.Printf("%s is heard.\n", change.Node)
			} else if change.OldStatus == svs.Expired {
				fmt.Printf("%s renewed.\n", change.Node)
			} else {
				fmt.Printf("%s expired.\n", change.Node)
			}
		case <-sigChannel:
			logger.Infof("Received signal %+v - exiting.", sigChannel)
			break loopCount
		}
	}
}
