package ebitenet

import "fmt"
import "net"
import "time"
import "os"

type Client struct {
	Number byte
	Ping time.Duration
	
	delay time.Duration
	
	conn net.Conn
	send chan Message
	recieve chan Message
	
	message Message
	WaitingForFrame byte
	
	inputs map[byte]bool
}

func (client *Client) handle() {
	go client.read()
	go client.write()
}

func (client *Client) read() {
	for {
		//Read header.
		header := make([]byte, 3)
		count, err := client.conn.Read(header)
		if err != nil {
			client.conn.Close()
			fmt.Println("disconnected.", err)
			os.Exit(1)
			break
		}
		
		if count < 3 {
			fmt.Println("header count wrong!", err)
			continue
		}
		
		var message Message
		message.Command = header[0]
		//println("recieved", message.Command)
		message.Len = header[1]
		message.Frame = header[2]
		
		
		//Read data.
		if message.Len > 0 {
			message.Data = make([]byte, message.Len)
		
			count, err := client.conn.Read(message.Data)
			if err != nil {
				client.conn.Close()
				fmt.Println("disconnected.", err)
				os.Exit(1)
				break
			}
			
			if count != int(message.Len) {
				fmt.Println("data packet wrong size!", err)
				continue
			}
		}
		
		if message.Command == SendPing {
			message.Command = SendPong
			client.send <- message
		} else if message.Command == SendPong {
			var ping = calculateping(message.Data)/2
			fmt.Println("Latency: ", ping)
			client.Ping = ping
		} else {
			
			client.recieve <- message
			
		}
		
		
		
		//fmt.Println(buf)
	}
}

func (client *Client) write() {
	for {
		var message = <- client.send
		
		if message.Ping {
			
			pingtime, _ := time.Now().MarshalBinary()
			if client.delay > 0 {
				time.Sleep(client.delay)
			}
			message.Command = SendPing
			message.Data = pingtime
			message.Len = byte(len(pingtime))
		}	
		
		client.conn.Write([]byte{message.Command, message.Len, message.Frame})
		if message.Len > 0 {
			client.conn.Write(message.Data)
		}
	}
}

