package errorstesting

//go:generate protoc -I ./ -I ../vendor ./empty.proto --go_out=plugins=grpc:.
