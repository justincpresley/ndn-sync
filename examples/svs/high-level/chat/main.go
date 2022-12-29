package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/apex/log"
	kyb "github.com/eiannone/keyboard"
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
	var input string
	var inputMutex sync.Mutex

	log.SetLevel(log.WarnLevel) // Change to "InfoLevel" to Look at Interests
	logger := log.WithField("module", "main")

	source := flag.String("source", "", "a string of the nodename")
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
	sourceName, _ := enc.NameFromStr(*source)
	callback := func(source string, seqno uint64, data ndn.Data) {
		inputMutex.Lock()
		fmt.Print("\n\033[1F\033[K")
		if data != nil {
			fmt.Println(source + ": " + string(data.Content().Join()))
		} else {
			fmt.Println("Unfetchable")
		}
		fmt.Print(input)
		inputMutex.Unlock()
	}
	sync := svs.NewSharedSync(app, svs.GetBasicSharedConfig(sourceName, syncPrefix, callback), svs.GetDefaultConstants())
	sync.Listen()
	sync.Activate(true)
	defer sync.Shutdown()
	fmt.Println("Entered the chatroom " + syncPrefix.String() + " as " + sourceName.String() + ".")

	if err := kyb.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = kyb.Close()
	}()

	fmt.Println("To leave, Press CTRL-C.")

InputLoop:
	for {
		char, key, err := kyb.GetKey()
		if err != nil {
			panic(err)
		}
		inputMutex.Lock()
		switch key {
		case kyb.KeyEnter:
			fmt.Print("\n\033[1F\033[K")
			if strings.TrimSpace(input) != "" {
				sync.PublishData([]byte(input))
				fmt.Println(sourceName.String() + ": " + input)
			}
			input = ""
		case kyb.KeyBackspace:
			if last := len(input) - 1; last >= 0 {
				input = input[:last]
			}
			fmt.Print("\n\033[1F\033[K")
			fmt.Print(input)
		case kyb.KeyBackspace2:
			if last := len(input) - 1; last >= 0 {
				input = input[:last]
			}
			fmt.Print("\n\033[1F\033[K")
			fmt.Print(input)
		case kyb.KeyCtrlC:
			fmt.Print("\n\033[1F\033[K")
			fmt.Println("Left the Chatroom, Exiting.")
			break InputLoop
		case kyb.KeySpace:
			input += " "
			fmt.Print(" ")
		default:
			input += string(char)
			fmt.Printf(string(char))
		}
		inputMutex.Unlock()
	}

	err = os.Remove("./" + *source + "_bolt.db")
	if err != nil {
		logger.Infof("Unable to remove the database that was created: %+v.", err)
	} else {
		logger.Info("Removed the svs database that was created.")
	}
}
