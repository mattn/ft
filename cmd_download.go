package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
	proto "github.com/mattn/ft/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

type downloader struct {
	dir      string
	client   proto.FileTransferServiceClient
	ctx      context.Context
	wg       sync.WaitGroup
	requests chan *proto.ListResponseType
	pool     *pb.Pool
}

func NewDownloader(ctx context.Context, client proto.FileTransferServiceClient, dir string) *downloader {
	d := &downloader{
		ctx:      ctx,
		client:   client,
		dir:      dir,
		requests: make(chan *proto.ListResponseType),
	}
	for i := 0; i < 5; i++ {
		d.wg.Add(1)
		go d.worker()
	}
	d.pool, _ = pb.StartPool()
	return d
}

func (d *downloader) Stop() {
	close(d.requests)
	d.wg.Wait()
	d.pool.RefreshRate = 500 * time.Millisecond
	d.pool.Stop()
}

func (d *downloader) worker() {
	defer d.wg.Done()

	for request := range d.requests {
		name := filepath.Join(d.dir, filepath.FromSlash(request.Name))
		if os.FileMode(request.Mode).IsDir() {
			err := os.MkdirAll(name, os.FileMode(request.Mode))
			if err != nil {
				log.Printf("%s: %v", request.Name, err)
			}
			continue
		}

		req := &proto.DownloadRequestType{
			Name: request.Name,
		}
		sdown, err := d.client.Download(d.ctx, req)
		if err != nil {
			log.Printf("%s: %v", request.Name, err)
			continue
		}

		f, err := os.Create(name)
		if err != nil {
			log.Printf("%s: %v", request.Name, err)
			sdown.CloseSend()
			continue
		}

		bar := pb.New64(request.Size).Postfix(" " + request.Name)
		bar.Units = pb.U_BYTES
		d.pool.Add(bar)
		for {
			res, err := sdown.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("%s: %v", request.Name, err)
				break
			}
			n, err := f.Write(res.Data)
			if err != nil {
				log.Printf("%s: %v", request.Name, err)
				break
			}
			bar.Add64(int64(n))
		}
		bar.Finish()
		f.Close()
		sdown.CloseSend()

		if err == nil && runtime.GOOS != "windows" {
			err = os.Chmod(name, os.FileMode(request.Mode))
			if err != nil {
				log.Printf("%s: %v", request.Name, err)
				break
			}
		}
		if err == nil {
			t := time.Unix(request.ModTime.Seconds, int64(request.ModTime.Nanos))
			err = os.Chtimes(name, t, t)
		}
	}
}

func (d *downloader) Do(file *proto.ListResponseType) {
	d.requests <- file
}

func downloadFiles(ctx context.Context, client proto.FileTransferServiceClient, dir string) error {
	slist, err := client.ListFiles(ctx, new(proto.ListRequestType))
	if err != nil {
		return err
	}

	d := NewDownloader(ctx, client, dir)
	for {
		file, err := slist.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		d.Do(file)
	}
	slist.CloseSend()
	d.Stop()

	return err
}

func downloadCommand() cli.Command {
	return cli.Command{
		Name:  "download",
		Usage: "download files from server",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "a",
				Value: ":11111",
				Usage: "server address",
			},
			cli.StringFlag{
				Name:  "d",
				Value: ".",
				Usage: "base directory",
			},
		},
		Action: func(c *cli.Context) error {
			conn, err := grpc.Dial(c.String("a"), grpc.WithInsecure())
			if err != nil {
				log.Fatalf("fail to dial: %v", err)
			}
			defer conn.Close()

			return downloadFiles(context.Background(), proto.NewFileTransferServiceClient(conn), c.String("d"))
		},
	}
}
