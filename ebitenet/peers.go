package ebitenet

import "bytes"
import "fmt"
import "net"
//import "time"

//This file deals with discovering local peers on a Local Area Network.

type Peer struct {
	Address string
}

func multicastServer(port int16) {
	
	fmt.Println("LAN discovery server listening on 224.0.0.1:9999")
	
	addr, err := net.ResolveUDPAddr("udp", "224.0.0.1:9999")
	if err != nil {
		fmt.Println("Could not resolve network discovery address!")
		return
	}

	listener, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Could not broadcast to network discovery clients!")
	}
	
	//listener.SetDeadline(time.Time{})
	buf := make([]byte, 8192)
	for {
		n, addr, err := listener.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Discovery client sent unusual message!", n)
			continue
		}
		fmt.Println("Received ",string(buf[0:n]), " from ",addr)
		
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Println("Could not communicate with discovery client!")
			conn.Close()
			continue
		}
		
			
		n, err = conn.Write(Int(int64(port)))
		if n != 8 || err != nil {
			fmt.Println("Could not notify client of existence!")
			conn.Close()
			continue
		}
		
		conn.Close()
	}
}

func multicastClient(p chan Peer) {
	addr, err := net.ResolveUDPAddr("udp", "224.0.0.1:9999")
	if err != nil {
		fmt.Println("Could not resolve network discovery address!")
		return
	}
	
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Could not find server!")
		return
	}


	n, err := conn.Write([]byte{1})
	if err != nil || n != 1 {
		fmt.Println("Could not send request to server!")
		return
	}
	
	println("Waiting...")
	
	conn.Close()
	
	ser, err := net.ListenUDP("udp", conn.LocalAddr().(*net.UDPAddr))
    if err != nil {
        fmt.Printf("Some error %v\n", err)
        return
    }
	
	buf := make([]byte, 8192)
	n, addr, err = ser.ReadFromUDP(buf)
	if n != 8 || err != nil {
		fmt.Println("Discovery server sent unusual message!")
		return
	}
	
	var port int64
	reader := bytes.NewReader(buf)
	GetInt(reader, &port)
	if port > 0 {
		fmt.Println("Found Peer!")
		p <- Peer{Address: addr.IP.String()}
	} else {
		fmt.Println("Server sent an invalid port")
	}
	ser.Close()
	
}

func (network *Network) Discovery() {
	network.peer = make(chan Peer, 1)
	network.peers = make(chan Peer, 8)
	go multicastClient(network.peers)
}

func (network *Network) HasPeer() bool {
	select {
    case p, ok := <-network.peers:
        if ok {
            network.peer <- p
            return true
        } else {
            return false
        }
    default:
        return false
    }
}

func (network *Network) GetPeer() Peer {
	return <- network.peer
}
