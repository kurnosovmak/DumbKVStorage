package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"testdb/internal/memtable"
	"testdb/internal/snapshot"
	"testdb/internal/wal"
	"time"
)

const (
	testTestSnapshotFolder = "./"
	testDur                = time.Duration(20 * time.Second)
)

func main() {
	log := slog.Default()

	log.Info("starting")

	header, err := snapshot.Init(testTestSnapshotFolder)
	if err != nil {
		log.Error("failed to initialize snapshot header", slog.Any("err", err))
		return
	}

	walNumber := header.GetWalFileNumber()
	walName := fmt.Sprintf("%s%d%s", wal.WalFilePrefix, walNumber, wal.WalFileExt)
	log.Info("starting WAL", slog.String("wal name", walName))

	w, err := wal.NewWal(walName)
	if err != nil {
		log.Error("failed to open wal", slog.Any("err", err))
	}

	mt := memtable.NewMemTable(log, w, &header)

	timeStart := time.Now()
	err = mt.Init()
	if err != nil {
		log.Error("failed to init memtable", slog.Any("err", err))
	}
	log.Info("memtable initialized", slog.String("time", time.Since(timeStart).String()))

	for i := 0; i < 20_00; i++ {
		val, exists := mt.Get(fmt.Sprintf("key #%d", i))
		if !exists {
			log.Warn("missing key", slog.String("key", fmt.Sprintf("key #%d", i)))
		}
		if string(val) != fmt.Sprintf("value #%d", i) {
			log.Warn("missing value", slog.String("value", fmt.Sprintf("value #%d", i)))
		}
	}

	daemon := snapshot.NewSnapshotDaemon(testDur, &header, log, mt)
	cxt, cancel := context.WithCancel(context.Background())
	go func() {
		err := daemon.Run(cxt)
		if err != nil {
			log.Error("failed to start snapshot daemon", slog.Any("err", err))
			return
		}
	}()

	sing := make(chan os.Signal, 1)
	signal.Notify(sing, os.Interrupt)

	<-sing
	log.Info("shutting down")
	cancel()

}
