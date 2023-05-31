package join

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/oklog/ulid/v2"

	"gochat/cmd/server/models"
	"gochat/internal/storage/inmemory/user"
)

type JoinHandler interface {
	Join(w http.ResponseWriter, r *http.Request)
}

func New(userStorage user.UserStorage) JoinHandler {
	return handler{
		userStorage: userStorage,
	}
}

type handler struct {
	*sync.Mutex
	userStorage user.UserStorage
}

func (h handler) Join(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err, "error reading body")
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		log.Fatal(err, "error unmarshalling body")
	}

	if _, err = h.userStorage.FindTokenByUsername(user.Name); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	token := ulid.Make()

	h.userStorage.Set(token, user.Name)

	tokenJson, err := json.Marshal(models.Token{
		Value: token,
	})
	if err != nil {
		log.Fatal(err, "error marshalling token")
	}

	w.Write(tokenJson)
}
