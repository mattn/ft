package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
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
	os.Stdout.WriteString("name,size,mode,modtime\n")
	for {
		file, err := slist.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		if runtime.GOOS != "windows" {
			os.Stdout.WriteString(fmt.Sprintf("\"%v\",\"%v\",\"%v\",\"%v\"\n",
				strings.Replace(file.Name, "\"", "\"\"", -1), file.Size, os.FileMode(file.Mode), time.Unix(file.ModTime.Seconds, 0).Format(time.RFC3339)))
		} else {
			os.Stdout.WriteString(fmt.Sprintf("\"%v\",\"%v\",\"%o\",\"%v\"\n",
				strings.Replace(file.Name, "\"", "\"\"", -1), file.Size, file.Mode, time.Unix(file.ModTime.Seconds, 0).Format(time.RFC3339)))
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
