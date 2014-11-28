package lookup

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type (
	// Contact is the data structure for a contact.
	Contact struct {
		ID        string
		Email     string
		FirstName string
		LastName  string
	}

	// User represents a user querying the service.
	User struct {
		Contacts       []*Contact
		FirstNameIndex []*Contact
		LastNameIndex  []*Contact
	}

	// UserStore is a data store for users.
	UserStore interface {
		Contacts(userID string) ([]*Contact, error)
	}

	// Function signature for a looker.
	looker func(*User, string, chan *Contact, chan bool)

	byEmail     []*Contact
	byFirstName []*Contact
	byLastName  []*Contact
)

const (
	// Max time for a Run() request.
	timeoutMS = 250
)

var (
	users      map[string]*User
	usersMutex sync.RWMutex

	lookers []looker
)

func (e byEmail) Len() int           { return len(e) }
func (e byEmail) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e byEmail) Less(i, j int) bool { return e[i].Email < e[j].Email }

func (e byFirstName) Len() int           { return len(e) }
func (e byFirstName) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e byFirstName) Less(i, j int) bool { return e[i].FirstName < e[j].FirstName }

func (e byLastName) Len() int           { return len(e) }
func (e byLastName) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e byLastName) Less(i, j int) bool { return e[i].LastName < e[j].LastName }

// WarmUp loads the data in memory for a user.
func WarmUp(userID string, userStore UserStore) {
	contacts, err := userStore.Contacts(userID)
	if err != nil {
		log.Println(err)
		return
	}
	for _, c := range contacts {
		c.Email = strings.ToLower(c.Email)
		c.FirstName = strings.ToLower(c.FirstName)
		c.LastName = strings.ToLower(c.LastName)
	}

	// Additional indexes.
	firstNameIndex := make([]*Contact, len(contacts))
	copy(firstNameIndex, contacts)
	lastNameIndex := make([]*Contact, len(contacts))
	copy(lastNameIndex, contacts)

	// Sort indexes for binary search.
	sort.Sort(byEmail(contacts))
	sort.Sort(byFirstName(firstNameIndex))
	sort.Sort(byLastName(lastNameIndex))

	usersMutex.Lock()
	users[userID] = &User{
		Contacts:       contacts,
		FirstNameIndex: firstNameIndex,
		LastNameIndex:  lastNameIndex,
	}
	usersMutex.Unlock()
}

// Run returns contacts matching the input string.
func Run(userID string, input string) []*Contact {
	usersMutex.RLock()
	user := users[userID]
	usersMutex.RUnlock()
	return lookup(user, input)
}

func lookup(user *User, input string) []*Contact {
	var results []*Contact
	timeout := make(chan bool, 1)
	doneCh := make(chan bool, len(lookers))
	resultsCh := make(chan *Contact)
	doneCount := 0
	go func() {
		time.Sleep(timeoutMS * time.Millisecond)
		timeout <- true
	}()

	for _, looker := range lookers {
		go looker(user, input, resultsCh, doneCh)
	}

	for {
		select {
		case <-timeout:
			return results
		case <-doneCh:
			doneCount++
			if doneCount == len(lookers) {
				return results
			}
		case result := <-resultsCh:
			results = append(results, result)
		}
	}
}

func lookupByEmail(user *User, input string, results chan *Contact, doneCh chan bool) {
	input = strings.ToLower(input)
	startPos := sort.Search(len(user.Contacts), func(index int) bool {
		return input <= user.Contacts[index].Email
	})
	curPos := startPos
	for curPos < len(user.Contacts) && strings.HasPrefix(user.Contacts[curPos].Email, input) {
		results <- user.Contacts[curPos]
		curPos++
	}
	doneCh <- true
}

func lookupByFirstName(user *User, input string, results chan *Contact, doneCh chan bool) {
	input = strings.ToLower(input)
	startPos := sort.Search(len(user.FirstNameIndex), func(index int) bool {
		return input <= user.FirstNameIndex[index].Email
	})
	curPos := startPos
	for curPos < len(user.FirstNameIndex) && strings.HasPrefix(user.FirstNameIndex[curPos].Email, input) {
		results <- user.FirstNameIndex[curPos]
		curPos++
	}
	doneCh <- true
}

func lookupByLastName(user *User, input string, results chan *Contact, doneCh chan bool) {
	input = strings.ToLower(input)
	startPos := sort.Search(len(user.LastNameIndex), func(index int) bool {
		return input <= user.LastNameIndex[index].Email
	})
	curPos := startPos
	for curPos < len(user.LastNameIndex) && strings.HasPrefix(user.LastNameIndex[curPos].Email, input) {
		results <- user.LastNameIndex[curPos]
		curPos++
	}
	doneCh <- true
}

func init() {
	// Register lookup functions
	lookers = []looker{
		lookupByEmail,
		lookupByFirstName,
		lookupByLastName,
	}
	users = make(map[string]*User)
}
