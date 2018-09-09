package service

//go:generate protoc -I ../pb --go_out=plugins=grpc:../pb ../pb/messages.proto

import (
	"context"
	"fmt"
	"github.com/aemengo/vpnkit-manager/pb"
	"log"
	"os/exec"
)

type Service struct {
	logger *log.Logger
}

func New(logger *log.Logger) (*Service, error) {
	logger.Println("Attempting to enable port forwarding...")
	err := runCommand("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err != nil {
		return nil, err
	}

	logger.Println("Attempting to enable internet connectivity for '10.0.0.1/16'...")
	err = runCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.0.0.1/16", "-o", "eth0", "-j", "MASQUERADE")
	if err != nil {
		return nil, err
	}

	return &Service{
		logger: logger,
	}, nil
}

func (s *Service) Ping(ctx context.Context, req *pb.Void) (*pb.TextParcel, error) {
	return &pb.TextParcel{Value: "pong"}, nil
}

func runCommand(path string, args ...string) error {
	output, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute: %s %v: %s: %s", path, args, err, output)
	}

	return nil
}