package lookup

import (
	"fmt"
	"math/rand"
	"testing"
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

func (FakeUserStore) Contacts(userID string) ([]*Contact, error) {
	var contacts []*Contact
	for i := 0; i < 10000; i++ {
		contacts = append(contacts, &Contact{
			fmt.Sprintf("id:%s", randSeq(10)),
			fmt.Sprintf("%s@%s.com", randSeq(10), randSeq(5)),
			randSeq(8),
			randSeq(10),
		})
	}
	return contacts, nil
}

func TestLookup(t *testing.T) {
	for i := 0; i < 1; i++ {
		Run("idc", randSeq(i%6))
	}

	/*contacts := Run("idc", "a")
	for _, c := range contacts {
		log.Println(c.Email)
	}*/

	/*contacts = Run("idc", "ab")
	for _, c := range contacts {
		log.Println(c.Email)
	}
	contacts = Run("idc", "abc")
	for _, c := range contacts {
		log.Println(c.Email)
	}
	contacts = Run("idc", "bzk")
	for _, c := range contacts {
		log.Println(c.Email)
	}*/
}

func init() {
	store := &FakeUserStore{}
	WarmUp("idc", store)
}
