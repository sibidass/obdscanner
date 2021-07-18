package internal

import (
	"fmt"
	"net"
	"sync"
)

func NewDevice(conn RawConnection) (*RealDevice, error) {
	var err error

	dev := &RealDevice{
		mutex: sync.Mutex{},
		state: deviceReady,
		conn:  conn,
	}

	err = dev.Reset()
	if err != nil {
		goto out
	}

	err = dev.setAutoProtocol()
	if err != nil {
		goto out
	}

out:
	if err != nil {
		return nil, err
	} else {
		return dev, nil
	}
}

func NewWifiDevice(address string) (*WifiDevice, error) {
	tcpAddr, err := net.ResolveTCPAddr(networkWifi, address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to establish tcp connection. Error: %v",
			err)
	}

	d, err := NewDevice(conn)
	if err != nil {
		return nil, err
	}

	return &WifiDevice{d}, nil

}
