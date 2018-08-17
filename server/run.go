package server

import (
	"log"

	"../messages"
	"github.com/golang/protobuf/proto"
)

func newHub() *Hub {
	return &Hub{
		request:    make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]int32),
		nextState:  0,
		peers:      make([]*messages.PeersInfo, maxPeers),
	}
}

// starts a run
// registers a peer when he want to participate in TX
// unregisters a peer
// listens for requests from peers and calls its corresponding handler
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if h.registration(client) {
				log.Printf("INCOMING C_JOIN_REQUEST - SUCCESSFUL")
			} else {
				log.Printf("INCOMING C_JOIN_REQUEST - FAILED")
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("INCOMING - USER UN-REGISTRATION - SUCCESSFUL")
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.request:
			handleRequest(message, h)
		}
	}
}

// adds a peer in h.peers if |h.peers| < MaxPeers
// send a failure message response to other peers
func (h *Hub) registration(client *Client) bool {
	h.Lock()
	defer h.Unlock()
	counter := int32(len(h.clients))

	if counter >= maxPeers {
		registration, err := proto.Marshal(&messages.RegisterResponse{
			Code:      messages.S_JOIN_RESPONSE,
			Id:        -1,
			Timestamp: timestamp(),
			Message:   "",
			Err:       "Limit Exceeded. Kindly try after some time",
		})

		checkError(err)
		client.send <- registration
		return false
	}

	counter++
	registration, err := proto.Marshal(&messages.RegisterResponse{
		Code:      messages.S_JOIN_RESPONSE,
		Id:        counter,
		Timestamp: timestamp(),
		Message:   "Welcome to CoinShuffle++. Waiting for other peers to join ...",
		Err:       "",
	})

	checkError(err)
	client.send <- registration

	h.clients[client] = counter
	h.peers[counter-1] = &messages.PeersInfo{Id: counter}
	h.peers[counter-1].MessageReceived = true

	if counter == maxPeers {
		// start DiceMix Light process
		// initRoundUUID(h)
		h.startDicemix()
	}
	return true
}

// initiates DiceMix-Light protocol
// send all peers ID's
func (h *Hub) startDicemix() {
	go broadcastDiceMixResponse(h, messages.S_START_DICEMIX, "Initiate DiceMix Protocol", "")
}
