

package Decoder

import (
    "fmt"
    "Msg"
)

func ProcessGate(m interface{}){
    fmt.Printf("Gate host is: `%v'\n", m.(*Msg.Gate).Host)
}

func ProcessHello(m interface{}){
    fmt.Printf("Hello to: `%v'\n", m.(*Msg.Hello).Name)
}
