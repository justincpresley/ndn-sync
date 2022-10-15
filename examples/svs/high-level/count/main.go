/*
 Copyright (C) 2022-2025, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 This library is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 This library is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. To see more details about the authors and
 contributors, please see AUTHORS.md. If absent, Both of which can be
 found within the GitHub repository:
          https://github.com/justincpresley/ndn-sync
*/

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
	callback := func(source string, seqno uint, data ndn.Data) {
		if data != nil {
			fmt.Println(source + ": " + string(data.Content().Join()))
		} else {
			fmt.Println("Unfetchable")
		}
	}
	sync := svs.NewNativeSync(app, svs.GetBasicNativeConfig(sourceName, syncPrefix, callback), svs.GetDefaultConstants())

	fmt.Println("Activating ...")
	sync.Listen()
	sync.Activate(true)
	defer sync.Shutdown()
	fmt.Println("Activated.\n")

	num := 1
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	clock := time.NewTimer(time.Duration(*interval) * time.Millisecond)
	fmt.Println("Starting Count ...")

loopCount:
	for {
		select {
		case <-clock.C:
			sync.Publish([]byte(strconv.Itoa(num)))
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
