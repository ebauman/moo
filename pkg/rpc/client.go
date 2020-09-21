package rpc

import "google.golang.org/grpc"

func SetupClients(hostname string, insecure bool) (MooClient, RulesClient, error) {
	var conn *grpc.ClientConn
	var err error

	if insecure {
		conn, err = grpc.Dial(hostname, grpc.WithInsecure())
	} else {
		conn, err = grpc.Dial(hostname)
	}

	if err != nil {
		return nil, nil, err
	}

	mooClient := NewMooClient(conn)
	rulesClient := NewRulesClient(conn)

	return mooClient, rulesClient, nil
}
