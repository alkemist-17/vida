package vida

import (
	"log"
	"net"
)

func loadFoundationNetworkIO() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["listen"] = GFn(netListen)
	return m
}

func netListen(args ...Value) (Value, error) {
	if len(args) > 1 {
		if network, ok := args[0].(*String); ok {
			if address, ok := args[1].(*String); ok {
				tcpSocket, err := net.Listen(network.Value, address.Value)
				if err != nil {
					return Error{Message: &String{Value: err.Error()}}, nil
				}
				defer tcpSocket.Close()
				for {
					conn, err := tcpSocket.Accept()
					if err != nil {
						return Error{Message: &String{Value: err.Error()}}, nil
					}
					go func(c net.Conn) {
						buffer := make([]byte, 1024)
						_, err := conn.Read(buffer)
						if err != nil {
							log.Println(err)
						}
						println(c.LocalAddr().String())
						c.Close()
					}(conn)
				}
			}
		}
	}
	return NilValue, nil
}
