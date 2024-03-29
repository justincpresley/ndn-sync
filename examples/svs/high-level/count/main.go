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
	sourceName, _ := enc.NameFromStr(*source)
	callback := func(source enc.Name, seqno uint64, data ndn.Data) {
		if data != nil {
			fmt.Println(source.String() + ": " + string(data.Content().Join()))
		} else {
			fmt.Println("Unfetchable")
		}
	}
	sync := svs.NewNativeSync(app, svs.GetBasicNativeConfig(sourceName, syncPrefix, callback), svs.GetDefaultConstants())

	fmt.Println("Activating ...")
	sync.Listen()
	sync.Activate(true)
	defer sync.Shutdown()
	fmt.Printf("Activated.\n\n")

	num := 1
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	clock := time.NewTimer(time.Duration(*interval) * time.Millisecond)
	fmt.Println("Starting Count ...")

loopCount:
	for {
		select {
		case <-clock.C:
			sync.PublishData([]byte(strconv.Itoa(num)))
			fmt.Println("Published: " + strconv.Itoa(num))
			clock.Reset(time.Duration(*interval) * time.Millisecond)
			num++
		case <-sigChannel:
			if !clock.Stop() {
				<-clock.C
			}
			logger.Infof("Received signal %+v - exiting.", sigChannel)
			break loopCount
		}
	}

	err = os.Remove("./" + *source + "_bolt.db")
	if err != nil {
		logger.Infof("Unable to remove the database that was created: %+v.", err)
	} else {
		logger.Info("Removed the svs database that was created.")
	}
}
