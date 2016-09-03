

package decoder

import (
    "encoding/binary"
    "encoding/json"
    "reflect"
    "fmt"
    "errors"
    "msg"
)

type Decoder struct {
    proto map [string] *DecoderInfo
}

type DecoderInfo struct {
    typeInfo reflect.Type
    msgHandler MsgHandler
}

type MsgHandler func(interface{})

func CreateDecoder()*Decoder{
    return (&Decoder{
        make(map[string]*DecoderInfo),
    }).init()
}

func (d*Decoder)init()*Decoder{
    d.Register(&msg.Hello{}, ProcessHello)
    d.Register(&msg.Gate{},  ProcessGate)
    return d
}

//~ Just any type
func (d*Decoder)Register(msg interface{}, msgHandler MsgHandler){
    msgType := reflect.TypeOf(msg)//~Is it new ? Is it unique ?
    if nil == msgType || msgType.Kind() != reflect.Ptr {
        panic("what the fuck")
    }
    msgID := msgType.Elem().Name()
    if _, ok := d.proto[msgID]; ok {
        panic("already registered")
    }
    d.proto[msgID] = &DecoderInfo{
        msgType,
        msgHandler,
    }
}

func (_*Decoder)Encode(m interface{})[]byte{
    msgID := reflect.TypeOf(m).Elem().Name()
    pac := map[string]interface{}{msgID:m}
    data, _ := json.Marshal(pac)
    buf1 := []byte(data)
    outBuff := make([]byte, 2 + len(buf1))
    binary.BigEndian.PutUint16(outBuff, uint16(len(buf1)))
    copy(outBuff[2:], buf1)
    return outBuff
}



func (d*Decoder)Decode(buf[]byte)interface{} {
    var m map[string]json.RawMessage
    err := json.Unmarshal(buf, &m)
    if err!=nil {
        panic(err)
    }
    if len(m) == 0 {
        panic(errors.New("No no no unmap failure"))
    }
    for msgID, data := range m {
        i, ok := d.proto[msgID]
        if !ok {
            panic(
                fmt.Errorf("message %v not registered",
                    msgID),
            )
        }
        msg := reflect.New(i.typeInfo.Elem()).Interface()
        json.Unmarshal(data, msg)
        return msg
    }
    panic("bug")
}

func (d*Decoder)Dispatch(m interface{}) {
    msgType := reflect.TypeOf(m)
    if nil == msgType || reflect.Ptr != msgType.Kind() {
        panic("No no no")
    }
    msgID := msgType.Elem().Name()
    info, ok := d.proto[msgID]
    if !ok {
        panic("No such " + msgID)
    }
    info.msgHandler(m)
}