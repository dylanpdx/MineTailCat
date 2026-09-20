package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tailscale/tailcat"
	"tailscale.com/wgengine/filter"
)

func main() {

	var mport uint16 = 25565
	args := os.Args[1:]

	if args[0] == "h" {
		host(mport)
	} else if args[0] == "c" {
		client(tailcat.Addr(args[1]))
	}
}

func client(addr tailcat.Addr) {
	c := tailcat.NewClient(addr)
	defer c.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Printf("Local Listen failed: %v", err)
		return
	}
	defer listener.Close()
	log.Printf("hosting @ %v", listener.Addr().String())

	local, err := listener.Accept()
	if err != nil {
		return
	}

	ctoh, err := c.DialTCPPort(context.Background(), 25565)
	if err != nil {
		log.Fatal(err)
	}
	tailcat.ProxyConns(local, ctoh)

}

func host(mport uint16) {
	s := &tailcat.Server{
		ServedTCPPorts: []filter.PortRange{{First: mport, Last: mport}},
		OnTCP: func(port uint16) func(net.Conn) {
			if port != mport { // what are you doing
				return nil
			}
			return func(c net.Conn) {
				minecraft, err := net.DialTimeout("tcp", "127.0.0.1:25565", 5*time.Second)
				if err != nil {
					log.Printf("DialTimeout -> MC failed: %v", err)
					return
				}
				defer minecraft.Close()

				tailcat.ProxyConns(c, minecraft) // MC <-> TC
			}
		},
	}
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(("TCA: " + s.TailcatAddr()))
	select {}
}
