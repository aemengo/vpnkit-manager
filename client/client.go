package client

import (
	"context"
	"fmt"
	"github.com/aemengo/vpnkit-manager/pb"
	"google.golang.org/grpc"
	"io"
	"strings"
)

func Ping(ctx context.Context, target string) error {
	client, err := newClient(target)
	if err != nil {
		return err
	}

	result, err := client.Ping(ctx, &pb.Void{})
	if err != nil {
		return err
	}

	if result.Value != "pong" {
		return fmt.Errorf("server returned unexpected value: %s", result.Value)
	}

	return nil
}

func Forward(ctx context.Context, target string, addresses []string) error {
	client, err := newClient(target)
	if err != nil {
		return err
	}

	stream, err := client.ExposeAddress(ctx)
	if err != nil {
		return err
	}

	for _, address := range addresses {
		elements := strings.Split(address, ":")
		if len(elements) != 4 {
			return fmt.Errorf("invalid forward address submitted: %s", address)
		}

		err = stream.Send(&pb.ExposeAddressOpts{
			HostIP:        elements[0],
			HostPort:      elements[1],
			ContainerIP:   elements[2],
			ContainerPort: elements[3]})
		if err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}

func ListForwarded(ctx context.Context, target string) ([]string, error) {
	client, err := newClient(target)
	if err != nil {
		return nil, err
	}

	stream, err := client.ListExposedAddresses(ctx, &pb.Void{})
	if err != nil {
		return nil, err
	}

	var addresses []string

	for {
		address, err := stream.Recv()
		if err == io.EOF {
			return addresses, nil
		}

		if err != nil {
			return nil, err
		}

		addresses = append(addresses, fmt.Sprintf("%s:%s:%s:%s",
			address.HostIP, address.HostPort, address.ContainerIP, address.HostPort))
	}
}

func newClient(target string) (pb.VpnkitManagerClient, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return pb.NewVpnkitManagerClient(conn), nil
}
