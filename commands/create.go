package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/garden"
)

type Create struct {
	Handle     string        `short:"n" long:"handle" description:"name to give container"`
	RootFS     string        `short:"r" long:"rootfs" description:"rootfs image with which to create the container"`
	Env        []string      `short:"e" long:"env" description:"set environment variables"`
	Grace      time.Duration `short:"g" long:"grace" description:"grace time (resetting ttl) of container"`
	Privileged bool          `short:"p" long:"privileged" description:"privileged user in the container is privileged in the host"`
	Network    string        `long:"network" description:"the subnet of the container"`
	BindMounts []string      `short:"m" long:"bind-mount" description:"bind mount host-path:container-path"`
	NetIn      []string      `short:"i" long:"net-in" description:"map a host port to a container port"`
	NetOut     []string      `short:"o" long:"net-out" description:"whitelist outbound network traffic"`
}

func (command *Create) Execute(args []string) error {
	var bindMounts []garden.BindMount

	for _, pair := range command.BindMounts {
		segs := strings.SplitN(pair, ":", 2)
		if len(segs) != 2 {
			fail(fmt.Errorf("invalid bind-mount segment (must be host-path:container-path): %s", pair))
		}

		bindMounts = append(bindMounts, garden.BindMount{
			SrcPath: segs[0],
			DstPath: segs[1],
			Mode:    garden.BindMountModeRW,
			Origin:  garden.BindMountOriginHost,
		})
	}

	var netIns []garden.NetIn

	for _, pair := range command.NetIn {
		segs := strings.SplitN(pair, ":", 2)
		if len(segs) != 2 {
			fail(fmt.Errorf("invalid net-in segment (must be host-path:container-path): %s", pair))
		}

		var ports []uint32
		for _, seg := range segs {
			port, _ := strconv.Atoi(seg)
			ports = append(ports, uint32(port))
		}

		netIns = append(netIns, garden.NetIn{
			HostPort:      ports[0],
			ContainerPort: ports[1],
		})
	}

	var ips []garden.IPRange

	for _, network := range command.NetOut {
		ip := net.ParseIP(network)
		ips = append(ips, garden.IPRangeFromIP(ip))
	}

	netOutRule := garden.NetOutRule{
		Protocol: garden.ProtocolTCP,
		Networks: ips,
	}

	container, err := globalClient().Create(garden.ContainerSpec{
		Handle:     command.Handle,
		GraceTime:  command.Grace,
		RootFSPath: command.RootFS,
		Privileged: command.Privileged,
		Env:        command.Env,
		Network:    command.Network,
		BindMounts: bindMounts,
		NetIn:      netIns,
		NetOut:     []garden.NetOutRule{netOutRule},
	})

	failIf(err)

	fmt.Println(container.Handle())

	return nil
}
