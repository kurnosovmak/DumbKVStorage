package grpc

import (
	"context"
	"testdb/internal/memtable"
	"testdb/pkg/proto/dumbkv"
)

type MemTableServer struct {
	dumbkv.UnimplementedDumbKVServiceServer
	table memtable.MemTable
}

func NewMemTableServer(table memtable.MemTable) *MemTableServer {
	return &MemTableServer{
		table: table,
	}
}

func (s *MemTableServer) Put(ctx context.Context, req *dumbkv.PutRequest) (*dumbkv.PutResponse, error) {
	err := s.table.Put(req.Key, []byte(req.Value))
	if err != nil {
		return nil, err
	}
	return &dumbkv.PutResponse{}, nil
}

func (s *MemTableServer) Get(ctx context.Context, req *dumbkv.GetRequest) (*dumbkv.GetResponse, error) {
	val, found := s.table.Get(req.Key)
	return &dumbkv.GetResponse{
		Value: string(val),
		Found: found,
	}, nil
}

func (s *MemTableServer) Delete(ctx context.Context, req *dumbkv.DeleteRequest) (*dumbkv.DeleteResponse, error) {
	err := s.table.Delete(req.Key)
	if err != nil {
		return nil, err
	}
	return &dumbkv.DeleteResponse{}, nil
}

func (s *MemTableServer) Size(ctx context.Context, _ *dumbkv.SizeRequest) (*dumbkv.SizeResponse, error) {
	size, err := s.table.Size()
	if err != nil {
		return nil, err
	}
	return &dumbkv.SizeResponse{Size: int32(size)}, nil
}
