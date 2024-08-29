# goWake
Small and simple library to broadcast magic packets

## Usage 

The usage is trivial. Call the `Send()` method with the mac address as parameter. It will send a magic packet via a broadcast (`255.255.255.255`) on port _9_. 

```go
import "github.com/mitsimi/goWake"

func main() {
    const macAddr = "AA:BB:CC:DD:EE:FF"   // mac address of the target

    err := wol.Send(macAddr)
    if err != nil {
        panic(err)
    }
}
```

It is possible to send a broadcast only through a specified interface:
```go
import "github.com/mitsimi/goWake"

func main() {
    const (
        macAddr = "AA:BB:CC:DD:EE:FF"   // mac address of the target
        iface   = "eth0"                // interface of the host device
    )

    err := wol.SendWithInterface(macAddr, iface)
    if err != nil {
        panic(err)
    }
}
```

> [!NOTE] 
> The default port is _9_, therefore the host discards the packet after scanning. Currently it is not supported to receive an echo through sending to port _7_.
