# goWake

Small and simple library to broadcast magic packets

## Installation

Install the latest version with:

```sh
$ go get -u github.com/mitsimi/goWake
```

## Usage

The usage is trivial. Call the `Send()` method with the mac address as parameter. It will send a magic packet via a broadcast (`255.255.255.255`) on port _9_.

```go
import "github.com/mitsimi/goWake"

func main() {
    const macAddr = "AA:BB:CC:DD:EE:FF"   // mac address of the target

    err := goWake.Send(macAddr)
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
        macAddr  = "00:11:22:33:44:55"   // mac address of the target
        iface    = "eth0"                // interface of the host device
        protocol = gowake.Discard        // target discards the message (port 9)
    )

    _, err = gowake.SendWithInterface(macAddr, iface, gowake.Discard)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Magic packet sent via Discard protocol")
    }
}
```

If your program needs the information if the target received the message or not, you can use the ECHO protocol on port 7

```go
import "github.com/mitsimi/goWake"

func main() {
    const (
        macAddr  = "00:11:22:33:44:55"   // mac address of the target
        iface    = "eth0"                // interface of the host device
        protocol = gowake.Echo           // target echoes the message (port 7)
    )

    echoMessage, err := gowake.SendWithInterface(macAddr, iface, gowake.Echo)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Received Echo Message:", echoMessage)
    }
}
```
