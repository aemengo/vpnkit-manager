package main

import (
	"flag"
	"github.com/aemengo/vpnkit-manager/pb"
	"github.com/aemengo/vpnkit-manager/service"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
	flag.Var(&addresses, "address", "Address to forward to the VM host in the following format: '0.0.0.0:9998:10.0.0.1:9998'")
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

	logger.Printf("[DEBUG] addresses: %v\n", addresses) //TODO <- remove
	forward(addresses)

	logger.Println("Initializing vpnkit-manager...")
	err = s.Serve(lis)
	expectNoError(err)
}

func forward(addresses addressFlags) {
	for _, address := range addresses {
		logger.Printf("Attempting to expose port for %q...\n", address)
		elements := strings.Split(address, ":")
		if len(elements) != 4 {
			logger.Printf("%q is an invalid address...\n", address)
			continue
		}

		cmd := exec.Command("/usr/bin/vpnkit-expose-port", "-i", "-no-local-ip",
			"-host-ip", elements[0],
			"-host-port", elements[1],
			"-container-ip", elements[2],
			"-container-port", elements[3])

		if err := cmd.Start(); err != nil {
			logger.Printf("Failed to expose address %q: %s\n", address, err)
		}
	}
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
