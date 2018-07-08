package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	proto "github.com/mattn/ft/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
			break
		}
		fmt.Printf("%q,\"%v\",\"%v\",\"%v\"\n",
			file.Name, file.Size, os.FileMode(file.Mode), time.Unix(file.ModTime.Seconds, 0).Format(time.RFC3339))
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
			cli.StringFlag{
				Name:  "tls-path",
				Value: "",
				Usage: "directory to the TLS server.crt file",
			},
		},
		Action: func(c *cli.Context) error {
			options := []grpc.DialOption{}
			if p := c.String("tls-path"); p != "" {
				creds, err := credentials.NewClientTLSFromFile(
					filepath.Join(p, "server.crt"),
					"")
				if err != nil {
					log.Println(err)
					return err
				}
				options = append(options, grpc.WithTransportCredentials(creds))
			} else {
				options = append(options, grpc.WithInsecure())
			}
			addr := c.String("a")
			if !strings.Contains(addr, ":") {
				addr += ":11111"
			}
			conn, err := grpc.Dial(addr, options...)
			if err != nil {
				log.Fatalf("cannot connect: %v", err)
			}
			defer conn.Close()

			return listFiles(context.Background(), proto.NewFileTransferServiceClient(conn))
		},
	}
}
