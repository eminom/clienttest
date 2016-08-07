package main

import (
    "encoding/json"
    "encoding/binary"
    "net"
	"fmt"
    "reflect"
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

func DoMarshal(m interface{})[]byte{
    msgID := reflect.TypeOf(m).Elem().Name()
    pac := map[string]interface{}{msgID:m}
    data, _ := json.Marshal(pac)
    buf1 := []byte(data)
    outBuff := make([]byte, 2 + len(buf1))
    binary.BigEndian.PutUint16(outBuff, uint16(len(buf1)))
    copy(outBuff[2:], buf1)
    return outBuff
}

//~ This is the sample you need to read. 
//~ Send the raw buff to server.
func (c*ClientPlayer)SendToHost(m interface{}) {
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
    c.conn.Write(DoMarshal(m))
}

func (c*ClientPlayer)WriteToHost(bytes[]byte){
    c.conn.Write(bytes)
}

func (c*ClientPlayer)ReadFromHost()[]byte{
    lbuff := make([]byte, 2)
    _, err := c.conn.Read(lbuff)
    if err != nil {
        panic(err)
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

//~ And the length is already written.(Fixed)
func clientLoop(msg chan[]byte, outbytes chan[] byte, closeSig chan bool, wg *sync.WaitGroup){
    defer wg.Done()
    wg.Add(1)
    lcp := CreateClientPlayer()
    defer lcp.Close()
    A100:for {
        select {
        case m:= <-msg:
            lcp.WriteToHost(m)
            inbuf := lcp.ReadFromHost()
            outbytes <- inbuf
        case <- closeSig:
            break A100
        }
    }
    fmt.Println("goroutine finished")
}

func main() {
    outMsgBuff := make(chan[]byte, 2048)
    inMsgBuff  := make(chan[]byte, 2048)
    closeSig   := make(chan bool)
    decoder    := ldecoder.CreateDecoder()
    var waitgroup sync.WaitGroup //~ Do not pass it by value. 
    go clientLoop(outMsgBuff, inMsgBuff, closeSig, &waitgroup)

    //~ This is the wrapper you need.
    SendToHost := func(m interface{}){
        outMsgBuff <- DoMarshal(m)
    }

    // SendToHost(&lmsg.Hello{"clienttest"})
    // The master loop
    tick := time.Tick(time.Second * 1)
    ctrlbreak := make(chan os.Signal)
    signal.Notify(ctrlbreak, os.Interrupt, os.Kill)
    A100:for{
        select {
        case <- tick:
            SendToHost(&lmsg.Hello{"right"})
        case <- ctrlbreak:
            fmt.Println("Do you need to break ??")
            closeSig <- true
            break A100
        case inBuff := <- inMsgBuff:
            fromSvr := decoder.Decode(inBuff)
            decoder.Dispatch(fromSvr)
        }
    }
    waitgroup.Wait()
    fmt.Println("done")
}