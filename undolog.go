package main

const (
	START = iota
	UPDATE
	STARTCHECKPOINT
	ENDCHECKPOINT
)

// Undolog record format
// {START, tID, 0, 0}
// {UPDATE, tID, userID, cash}
// {STARTCHECKPOINT, 0, 0, 0}
// {ENDCHECKPOINT, 0, 0, 0}
type Record struct {
	Op            int
	TranscationId int
	UserId        int
	Cash          int
}
