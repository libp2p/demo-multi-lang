SRC		= dht-interop.go
BIN		= $(SRC:.go=)
DHT_SERVER 	= libp2p-bootstrap.goelzer.io

all: $(SRC)
	go build -o $(BIN) $(SRC)

install:
	-ssh $(DHT_SERVER) killall -9 dht-interop
	scp $(SRC:.go=) $(DHT_SERVER):~/
	ssh $(DHT_SERVER) ./$(BIN)
