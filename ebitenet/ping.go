package ebitenet

import "time"


func calculateping(pingdata []byte) time.Duration {
	var pingtime time.Time
	pingtime.UnmarshalBinary(pingdata)
	return time.Since(pingtime)
}

func (network *Network) Ping() {
	if network.PingDelay == 0 {
		network.PingDelay = time.Second
	}
	
	if time.Since(network.lastping) > network.PingDelay {
		for _, client := range network.clients {
			client.send <- Message{Ping:true}
		}
		network.lastping = time.Now()
	}
}
