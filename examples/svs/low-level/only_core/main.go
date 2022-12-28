/*
 Copyright (C) 2022-2030, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 ndn-sync is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 ndn-sync is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. If absent, it can be found within the
 GitHub repository:
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
	nid, _ := enc.NameFromStr(*source)

	fmt.Println("Activating Core ...")
	config := &svs.CoreConfig{
		Source:     nid,
		SyncPrefix: syncPrefix,
	}
	core := svs.NewCore(app, config, svs.GetDefaultConstants())
	core.Listen()
	core.Activate(true)
	defer core.Shutdown()
	fmt.Printf("Activated.\n\n")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	send := time.NewTimer(time.Duration(*interval) * time.Millisecond)
	recv := core.MissingChan()
	var temp uint64
	fmt.Println("Reporting all updates only while updating Core.")

loopCount:
	for {
		select {
		// Send updates peroidically
		case <-send.C:
			core.SetSeqno(core.GetSeqno() + 1)
			send.Reset(time.Duration(*interval) * time.Millisecond)

		// Receive code when avaliable
		case missing := <-recv:
			for _, m := range *missing {
				temp = m.LowSeqno()
				for temp <= m.HighSeqno() {
					fmt.Println(m.Source() + ": " + strconv.FormatUint(temp, 10))
					temp++
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
