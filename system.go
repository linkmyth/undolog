package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

// User saves user's information
type User struct {
	ID   int
	Name string
	Cash int
	mu   sync.Mutex
}

// Transcation record a transcation.
type Transcation struct {
	TranscationID int
	FromID        int
	ToID          int
	Cash          int
}

// System keeps the user and transcation information
type System struct {
	sync.RWMutex

	Users map[int]*User

	// active transactions in system
	Transcations []*Transcation

	// TODO: add some variables about undo log
	Undolog []*Record
}

// NewSystem returns a System
func NewSystem() *System {
	return &System{
		Users:        make(map[int]*User),
		Transcations: make([]*Transcation, 0, 10),
		Undolog:      make([]*Record, 0),
	}
}

// AddUser adds a new user to the system
func (s *System) AddUser(u *User) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Users[u.ID]; ok {
		return errors.New("user id is already exists")
	}

	s.Users[u.ID] = u

	return nil
}

func (s *System) GetUsers(t *Transcation) (*User, *User) {
	fromUserID := t.FromID
	toUserID := t.ToID

	s.RLock()
	fromUser := s.Users[fromUserID]
	toUser := s.Users[toUserID]
	s.RUnlock()

	return fromUser, toUser
}

func (s *System) LockTwoUser(user1, user2 *User) error {
	// acquire lock from user with small userid
	if user1.ID < user2.ID {
		user1.mu.Lock()
		user2.mu.Lock()
	} else {
		user2.mu.Lock()
		user1.mu.Lock()
	}

	return nil
}

func (s *System) UnlockTwoUser(user1, user2 *User) {
	user1.mu.Unlock()
	user2.mu.Unlock()
}

func (s *System) Rollback(t *Transcation) error {
	user1, user2 := s.GetUsers(t)
	s.Lock()
	defer s.Unlock()

	s.LockTwoUser(user1, user2)
	defer s.UnlockTwoUser(user1, user2)

	tID := t.TranscationID

	for i := len(s.Undolog) - 1; i >= 0; i-- {
		if s.Undolog[i].Op == START && s.Undolog[i].TranscationId == tID {
			break
		}
		if s.Undolog[i].Op == UPDATE && s.Undolog[i].TranscationId == tID {
			s.Users[s.Undolog[i].UserId].Cash = s.Undolog[i].Cash
		}
	}

	return nil
}

func (s *System) RemoveTransaction(t *Transcation) error {
	s.Lock()
	defer s.Unlock()
	for i, transaction := range s.Transcations {
		if t == transaction {
			s.Transcations = append(s.Transcations[:i], s.Transcations[i+1:]...)
			return nil
		}
	}

	return errors.New("cannot find transactions to be deleted")
}

// DoTransaction applys a transaction
func (s *System) DoTransaction(t *Transcation) error {
	// TODO: implement DoTransaction
	// if after this transcation, user's cash is less than zero,
	// rollback this transcation according to undo log.
	s.Lock()
	s.Transcations = append(s.Transcations, t)
	//log.Printf("%d active transactions in system transaction id %d", len(s.Transcations), t.TranscationID)
	s.Unlock()
	defer s.RemoveTransaction(t)

	s.writeUndoLog(t)

	fromUser, toUser := s.GetUsers(t)
	s.LockTwoUser(fromUser, toUser)

	fromUser.Cash -= t.Cash
	toUser.Cash += t.Cash
	if fromUser.Cash < 0 {
		s.UnlockTwoUser(fromUser, toUser)
		if err := s.Rollback(t); err != nil {
			log.Printf("rollback failed")
		}
		return errors.New("from user does not have enough money")
	}

	s.UnlockTwoUser(fromUser, toUser)

	return nil
}

// writeUndoLog writes undo log to file
func (s *System) writeUndoLog(t *Transcation) error {
	// TODO: implement writeUndoLog
	fromUserID := t.FromID
	toUserID := t.ToID
	fromUser, toUser := s.GetUsers(t)
	s.Lock()
	defer s.Unlock()
	s.LockTwoUser(fromUser, toUser)
	defer s.UnlockTwoUser(fromUser, toUser)

	tID := t.TranscationID

	startRecord := &Record{START, tID, 0, 0}
	fromRecord := &Record{UPDATE, tID, fromUserID, fromUser.Cash}
	toRecord := &Record{UPDATE, tID, toUserID, toUser.Cash}
	s.Undolog = append(s.Undolog, startRecord)
	s.Undolog = append(s.Undolog, fromRecord)
	s.Undolog = append(s.Undolog, toRecord)
	return nil
}

// gcUndoLog the old undo log
func (s *System) gcUndoLog() {
	// TODO: implement gcUndoLog

	// begin trim undolog
	s.Lock()
	foundEnd := false
	for i := len(s.Undolog) - 1; i > 0; i-- {
		if s.Undolog[i].Op == ENDCHECKPOINT {
			foundEnd = true
		} else if foundEnd && s.Undolog[i].Op == STARTCHECKPOINT {
			s.Undolog = s.Undolog[i:]
			break
		}
	}
	s.Unlock()

	// try checkpoint undolog
	// mark all current active transaction
	s.Lock()
	waitTrans := make(map[int]int)
	for i, transcation := range s.Transcations {
		waitTrans[transcation.TranscationID] = i
	}
	record := &Record{STARTCHECKPOINT, 0, 0, 0}
	s.Undolog = append(s.Undolog, record)
	s.Unlock()
	for {
		end := true
		s.RLock()
		// wait all marked transaction finish
		for i, v := range waitTrans {
			if s.Transcations[v].TranscationID == i {
				end = false
				break
			}
		}
		s.RUnlock()
		if end {
			record := &Record{ENDCHECKPOINT, 0, 0, 0}
			s.Lock()
			s.Undolog = append(s.Undolog, record)
			s.Unlock()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// UndoTranscation roll back some transcations
func (s *System) UndoTranscation(fromID int) error {
	// TODO: implement UndoTranscation
	// undo transcation from fromID to the last transcation
	s.Lock()
	defer s.Unlock()
	for i := len(s.Undolog) - 1; i >= 0; i-- {
		if s.Undolog[i].Op == UPDATE && s.Undolog[i].TranscationId >= fromID {
			s.Users[s.Undolog[i].UserId].Cash = s.Undolog[i].Cash
		}
	}

	return nil
}
