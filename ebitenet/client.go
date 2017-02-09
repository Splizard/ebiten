package ebitenet

import "fmt"
import "net"
import "time"
import "os"
import "bytes"
import "io"

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
		count, err := io.ReadFull(client.conn, header)
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
		
		//println("read net:", message.Len)
		
		//Read data.
		if message.Len == 255 {
			//Read length.
			l := make([]byte, 8)
			count, err := io.ReadFull(client.conn, l)
			if err != nil {
				client.conn.Close()
				fmt.Println("disconnected.", err)
				os.Exit(1)
				break
			}
			var length int64
			var b = bytes.NewReader(l)
			GetInt(b, &length)
			
			message.Data = make([]byte, length)
			
			count, err = io.ReadFull(client.conn, message.Data)
			if err != nil {
				client.conn.Close()
				fmt.Println("disconnected.", err)
				os.Exit(1)
				break
			}
			
			if count != int(length) {
				fmt.Println(message.Command-Command, "data packet wrong size!", count, "expecting", length)
				continue
			}
		
		} else if message.Len > 0 {
			message.Data = make([]byte, message.Len)
		
			count, err := io.ReadFull(client.conn, message.Data)
			if err != nil {
				client.conn.Close()
				fmt.Println("disconnected.", err)
				os.Exit(1)
				break
			}
			
			if count != int(message.Len) {
				fmt.Println(message.Command-Command, "data packet wrong size!", count, "expecting", int(message.Len))
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
		
		//println("write net: ", len(message.Data))
		if len(message.Data) > 255 {
			client.conn.Write([]byte{message.Command, 255, message.Frame})
			client.conn.Write(Int(int64(len(message.Data))))
			client.conn.Write(message.Data)
			fmt.Println("Sent large packet of: ", len(message.Data))
		} else {
			client.conn.Write([]byte{message.Command, message.Len, message.Frame})
			if message.Len > 0 {
				client.conn.Write(message.Data)
			}
		}
	}
}

