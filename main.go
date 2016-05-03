/*
Copyright 2009-2016 Weibo, Inc.

All files licensed under the Apache License, Version 2.0 (the "License");
you may not use these files except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/errors"
	"github.com/weibocom/wqs/config"
	"github.com/weibocom/wqs/engine/queue"
	"github.com/weibocom/wqs/log"
	"github.com/weibocom/wqs/protocol/http"
	"github.com/weibocom/wqs/protocol/mc"
)

var (
	configFile = flag.String("config", "config.properties", "qservice's configure file")
)

func initLogger(conf *config.Config) error {
	loggerInfo, err := log.NewLogger(conf.LogInfo).Open()
	if err != nil {
		return errors.Trace(err)
	}

	loggerDebug, err := log.NewLogger(conf.LogDebug).Open()
	if err != nil {
		loggerInfo.Close()
		return errors.Trace(err)
	}

	loggerProfile, err := log.NewLogger(conf.LogProfile).Open()
	if err != nil {
		loggerInfo.Close()
		loggerDebug.Close()
		return errors.Trace(err)
	}

	loggerInfo.SetFlags(log.LstdFlags)
	loggerInfo.SetLogLevel(log.LogInfo)
	log.RestLogger(loggerInfo, log.LogInfo)

	loggerDebug.SetFlags(log.LstdFlags | log.Llevel)
	loggerDebug.SetLogLevel(log.LogDebug)
	log.RestLogger(loggerDebug, log.LogFatal, log.LogError, log.LogWarning, log.LogDebug)

	loggerProfile.SetFlags(log.LstdFlags)
	loggerProfile.SetLogLevel(log.LogInfo)
	log.RestProfileLogger(loggerProfile)

	return nil
}

func main() {

	flag.Parse()

	conf, err := config.NewConfigFromFile(*configFile)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	if err = initLogger(conf); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	queue, err := queue.NewQueue(conf)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
		return
	}

	httpServer := http.NewHttpServer(queue, conf)
	go httpServer.Start()
	mcServer := mc.NewMcServer(queue, conf)
	go mcServer.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt, os.Kill)
	log.Info("Process start")
	<-c
	mcServer.Close()
	log.Info("Process stop")

}
