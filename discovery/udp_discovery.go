package discovery

import (
	"fmt"
	"net"
	"time"
)

const (
	SERVER_PORT = "38747"
	CLIENT_PORT = "38748"

	DISCOVER = "DISCOVER"
	READY    = "READY"
)

func RunDiscovery() (net.Conn, error) {
	doneChan := make(chan net.Conn)
	abortChan := make(chan struct{})
	errChan := make(chan error)

	go func() {
		conn, err := ListenForDiscover(abortChan)
		if err != nil {
			errChan <- err
		} else {
			doneChan <- conn
		}
	}()

	go func() {
		conn, err := Discover()
		if err, ok := err.(net.Error); ok && err.Timeout() {
			abortChan <- struct{}{}
		} else if err != nil {
			errChan <- err
		} else {
			doneChan <- conn
		}
	}()

	select {
	case conn := <-doneChan:
		abortChan <- struct{}{}
		return conn, nil
	case err := <-errChan:
		abortChan <- struct{}{}
		return nil, err
	}
}

func ListenForDiscover(abortCh chan struct{}) (net.Conn, error) {
	lan, err := GetLanNetwork()
	if err != nil {
		return nil, fmt.Errorf("error getting LAN network: %v", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", GetMyIp(lan)+":"+SERVER_PORT)
	if err != nil {
		return nil, fmt.Errorf("error resolving UDP address: %v", err)
	}

	pc, err := net.ListenPacket("udp4", ":"+SERVER_PORT)
	if err != nil {
		return nil, fmt.Errorf("error setting up UDP listener: %v", err)
	}

	buf := make([]byte, 32)

	for {
		select {
		case <-abortCh:
			return nil, nil
		default:
			// Expecting DISCOVER message
			n, remoteUdpAddr, err := pc.ReadFrom(buf)
			fmt.Println("Received ", string(buf[:n]), " from ", remoteUdpAddr)
			if err != nil {
				return nil, fmt.Errorf("error reading from UDP connection: %v", err)
			}
			if string(buf[:n]) != DISCOVER {
				continue
			}

			// Send a response over TCP
			tcpAddr, err := net.ResolveTCPAddr("tcp", udpAddr.String())
			if err != nil {
				return nil, fmt.Errorf("error resolving TCP address: %v", err)
			}

			remoteTcpAddr, err := net.ResolveTCPAddr("tcp", remoteUdpAddr.String())
			if err != nil {
				return nil, fmt.Errorf("error resolving remote TCP address: %v", err)
			}

			tcpConn, err := net.DialTCP("tcp", tcpAddr, remoteTcpAddr)
			if err != nil {
				return nil, fmt.Errorf("error setting up TCP connection: %v", err)
			}
			_, err = tcpConn.Write([]byte(READY))
			if err != nil {
				return nil, fmt.Errorf("error sending READY message: %v", err)
			}

			return tcpConn, nil
		}
	}
}

func Discover() (net.Conn, error) {
	lan, err := GetLanNetwork()
	if err != nil {
		return nil, fmt.Errorf("error getting LAN network: %v", err)
	}

	broadcastAddr, err := net.ResolveUDPAddr("udp", GetBroadcastIp(lan)+":"+SERVER_PORT)
	if err != nil {
		return nil, fmt.Errorf("error resolving remote address: %v", err)
	}

	localUdpAddr, err := net.ResolveUDPAddr("udp", GetMyIp(lan)+":"+CLIENT_PORT)
	if err != nil {
		return nil, fmt.Errorf("error resolving local address: %v", err)
	}

	udpConn, err := net.DialUDP("udp", localUdpAddr, broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("error setting up UDP connection: %v", err)
	}
	defer udpConn.Close()

	abortCh := make(chan struct{})
	// Send a discovery message every 500 ms
	go func() {
		for {
			select {
			case <-abortCh:
				return
			default:
				_, err := udpConn.Write([]byte(DISCOVER))
				if err != nil {
					fmt.Printf("error sending discovery message: %v\n", err)
				}
				fmt.Println("Sent discovery message to ", broadcastAddr)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	// Listen for a tcp response
	// Expecting READY message
	localTcpAddr, err := net.ResolveTCPAddr("tcp", localUdpAddr.String())
	if err != nil {
		return nil, fmt.Errorf("error resolving local TCP address: %v", err)
	}
	listener, err := net.ListenTCP("tcp", localTcpAddr)
	if err != nil {
		return nil, fmt.Errorf("error setting up TCP listener: %v", err)
	}
	defer listener.Close()

	err = listener.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("error setting deadline: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return nil, fmt.Errorf("error accepting TCP connection: %v", err)
		}
		buf := make([]byte, 32)
		n, err := conn.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("error reading from TCP connection: %v", err)
		}
		if string(buf[:n]) != READY {
			continue
		}
		abortCh <- struct{}{}
		return conn, nil
	}
}
