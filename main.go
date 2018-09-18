package main

import (
	"flag"
	"github.com/aemengo/vpnkit-manager/pb"
	"github.com/aemengo/vpnkit-manager/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

var (
	logger *log.Logger
)

func main() {
	var bindAddr string
	var addresses addressFlags
	flag.StringVar(&bindAddr, "bind-addr", "0.0.0.0:9998", "Bind on a tcp address in the following format: '0.0.0.0:9998'")
	flag.Var(&addresses, "expose", "Address to forward to the VM host in the following format: '0.0.0.0:8080:10.0.0.1:8080'")
	flag.Parse()

	logger = log.New(os.Stdout, "[VMGR] ", log.LstdFlags)

	lis, err := net.Listen("tcp", bindAddr)
	expectNoError(err)

	s := grpc.NewServer()
	srv, err := service.New(logger)
	expectNoError(err)

	pb.RegisterVpnkitManagerServer(s, srv)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGKILL)
	go killServerWhenStopped(sigs, s, logger)

	srv.ExposeAddressFlags(addresses)

	logger.Println("Initializing vpnkit-manager...")
	err = s.Serve(lis)
	expectNoError(err)
}

func killServerWhenStopped(sigs chan os.Signal, server *grpc.Server, logger *log.Logger) {
	<-sigs
	logger.Println("Shutting down vpnkit-manager...")
	server.Stop()
}

func expectNoError(err error) {
	if err != nil {
		logger.Fatalf("failed to initialize: %s\n", err)
	}
}
