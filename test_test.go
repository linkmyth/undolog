package main

import (
	"log"
	"math/rand"
	"testing"
)

func TestDoTransactionSuccess(t *testing.T) {
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
	}
	transcation := &Transcation{
		TranscationID: 1,
		FromID:        1,
		ToID:          2,
		Cash:          10,
	}
	finalstate := []int{0, 20}

	for _, user := range users {
		if err := system.AddUser(user); err != nil {
			log.Printf("add user failed %v", err)
		}
	}

	if err := system.DoTransaction(transcation); err != nil {
		log.Printf("do transcation %d failed %v", transcation.TranscationID, err)
	}

	for i := range users {
		if users[i].Cash != finalstate[i] {
			t.Fatalf("user %d expected cash %v got %v", users[i].ID, finalstate[i], users[i].Cash)
		}
	}
}

func TestDoTransactionFail(t *testing.T) {
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
	}
	transcation := &Transcation{
		TranscationID: 1,
		FromID:        1,
		ToID:          2,
		Cash:          15,
	}
	finalstate := []int{10, 10}

	for _, user := range users {
		if err := system.AddUser(user); err != nil {
			log.Printf("add user failed %v", err)
		}
	}

	if err := system.DoTransaction(transcation); err != nil {
		log.Printf("do transcation %d failed %v", transcation.TranscationID, err)
	}

	for i := range users {
		if users[i].Cash != finalstate[i] {
			t.Fatalf("user %d expected cash %v got %v", users[i].ID, finalstate[i], users[i].Cash)
		}
	}
}

func TestUndoTransaction(t *testing.T) {
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
			Cash: 15,
		},
		{
			ID:   4,
			Name: "Bob",
			Cash: 20,
		},
	}
	undolog := []*Record{
		{
			Op:            START,
			TranscationId: 1,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            UPDATE,
			TranscationId: 1,
			UserId:        1,
			Cash:          15,
		},
		{
			Op:            START,
			TranscationId: 2,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            UPDATE,
			TranscationId: 2,
			UserId:        3,
			Cash:          30,
		},
		{
			Op:            UPDATE,
			TranscationId: 1,
			UserId:        2,
			Cash:          15,
		},
		{
			Op:            UPDATE,
			TranscationId: 2,
			UserId:        4,
			Cash:          35,
		},
	}
	finalstate := []int{15, 15, 30, 35}

	for _, user := range users {
		if err := system.AddUser(user); err != nil {
			log.Printf("add user failed %v", err)
		}
	}

	system.Undolog = undolog

	if err := system.UndoTranscation(1); err != nil {
		log.Printf("undo transcation %d failed", 1)
	}

	for i := range users {
		if users[i].Cash != finalstate[i] {
			t.Fatalf("user %d expected cash %v got %v", users[i].ID, finalstate[i], users[i].Cash)
		}
	}
}

func TestTransactionMany(t *testing.T) {
	system := NewSystem()

	users := []*User{
		{
			ID:   1,
			Name: "Tom",
			Cash: 100,
		},
		{
			ID:   2,
			Name: "Jerry",
			Cash: 100,
		},
		{
			ID:   3,
			Name: "Spike",
			Cash: 150,
		},
		{
			ID:   4,
			Name: "Bob",
			Cash: 200,
		},
		{
			ID:   5,
			Name: "Alice",
			Cash: 200,
		},
	}

	originalstate := []int{100, 100, 150, 200, 200}
	transcations := make([]*Transcation, 0, 20)

	maxUserID := 5

	for i := 1; i <= 20; i++ {
		FromID := rand.Intn(maxUserID) + 1
		ToID := rand.Intn(maxUserID) + 1
		for FromID == ToID {
			ToID = rand.Intn(maxUserID) + 1
		}
		transcations = append(transcations, &Transcation{
			TranscationID: i,
			FromID:        FromID,
			ToID:          ToID,
			Cash:          rand.Intn(20) + 1,
		})
	}

	originalSum := 0

	for _, user := range users {
		originalSum += user.Cash
		if err := system.AddUser(user); err != nil {
			log.Printf("add user failed %v", err)
		}
	}

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

	newSum := 0

	for _, user := range users {
		newSum += user.Cash
	}

	if originalSum != newSum {
		t.Fatalf("system in non-consistent state after many transactions")
	}

	if err := system.UndoTranscation(1); err != nil {
		log.Printf("undo transcation failed %v", err)
	}

	for i := range users {
		if users[i].Cash != originalstate[i] {
			t.Fatalf("user %d expected cash %v got %v", users[i].ID, originalstate[i], users[i].Cash)
		}
	}
}

func TestGcUndoLog(t *testing.T) {
	system := NewSystem()

	undolog := []*Record{
		{
			Op:            START,
			TranscationId: 1,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            UPDATE,
			TranscationId: 1,
			UserId:        1,
			Cash:          15,
		},
		{
			Op:            START,
			TranscationId: 2,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            UPDATE,
			TranscationId: 2,
			UserId:        3,
			Cash:          30,
		},
		{
			Op:            UPDATE,
			TranscationId: 1,
			UserId:        2,
			Cash:          15,
		},
		{
			Op:            UPDATE,
			TranscationId: 2,
			UserId:        4,
			Cash:          35,
		},
		{
			Op:            STARTCHECKPOINT,
			TranscationId: 0,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            START,
			TranscationId: 3,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            UPDATE,
			TranscationId: 3,
			UserId:        1,
			Cash:          15,
		},
		{
			Op:            START,
			TranscationId: 4,
			UserId:        0,
			Cash:          0,
		},
		{
			Op:            UPDATE,
			TranscationId: 4,
			UserId:        3,
			Cash:          30,
		},
		{
			Op:            UPDATE,
			TranscationId: 3,
			UserId:        2,
			Cash:          15,
		},
		{
			Op:            UPDATE,
			TranscationId: 4,
			UserId:        4,
			Cash:          35,
		},
		{
			Op:            ENDCHECKPOINT,
			TranscationId: 0,
			UserId:        0,
			Cash:          0,
		},
	}

	system.Undolog = undolog

	system.gcUndoLog()

	expectedLen := 10

	if len(system.Undolog) != expectedLen {
		t.Fatalf("gcUndoLog not work expected len %d current len %d", expectedLen, len(system.Undolog))
	}

}
