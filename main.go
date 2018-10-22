package main

import (
	"log"
	"time"
)

func main() {
	system := NewSystem()

	users := []*User{
		{
			ID:   1,
			Name: "Tom",
			Cash: 10,
		},
		{
			ID:   2,
			Name: "Jerry",
			Cash: 10,
		},
		{
			ID:   3,
			Name: "Spike",
			Cash: 10,
		},
	}
	transcations := []*Transcation{
		{
			TranscationID: 1,
			FromID:        1,
			ToID:          2,
			Cash:          10,
		},
		{
			TranscationID: 2,
			FromID:        2,
			ToID:          3,
			Cash:          5,
		},
		{
			TranscationID: 3,
			FromID:        3,
			ToID:          1,
			Cash:          20,
		},
		{
			TranscationID: 4,
			FromID:        2,
			ToID:          1,
			Cash:          10,
		},
	}

	for _, user := range users {
		if err := system.AddUser(user); err != nil {
			log.Printf("add user failed %v", err)
		}
	}

	complete := make(chan bool)

	go func() {
		for {
			select {
			case <-complete:
				break
			case <-time.After(500 * time.Millisecond):
				system.gcUndoLog()
			}
		}
	}()

	ch := make(chan int)

	// TODO: do transcation parallel
	for _, transcation := range transcations {
		go func(t *Transcation) {
			if err := system.DoTransaction(t); err != nil {
				log.Printf("do transcation %d failed %v", t.TranscationID, err)
			}
			ch <- 1
		}(transcation)
	}

	// wait for transactions complete
	for i := 0; i < len(transcations); i++ {
		<-ch
	}

	for _, user := range system.Users {
		log.Printf("after transcation, %s has %d money", user.Name, user.Cash)
	}

	log.Printf("%d active transactions in system", len(system.Transcations))

	if err := system.UndoTranscation(2); err != nil {
		log.Printf("undo transcation failed %v", err)
	}

	for _, user := range system.Users {
		log.Printf("after undo transcation, %s has %d money", user.Name, user.Cash)
	}

	complete <- true
}
