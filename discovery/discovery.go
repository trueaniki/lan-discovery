package discovery

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	CONN_PORT = "38747"

	DISCOVER = "DISCOVER"
	READY    = "READY"
)

func ListenForDiscover() (net.Conn, error) {
	lan, err := GetLanNetwork()
	if err != nil {
		return nil, fmt.Errorf("error getting LAN network: %v", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", GetMyIp(lan)+":"+CONN_PORT)
	if err != nil {
		return nil, fmt.Errorf("error resolving UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("error setting up UDP listener: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 32)

	for {
		// Expecting DISCOVER message
		n, remoteUdpAddr, err := conn.ReadFromUDP(buf)
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

func Discover() (net.Conn, error) {
	lan, err := GetLanNetwork()
	if err != nil {
		return nil, fmt.Errorf("error getting LAN network: %v", err)
	}

	broadcastAddr, err := net.ResolveUDPAddr("udp", GetBroadcastIp(lan)+":"+CONN_PORT)
	if err != nil {
		return nil, fmt.Errorf("error resolving remote address: %v", err)
	}

	localUdpAddr, err := net.ResolveUDPAddr("udp", GetMyIp(lan)+":"+"0")
	if err != nil {
		return nil, fmt.Errorf("error resolving local address: %v", err)
	}

	conn, err := net.DialUDP("udp", localUdpAddr, broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("error setting up UDP connection: %v", err)
	}
	defer conn.Close()

	// Send a discovery message
	buf := []byte(DISCOVER)
	_, err = conn.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("error sending discovery message: %v", err)
	}
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

	err = listener.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("error setting deadline: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return nil, fmt.Errorf("error accepting TCP connection: %v", err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("error reading from TCP connection: %v", err)
		}
		if string(buf[:n]) != READY {
			continue
		}
		return conn, nil
	}
}

func GetLanNetwork() (*net.IPNet, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet, nil
			}
		}
	}
	return nil, fmt.Errorf("no network found")
}

func GetMyIp(ipNet *net.IPNet) string {
	return ipNet.IP.To4().String()
}

func GetBroadcastIp(ipNet *net.IPNet) string {
	ip := ipNet.IP.To4()
	mask := ipNet.Mask

	broadcast := net.IPv4(0, 0, 0, 0).To4()
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}

	return broadcast.String()
}
