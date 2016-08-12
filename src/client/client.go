package main

import (
    "encoding/binary"
    "net"
	"fmt"
    "time"
    "os"
    "os/signal"
    "sync"
    ldecoder "Decoder"
    lmsg     "Msg"
)

/*
func DecodeInBuf(inbuf[]byte)*FullPack{
    dec := json.NewDecoder(
            strings.NewReader(string(inbuf),
        ),
    )
    for {
        var h FullPack
        if err := dec.Decode(&h); io.EOF == nil {
            break
        } else if err != nil {
            panic(err)
        }
        return &h
    }
    panic("No decoding")
}*/

const HOST string = "127.0.0.1:3563"

type ClientPlayer struct {
    conn net.Conn   // It is Ptr already.
}

func CreateClientPlayer()*ClientPlayer{
    c, err := net.Dial("tcp", HOST)
    if err != nil {
        panic(err)
    }
    return &ClientPlayer{
        c,
    }
}


//~ This is the sample you need to read. 
//~ Send the raw buff to server.
//func (c*ClientPlayer)SendToHost(m interface{}) {
    /*
    data := []byte(`{
        "Hello": {
            "Name": "leaf"
        }
    }`)
    m := make([]byte, 2+len(data))
    binary.BigEndian.PutUint16(m, uint16(len(data)))
    copy(m[2:], data)
    c.conn.Write(m)
    */
    // c.conn.Write(DoMarshal(m))
//}

func (c*ClientPlayer) WriteToHost(bytes[]byte) (int, error) {
    return c.conn.Write(bytes)
}

func (c*ClientPlayer) ReadFromHost()[]byte{
    lbuff := make([]byte, 2)
    _, err := c.conn.Read(lbuff)
    if err != nil {  //经试验证明, 如果在别的地方Close, 那么这里就会出一个error(吃掉它)
        return nil
        //panic(err)
    }
    icome := binary.BigEndian.Uint16(lbuff)
    inbuf := make([]byte, icome)
    _, err = c.conn.Read(inbuf)
    if nil != err {
        panic(err)
    }
    return inbuf
}

func (c*ClientPlayer) Close() {
    c.conn.Close()
}

//~ And go.
func (lcp*ClientPlayer) writeProc(msg chan[]byte, wg *sync.WaitGroup){
    defer wg.Done()
    wg.Add(1)
    /*A100:for {
        select {
        case outGoingMsg:= <- msg:
            lcp.WriteToHost(outGoingMsg)
        case <-closeSig:
            break A100
        }
    }*/
    for outGoingMsg := range msg {
        if nil == outGoingMsg {
            fmt.Println("Out-going-msg is nil now !")
            break
        }
        _, err := lcp.WriteToHost(outGoingMsg)
        if err != nil {
            fmt.Println("Writing failed !")
            break
        }
        fmt.Println("Msg sent.")
    }
    fmt.Println("End of writeProc")
}

func (lcp*ClientPlayer) readProc(msg chan[]byte, wg *sync.WaitGroup){ //, closeSig chan bool){
    defer wg.Done()
    wg.Add(1)
    for {
        inBuff := lcp.ReadFromHost()
        if inBuff!=nil {
            msg <- inBuff //` Assume that it is big enough
        } else {
            fmt.Println("Read nil !")
            break
        }
    }
}

//~ And the length is already written.(Fixed)
// func clientLoop(msg chan[]byte, outbytes chan[] byte, closeSig chan bool, wg *sync.WaitGroup){
//     defer wg.Done()
//     wg.Add(1)
//     lcp := CreateClientPlayer()
//     defer lcp.Close()
//     A100:for {
//         select {
//         case m:= <-msg:
//             lcp.WriteToHost(m)
//             inbuf := lcp.ReadFromHost()
//             outbytes <- inbuf
//         case <- closeSig:
//             break A100
//         }
//     }
//     fmt.Println("goroutine finished")
// }

func main() {
    outMsgBuff := make(chan[]byte, 2048)
    inMsgBuff  := make(chan[]byte, 2048)
    // closeSig   := make(chan bool)
    decoder    := ldecoder.CreateDecoder()
    var waitgroup sync.WaitGroup //~ Do not pass it by value. 
    cp := CreateClientPlayer()
    //go clientLoop(outMsgBuff, inMsgBuff, closeSig, &waitgroup)
    go cp.writeProc(outMsgBuff, &waitgroup)
    go cp.readProc(inMsgBuff,   &waitgroup)
    //~ This is the wrapper you need.
    SendToHost := func(m interface{}){
        outMsgBuff <- decoder.Encode(m)
    }
    // SendToHost(&lmsg.Hello{"clienttest"})
    // The master loop
    tick := time.Tick(time.Second * 1)
    ctrlbreak := make(chan os.Signal)
    signal.Notify(ctrlbreak, os.Interrupt, os.Kill)
    A100:for{
        select {
        case <- tick:
            fmt.Println("One more out >>")
            SendToHost(&lmsg.Hello{"right"})
        case <- ctrlbreak:
            fmt.Println("Do you need to break ??")
            // closeSig <- true
            break A100
        case inBuff := <- inMsgBuff:
            fromSvr := decoder.Decode(inBuff)
            decoder.Dispatch(fromSvr)
        }
    }
    cp.Close()
    close(outMsgBuff)//~ out关闭, range直接退出. (都不会去整一个nil来糊弄玩家)
    waitgroup.Wait()
    fmt.Println("done")
}