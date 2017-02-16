package ebitenet

import (
	"net"
	"fmt"
	"time"
	"encoding/binary"
	"bytes"
	"os"
)

const (
	Tcp = iota
)

type Message struct {
	Ping bool
	Pong bool
	
	Command byte
	Len byte
	Frame byte
	
	Data []byte
}

type Network struct {
	Id byte
	
	Mode byte
	Port int16
	
	Frame byte
	Sent bool
	
	Synced bool
	
	Hosting bool
	
	DisableInput bool
	
	//This is the event handler.
	Event func(byte, byte, []byte)
	
	//This value adds fake lag to the network. 
	//This is useful for debugging.
	Delay time.Duration
	
	lastping time.Time
	PingDelay time.Duration
	
	//Internal variables.
	listener net.Listener
	
	//Mutex
	clients []*Client
	
	//LAN discovery.
	peer chan Peer
	peers chan Peer
	
	//We are a client to ourselves.
		self chan Message
	
		WaitingForFrame byte
		message Message
	
		//Lockstep input.
		inputs map[byte]bool
	
	//Syncing networking
	LastSentFrame byte
	SelfSync byte
	Skip byte
	
	Singleplayer bool
}

func Int(value int64) []byte {
	var buf = new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, value)
	return buf.Bytes()
}

func GetInt(data *bytes.Reader, value *int64) {
	binary.Read(data, binary.LittleEndian, value)
}

func (network *Network) Request(command byte, data []byte) {
	for _, client := range network.clients {
		client.send <- Message{Command:command, Len: byte(len(data)), Data:data, Frame:0}
	}
}

func (network *Network) Respond(client byte, command byte, data []byte) {
	if len(network.clients) > int(client) {
		network.clients[int(client)].send <- Message{Command:command, Len: byte(len(data)), Data:data, Frame:0}
	}
}

func (network *Network) Send(command byte, data []byte) {
	if !network.Sent && len(network.clients) > 0  {
		
		if network.LastSentFrame == 0 {
			//Find the largest ping, we have to wait for them.
			var greatestPing time.Duration = 0
			for _, client := range network.clients {
		
				//Wait for ping info.
				for  {
					if client.Ping > 0 {
						println("ping:", client.Ping)
						break
					}
				}
		
				if client.Ping > greatestPing {
					greatestPing = client.Ping
				}
			}
			println("gr ", greatestPing)
		
			//Lockstep, choose a target frame to send the message to.
			var FrameDelay = (byte(greatestPing/(time.Millisecond*16))+1)
			FrameDelay = FrameDelay + FrameDelay/2
			var FrameTarget = network.Frame+FrameDelay
		
			FrameTarget = FrameTarget % 60
			if FrameTarget == 0 {
				FrameTarget = 60
			}
		
			network.LastSentFrame = FrameTarget
		} else {
			network.LastSentFrame++
			network.LastSentFrame = network.LastSentFrame % 60
			if network.LastSentFrame == 0 {
				network.LastSentFrame = 60
			}
		}
		
		//fmt.Println("Send frame:", network.LastSentFrame)
		
		for _, client := range network.clients {
			client.send <- Message{Command:command, Len: byte(len(data)), Data:data, Frame:network.LastSentFrame}
		}
		if network.Synced {
			network.self <- Message{Command:command, Len: byte(len(data)), Data:data, Frame:network.LastSentFrame}
		}
		network.Sent = true
	} else if !network.Sent && len(network.clients) == 0 {
		network.self <- Message{Command:command, Len: byte(len(data)), Data:data, Frame:1}
		network.Sent = true
	}
}

