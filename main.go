package main

import (
	"fmt"
	"log"
	"math/rand"
	"unsafe"

	"github.com/janicduplessis/golookup/lookup"
)

type FakeUserStore struct{}

// Generate random letters
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (FakeUserStore) Contacts(userID string) ([]*lookup.Contact, error) {
	var contacts []*lookup.Contact
	for i := 0; i < 100000; i++ {
		contacts = append(contacts, &lookup.Contact{
			fmt.Sprintf("id:%s", randSeq(10)),
			fmt.Sprintf("%s@%s.com", randSeq(10), randSeq(5)),
			randSeq(8),
			randSeq(10),
		})
	}
	return contacts, nil
}

func main() {
	store := &FakeUserStore{}
	lookup.WarmUp("idc", store)

	for i := 0; i < 10000; i++ {
		lookup.Run("idc", randSeq(i%6))
	}

	contacts := make(map[int]lookup.Contact)
	for i := 0; i < 10000; i++ {
		contacts[i] = lookup.Contact{}
	}
	size := unsafe.Sizeof(contacts)
	log.Printf("Size: %v", size)
}
