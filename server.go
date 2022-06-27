package polaris

import (
	"net"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
)

type ServerOptions struct {
	Metadata map[string]string
}

type ServerSuite struct {
	Registry     registry.Registry
	RegistryInfo *registry.Info
	ServiceAddr  net.Addr
}

// Options implements the client.Suite interface.
func (ss *ServerSuite) Options() []server.Option {
	var opts []server.Option

	opts = append(opts, server.WithRegistry(ss.Registry))
	opts = append(opts, server.WithRegistryInfo(ss.RegistryInfo))
	opts = append(opts, server.WithServiceAddr(ss.ServiceAddr))
	
	return opts
}
