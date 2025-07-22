package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	protoServer "testdb/internal/grpc"
	"testdb/internal/memtable"
	"testdb/internal/snapshot"
	"testdb/internal/wal"
	"testdb/pkg/proto/dumbkv"
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

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Error("failed to listen", slog.Any("err", err))
		return
	}

	grpcServer := grpc.NewServer()
	ms := protoServer.NewMemTableServer(mt)
	dumbkv.RegisterDumbKVServiceServer(grpcServer, ms)
	log.Info("Server started on :50051")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("failed to serve", slog.Any("err", err))
		}
	}()

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
	grpcServer.GracefulStop()

}
