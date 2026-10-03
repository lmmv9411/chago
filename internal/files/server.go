package files

import (
	"fmt"
	"net"
)

func StartServer() {

	listener, err := net.Listen("tcp", ":8081")

	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Servidor archivos escuchando en el puerto 8081...")

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error al aceptar la conexión: ", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
}
