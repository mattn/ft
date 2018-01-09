package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	proto "github.com/mattn/ft/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func listFiles(ctx context.Context, client proto.FileTransferServiceClient) error {
	slist, err := client.ListFiles(ctx, new(proto.ListRequestType))
	if err != nil {
		return err
	}
	fmt.Println("name,size,mode,modtime")
	for {
		file, err := slist.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		if runtime.GOOS != "windows" {
			fmt.Printf("%q,\"%v\",\"%v\",\"%v\"\n",
				file.Name, file.Size, os.FileMode(file.Mode), time.Unix(file.ModTime.Seconds, 0).Format(time.RFC3339))
		} else {
			fmt.Printf("%q,\"%v\",\"%o\",\"%v\"\n",
				file.Name, file.Size, file.Mode, time.Unix(file.ModTime.Seconds, 0).Format(time.RFC3339))
		}
	}
	slist.CloseSend()
	return err
}

func listCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "list files from server by CSV format",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "a",
				Value: ":11111",
				Usage: "server address",
			},
		},
		Action: func(c *cli.Context) error {
			conn, err := grpc.Dial(c.String("a"), grpc.WithInsecure())
			if err != nil {
				log.Fatalf("fail to dial: %v", err)
			}
			defer conn.Close()

			return listFiles(context.Background(), proto.NewFileTransferServiceClient(conn))
		},
	}
}
