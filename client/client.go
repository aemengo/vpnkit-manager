package client

import (
	"context"
	"fmt"
	"github.com/aemengo/vpnkit-manager/pb"
	"google.golang.org/grpc"
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
	defer stream.CloseAndRecv()

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

	return nil
}

func newClient(target string) (pb.VpnkitManagerClient, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return pb.NewVpnkitManagerClient(conn), nil
}
