package service

//go:generate protoc -I ../pb --go_out=plugins=grpc:../pb ../pb/messages.proto

import (
	"context"
	"fmt"
	"github.com/aemengo/vpnkit-manager/pb"
	"io"
	"log"
	"os/exec"
	"strings"
)

type Service struct {
	logger         *log.Logger
	savedAddresses []*pb.ExposeAddressOpts
}

func New(logger *log.Logger) (*Service, error) {
	logger.Println("Attempting to enable port forwarding...")
	err := runCommand("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err != nil {
		return nil, err
	}

	logger.Println("Attempting to enable internet connectivity...")
	err = runCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE")
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

func (s *Service) ExposeAddress(stream pb.VpnkitManager_ExposeAddressServer) error {
	for {
		addr, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		err = s.exposeAddress(addr.HostIP, addr.HostPort, addr.ContainerIP, addr.ContainerPort)
		if err != nil {
			return err
		}
	}

	return stream.SendAndClose(&pb.Void{})
}

func (s *Service) ExposeAddressFlags(addresses []string) {
	for _, address := range addresses {
		s.logger.Printf("Attempting to expose port for %q...\n", address)

		elements := strings.Split(address, ":")
		if len(elements) != 4 {
			s.logger.Printf("%q is an invalid address...\n", address)
			continue
		}

		err := s.exposeAddress(elements[0], elements[1], elements[2], elements[3])
		if err != nil {
			s.logger.Printf("Failed to expose address %q: %s\n", address, err)
		}
	}
}

func (s *Service) PerformPortMappings(addresses []string) {
	for _, address := range addresses {
		s.logger.Printf("Attempting to map port for %q...\n", address)

		elements := strings.Split(address, ":")
		if len(elements) != 3 {
			s.logger.Printf("%q is an invalid port mapping address...\n", address)
			continue
		}

		err := s.mapPort(elements[0], elements[1], elements[2])
		if err != nil {
			s.logger.Println(err)
		}
	}
}

func (s *Service) ListExposedAddresses(_ *pb.Void, stream pb.VpnkitManager_ListExposedAddressesServer) error {
	for _, addr := range s.savedAddresses {
		stream.Send(addr)
	}

	return nil
}

func (s *Service) exposeAddress(hostIP, hostPort, containerIP, containerPort string) error {
	if s.isExposed(hostIP, hostPort, containerIP, containerPort) {
		return nil
	}

	err := exec.Command("/usr/bin/vpnkit-expose-port", "-i", "-no-local-ip",
		"-host-ip", hostIP,
		"-host-port", hostPort,
		"-container-ip", containerIP,
		"-container-port", containerPort).Start()
	if err != nil {
		return err
	}

	s.savedAddresses = append(s.savedAddresses, &pb.ExposeAddressOpts{
		HostIP:        hostIP,
		HostPort:      hostPort,
		ContainerIP:   containerIP,
		ContainerPort: containerPort})
	return nil
}

func (s *Service) mapPort(srcPort, destinationIP, destinationPort string) error {
	output, err := exec.Command("iptables",
		"-t", "nat",
		"-A", "PREROUTING",
		"-i", "eth0",
		"-p", "tcp",
		"--dport", srcPort,
		"-j", "DNAT",
		"--to-destination", destinationIP+":"+destinationPort).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to map port %s -> %s:%s: %s: %s", srcPort, destinationIP, destinationPort, err, output)
	}

	return nil
}

func (s *Service) isExposed(hostIP, hostPort, containerIP, containerPort string) bool {
	for _, addr := range s.savedAddresses {
		if addr.HostIP == hostIP &&
			addr.HostPort == hostPort &&
			addr.ContainerIP == containerIP &&
			addr.ContainerPort == containerPort {
				return true
		}
	}

	return false
}

func runCommand(path string, args ...string) error {
	output, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute: %s %v: %s: %s", path, args, err, output)
	}

	return nil
}
