package internal

import (
	"reflect"
	"testing"
)

func TestNewRoom(t *testing.T) {
	expectedRoomId := RoomId("Test")
	room := NewRoom(expectedRoomId, make(chan<- RoomId))

	if room.id != expectedRoomId {
		t.Errorf("want room id %v, got %v", expectedRoomId, room.id)
	}

	if room.InProgress {
		t.Error("expected room not to be in progress")
	}
}

func TestRoom_everyDevGuessed(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		clients map[*Client]bool
	}{
		{
			name: "everyone guessed",
			want: true,
			clients: map[*Client]bool{
				&Client{
					Guess: 1,
					Role:  Developer,
				}: true,
				&Client{
					Role: ProductOwner,
				}: true,
			},
		},
		{
			name: "not everyone guessed",
			want: false,
			clients: map[*Client]bool{
				&Client{
					Guess: 0,
					Role:  Developer,
				}: true,
				&Client{
					Role: ProductOwner,
				}: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room := &Room{
				clients: test.clients,
			}
			got := room.everyDevGuessed()
			if got != test.want {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func TestRoom_Run_RegisteringAClient(t *testing.T) {
	room := NewRoom("Test", make(chan<- RoomId))
	client := &Client{}
	go room.Run()

	room.join <- client

	if _, ok := room.clients[client]; !ok {
		t.Error("expected room to have client")
	}
}

func TestRoom_Run_DeletingAClientAndDestroyingTheRoom(t *testing.T) {
	destroyChannel := make(chan RoomId)
	roomId := RoomId("Test")
	client := &Client{}
	room := &Room{
		id:         roomId,
		InProgress: false,
		leave:      make(chan *Client),
		join:       nil,
		clients: map[*Client]bool{
			client: true,
		},
		broadcast: nil,
		destroy:   destroyChannel,
	}
	go room.Run()

	room.leave <- client

	gotId := <-destroyChannel
	if gotId != roomId {
		t.Errorf("want room id %v, got %v", roomId, gotId)
	}

	if _, ok := room.clients[client]; ok {
		t.Error("expected room not to have client")
	}
}

func TestRoom_Run_BroadcastEstimate(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
	}
	room := &Room{
		id:         "Test",
		InProgress: false,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()

	msg := clientMessage{
		Type: estimate,
		Data: nil,
	}
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, msg) {
		t.Errorf("want message %v, got %v", msg, gotClientMsg)
	}

	if !room.InProgress {
		t.Error("expected room to be in progress")
	}
}

func TestRoom_Run_BroadcastDeveloperGuessed_EveryDeveloperGuessed(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		Guess: 1,
	}
	room := &Room{
		id:         "Test",
		InProgress: false,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()
	room.broadcast <- developerGuessed{}

	gotClientMsg := <-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, newEveryoneGuessed()) {
		t.Errorf("want msg %v, got %v", newEveryoneGuessed(), gotClientMsg)
	}
}

func TestRoom_Run_BroadcastDeveloperGuessed_NotEveryoneGuessed(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	room := &Room{
		id:         "Test",
		InProgress: false,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client: true,
			&Client{
				Role:  Developer,
				Guess: 0,
			}: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()
	msg := developerGuessed{}
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, msg) {
		t.Errorf("want msg %v, got %v", msg, gotClientMsg)
	}
}

func TestRoom_Run_BroadcastResetRound(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	developerToReset := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		Guess: 2,
	}
	room := &Room{
		id:         "Test",
		InProgress: true,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()

	msg := resetRound{}
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel
	<-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, msg) {
		t.Errorf("want msg %v, got %v", msg, gotClientMsg)
	}

	if room.InProgress {
		t.Error("expected room not to be in progress")
	}

	if developerToReset.Guess > 0 {
		t.Error("expected developer to be resetted")
	}
}