func (network *Network) Update() {

	if len(network.clients) > 0 {
		network.Ping()
	}
	network.SendInputs()

	//If nothing was sent, send an empty command.
	if !network.Sent {
		network.Send(SendNothing, nil)
	}
	network.Sent = false

	if len(network.clients) == 0 {
		network.Singleplayer = true
		message := <- network.self
		
		network.inputs = make(map[byte]bool)
		
		//Process the message
		if message.Command == SendKey {
			network.inputs[message.Data[0]] = true
			
		} else if message.Command > SendKey {
			//TODO deal with multiple players.
			network.Event(message.Command, 1, message.Data)
		}
		
		return
	}
	if network.Singleplayer {
		network.self = make(chan Message, 60)
		network.Singleplayer = false
	}
	
	if network.Synced {
		network.Frame++
		if network.Frame == 61 {
			network.Frame = 1
		}
	} else {
		network.Frame = 1
	}
	
	if network.Skip > 0 {
		network.Skip--
		return
	}
	
	if !network.Synced {
		network.self = make(chan Message, 60)
	}
	
	if network.Synced {
		var message Message
		
		message = <- network.self
		
		message.Frame += network.SelfSync
		message.Frame = message.Frame % 60
		if message.Frame == 0 {
			message.Frame = 60
		}
		

		if message.Frame != network.Frame {
			if network.SelfSync == 0 {
				network.SelfSync = network.Frame-message.Frame
			} else {
				fmt.Println("FATAL selfish error, frame mismatch!")
				fmt.Println(message.Frame, network.Frame)
				os.Exit(1)
			}
		}

		network.inputs = make(map[byte]bool)

		//Process the message
		if message.Command == SendKey {
			network.inputs[message.Data[0]] = true
			
		} else if message.Command > SendKey {
			var id byte = 2
			if network.Hosting {
				id = 1
			}
			//TODO deal with multiple players.
			network.Event(message.Command, id, message.Data)
		}
	}
	
	//This is where we stop to recieve the information for the next frame.
	for i, client := range network.clients {
		var message Message
		
		//var now = time.Now()
		
		getframe:
		message = <- client.recieve
		
		if message.Frame == 0 {
		
			var id byte = 1
			if network.Hosting {
				id = 2
			}
			//TODO deal with multiple players.
			network.Event(message.Command, id, message.Data)
			
			println(string(message.Data))
			goto getframe
		}
		
		//fmt.Println("waited for message ", time.Since(now))
		
		//fmt.Println("Frame:", network.Frame)
		
		if !network.Synced && message.Frame != network.Frame {
			network.Synced = true
			network.Skip = message.Frame-network.Frame
			//println("skipping", network.Skip, "frames")
			continue
		}
		
		if message.Frame != network.Frame {
			fmt.Println("FATAL error, frame mismatch!")
			fmt.Println(message.Frame, network.Frame)
			os.Exit(1)
		}
		//Process the message
		client.inputs = make(map[byte]bool)
		if message.Command == SendKey {
			client.inputs[message.Data[0]] = true
		} else if message.Command > SendKey {
			var id byte = 1
			if network.Hosting {
				id = 2
			}
			//TODO deal with multiple players.
			network.Event(message.Command, id, message.Data)	
		}
		
		network.clients[i] = client
	}
}

func (network *Network) Connect(address string) (err error) {
	connection, err := net.Dial("tcp", address+":"+fmt.Sprint(network.Port))
	if err != nil {
		return err
	}

	var client Client
	client.conn = connection
	client.send = make(chan Message, 60)
	client.recieve = make(chan Message, 60)
	client.Number = byte(len(network.clients)+1)
	client.delay = network.Delay

	//LOCK
	network.clients = append(network.clients, &client)

	network.self = make(chan Message, 60)
	
	network.inputs = make(map[byte]bool)
	
	network.Id = 2

	go client.handle()
	return 
}

func (network *Network) Host() (err error) {
	network.listener, err = net.Listen("tcp", ":"+fmt.Sprint(network.Port))
	if err != nil {
		return err
	}

	network.Id = 1
	network.Hosting= true
	
	network.self = make(chan Message, 60)
	
	network.inputs = make(map[byte]bool)
	
	go multicastServer(network.Port)
	
	network.clients = make([]*Client, 0)
	go func() {
		for {
			connection, err := network.listener.Accept()
			if err != nil {
            	fmt.Println("Error accepting: ", err.Error()) 
            } else {
            	fmt.Println("Connection!")
            	
            	var client Client
				client.conn = connection
				client.send = make(chan Message, 60)
				client.recieve = make(chan Message, 60)
				client.Number = byte(len(network.clients)+1)
				client.delay = network.Delay
            	
            	
            	network.clients = append(network.clients, &client)
            	go client.handle()
            	client.send <- Message{Ping:true}
            }
		}
	}()
    return err
}
