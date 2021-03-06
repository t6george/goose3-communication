package main

import (
	"fmt"
	"os"
	"time"

	"./lib/tls"
	"./lib/wpool"
	"./lib/wstream"

	"github.com/lucas-clemente/quic-go"
	"github.com/mogball/wcomms/wjson"
)

const podAddr = ":10000"

func main() {
	// Choose port to listen from
	config := quic.Config{IdleTimeout: 0}
	listener, err := quic.ListenAddr(podAddr, tls.GenerateConfig(), &config)
	CheckError(err)

	fmt.Println("Server started")
	go func() {
		for {
			session, err := listener.Accept() // Wait for call and return a Conn
			if err != nil {
				break
			}
			go HandleClient(session)
		}
	}()
	fmt.Println("Pool created")
	wpool := wpool.CreateWPool("localhost:12345", "localhost:12346", wpool.CommandHandler)
	go wpool.Serve()
	for {
		time.Sleep(time.Second)
		wpool.BroadcastPacket()
	}
}

// HandleClient accepts a wstream connection from the pod
func HandleClient(session quic.Session) {
	//defer session.Close(nil)
	wconn := wstream.AcceptConn(&session, []string{"sensor1", "sensor2", "sensor3", "command", "log"})
	fmt.Printf("%s %+v\n", "sss", wconn.Streams())
	for k, v := range wconn.Streams() {
		go HandleStream(k, v)
	}
}

// HandleStream takes each stream and reads the packets being sent
func HandleStream(channel string, wstream wstream.Stream) {
	defer wstream.Close()
	if (channel == "sensor1") || (channel == "sensor2") || (channel == "sensor3") {
		for {
			AcknowledgeMessage(wstream, 123)
			time.Sleep(time.Second)
		}
	} else {
		for {
			packet, err := wstream.ReadCommPacketSync()
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s %+v\n", channel, packet)
		}
	}
}

// AcknowledgeMessage lets the client know a message was recieved
func AcknowledgeMessage(wstream wstream.Stream, id uint8) {
	packet := &wjson.CommPacketJson{
		Time: 1323,
		Type: "State",
		Id:   id,
		Data: []float32{32.2323, 1222.22, 2323.11},
	}
	wstream.WriteCommPacketSync(packet)
}

// CheckError checks and print errors
func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
	}
}
