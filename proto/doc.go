package proto

//go:generate go get github.com/golang/protobuf/protoc-gen-go
//go:generate protoc -Iinclude -I. --go_out=plugins=grpc:. filetransfer_service.proto
