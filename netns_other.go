//go:build !linux

package main

import (
	"fmt"
	"net"
)

// dialerInNetns 在非 Linux 平台上的桩实现（主要用于本地测试与编译）。
func dialerInNetns(nsName string) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		return nil, fmt.Errorf("netns is only supported on linux")
	}
}

// forceIPv4Network 把 tcp/udp 收敛成 tcp4/udp4，已经指定版本的原样返回。
func forceIPv4Network(network string) string {
	switch network {
	case "tcp":
		return "tcp4"
	case "udp":
		return "udp4"
	}
	return network
}
