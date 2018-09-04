#SRC    	= main.go
SRC		= main-dht-bootstrap.go
BIN		= $(SRC:.go=)
DHT_SERVER 	= libp2p-bootstrap.goelzer.io

all: $(SRC)
#	go build -o $(SRC:.go=) $(SRC)
	go build -o $(BIN) $(SRC)

install:
	scp $(SRC:.go=) $(DHT_SERVER):~/
	ssh $(DHT_SERVER) ./$(BIN)
