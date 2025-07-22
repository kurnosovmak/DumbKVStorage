package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testdb/pkg/proto/dumbkv"
	"time"

	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := dumbkv.NewDumbKVServiceClient(conn)
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("MemTable CLI")
	fmt.Println("Commands: put <key> <value>, get <key>, delete <key>, size, exit")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		args := strings.SplitN(line, " ", 3)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		switch strings.ToLower(args[0]) {
		case "put":
			if len(args) < 3 {
				fmt.Println("usage: put <key> <value>")
				continue
			}
			_, err := client.Put(ctx, &dumbkv.PutRequest{
				Key:   args[1],
				Value: args[2],
			})
			if err != nil {
				fmt.Println("Put error:", err)
			} else {
				fmt.Println("OK")
			}

		case "get":
			if len(args) < 2 {
				fmt.Println("usage: get <key>")
				continue
			}
			resp, err := client.Get(ctx, &dumbkv.GetRequest{Key: args[1]})
			if err != nil {
				fmt.Println("Get error:", err)
			} else if !resp.Found {
				fmt.Println("Not found")
			} else {
				fmt.Printf("Value: %s\n", string(resp.Value))
			}

		case "delete":
			if len(args) < 2 {
				fmt.Println("usage: delete <key>")
				continue
			}
			_, err := client.Delete(ctx, &dumbkv.DeleteRequest{Key: args[1]})
			if err != nil {
				fmt.Println("Delete error:", err)
			} else {
				fmt.Println("Deleted")
			}

		case "size":
			resp, err := client.Size(ctx, &dumbkv.SizeRequest{})
			if err != nil {
				fmt.Println("Size error:", err)
			} else {
				fmt.Printf("Size: %d\n", resp.Size)
			}

		case "exit", "quit":
			fmt.Println("Bye!")
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}
