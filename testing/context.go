package errorstesting

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
)

type TestContext struct {
	t *testing.T

	ServerOpts []grpc.ServerOption
	ClientOpts []grpc.DialOption

	serverListener net.Listener
	server         *grpc.Server
	clientConn     *grpc.ClientConn

	Service TestServiceServer
	Client  TestServiceClient
}

func CreateTestContext(t *testing.T) *TestContext {
	return &TestContext{t: t}
}

func (c *TestContext) AddUnaryServerInterceptor(i grpc.UnaryServerInterceptor) {
	c.ServerOpts = []grpc.ServerOption{
		grpc.UnaryInterceptor(i),
	}
}

func (c *TestContext) Setup() {
	if c.Service == nil {
		c.t.Fatal("Should set errorstesting.TestService implementaiton")
	}
	c.setupServer()
	c.setupClient()
}

func (c *TestContext) Teardown() {
	time.Sleep(10 * time.Millisecond)
	if c.serverListener != nil {
		c.server.GracefulStop()
		c.serverListener.Close()
	}
	if c.clientConn != nil {
		c.clientConn.Close()
	}
}

func (c *TestContext) ServerAddr() string {
	return c.serverListener.Addr().String()
}

func (c *TestContext) setupServer() {
	var err error
	c.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		c.Teardown()
		c.t.Fatal("Failed to listen local network")
	}
	c.server = grpc.NewServer(c.ServerOpts...)
	RegisterTestServiceServer(c.server, c.Service)
	go c.server.Serve(c.serverListener)
}

func (c *TestContext) setupClient() {
	var err error
	dialOpts := append(c.ClientOpts, grpc.WithBlock(), grpc.WithTimeout(2*time.Second), grpc.WithInsecure())
	c.clientConn, err = grpc.Dial(c.ServerAddr(), dialOpts...)
	if err != nil {
		c.Teardown()
		c.t.Fatal("Failed to create a client connection")
	}
	c.Client = NewTestServiceClient(c.clientConn)
}
