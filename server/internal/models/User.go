package models

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var (
	// user databse gloabal
	UsersDatabase []User

	// Mutex to access database by multiple service
	usersMutex sync.Mutex
)

type User struct {
	ID   uuid.UUID `json:"id"`
	Nome string    `json:"nome"`
	Età  int       `json:"età"`
}

func AddUser(user User) {
	// lock user mutex
	usersMutex.Lock()
	// unlock at end of
	defer usersMutex.Unlock()
	UsersDatabase = append(UsersDatabase, user)
}

func PrintUser() {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	fmt.Println("Utenti nel database:")
	for _, utente := range UsersDatabase {
		fmt.Printf("ID: %d, Nome: %s, Età: %d\n", utente.ID, utente.Nome, utente.Età)
	}
}
func SearchUserByID(id uuid.UUID) []User {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	for _, user := range UsersDatabase {
		if user.ID == id {
			return append([]User{}, user)
		}
	}

	return nil
}

// dummy inizialize databse setup
func Setup() {
	parsedUUID, err := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		fmt.Println("Error parsing UUID:", err)
		return
	}
	AddUser(User{ID: parsedUUID, Età: 21, Nome: "Mario"})
	AddUser(User{ID: uuid.New(), Età: 14, Nome: "Francesco"})
}
