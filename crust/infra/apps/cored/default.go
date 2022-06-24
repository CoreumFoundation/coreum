package cored

// DefaultPorts are the default ports cored listens on
var DefaultPorts = Ports{
	RPC:        26657,
	P2P:        26656,
	GRPC:       9090,
	GRPCWeb:    9091,
	PProf:      6060,
	Prometheus: 26660,
}
