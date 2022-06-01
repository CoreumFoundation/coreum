package targets

import (
	"net"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

var ipLocalhost = net.IPv4(127, 0, 0, 1)

type containerIPResolver struct {
}

func (ipr containerIPResolver) IPOf(app infra.IPProvider) net.IP {
	// FIXME (wojciech): if this returns 127.0.0.1 it means that container wants to connect to app running on host so host.docker.internal should be returned instead
	return app.IPSource().FromContainerIP()
}

type hostIPResolver struct {
}

func (ipr hostIPResolver) IPOf(app infra.IPProvider) net.IP {
	return app.IPSource().FromHostIP()
}
