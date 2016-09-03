

package decoder

import (
    "fmt"
    "msg"
)

func ProcessGate(m interface{}){
    fmt.Printf("Gate host is: `%v'\n", m.(*msg.Gate).Host)
}

func ProcessHello(m interface{}){
    fmt.Printf("Hello to: `%v'\n", m.(*msg.Hello).Name)
}
