package ebitenet

import "github.com/hajimehoshi/ebiten"
import "fmt"

const (
	SendNothing = iota
	SendPing
	SendPong
	SendKey
	Command = SendKey+1
)

var keyNames = map[ebiten.Key]string{
        ebiten.KeyBackspace: "BS",
        ebiten.KeyComma:     ",",
        ebiten.KeyDelete:    "Del",
        ebiten.KeyEnter:     "Enter",
        ebiten.KeyEscape:    "Esc",
        ebiten.KeyPeriod:    ".",
        ebiten.KeySpace:     "Space",
        ebiten.KeyTab:       "Tab",

        // Arrows
        ebiten.KeyDown:  "Down",
        ebiten.KeyLeft:  "Left",
        ebiten.KeyRight: "Right",
        ebiten.KeyUp:    "Up",

        // Mods
        ebiten.KeyShift:   "Shift",
        ebiten.KeyControl: "Ctrl",
        ebiten.KeyAlt:     "Alt",
}

func (network *Network) IsKeyPressed(Client byte, key ebiten.Key) bool {

	if Client == 1 {
		if network.Hosting {
			return network.inputs[byte(key)]
		} else {
			if len(network.clients) == 0 {
				return false
			}
			return network.clients[0].inputs[byte(key)]
		}
	}
	
	if Client > 1 && !network.Hosting {
	
		return network.inputs[byte(key)]
	
	} else {
		if len(network.clients) == 0 {
				return false
			}
		return network.clients[0].inputs[byte(key)]
	
	}
}


func (network *Network) SendInputs() {

	if network.DisableInput {
		return
	}

	pressed := []byte{}
	//Numbers.
    for i := 0; i <= 9; i++ {
            if ebiten.IsKeyPressed(ebiten.Key(i) + ebiten.Key0) {
                    pressed = append(pressed, byte(ebiten.Key(i) + ebiten.Key0))
            }
    }
    //Letters.
    for c := 'A'; c <= 'Z'; c++ {
            if ebiten.IsKeyPressed(ebiten.Key(c) - 'A' + ebiten.KeyA) {
                    pressed = append(pressed, byte(ebiten.Key(c) - 'A' + ebiten.KeyA))
            }
    }
    //Function buttons.
    for i := 1; i <= 12; i++ {
            if ebiten.IsKeyPressed(ebiten.Key(i) + ebiten.KeyF1 - 1) {
                    pressed = append(pressed, byte(ebiten.Key(i) + ebiten.KeyF1 - 1))
            }
    }
    //Other.
    for key, _ := range keyNames {
            if ebiten.IsKeyPressed(key) {
                    pressed = append(pressed, byte(key))
            }
    }
    if false && len(pressed) > 0 {
   		fmt.Println(pressed)
   	}
   	for _, key := range pressed {
   		network.Send(SendKey, []byte{key})
   	}
}
