SRC		= go-peer.go
BIN		= $(SRC:.go=)
DHT_SERVER 	= libp2p-bootstrap.goelzer.io

all: $(SRC)
	go build -o $(BIN) $(SRC)

install:
	scp $(SRC:.go=) $(DHT_SERVER):~/
	ssh $(DHT_SERVER) ./$(BIN) --bootstrap
