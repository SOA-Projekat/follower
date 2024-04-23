package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"follower.xws.com/model"
	"follower.xws.com/repo"
	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type FollowersHandler struct {
	logger *log.Logger
	repo   *repo.FollowersRepo
}

func NewFollowersHandler(l *log.Logger, r *repo.FollowersRepo) *FollowersHandler {
	return &FollowersHandler{l, r}
}

func (f *FollowersHandler) CreateUser(rw http.ResponseWriter, h *http.Request) {
	user := h.Context().Value(KeyProduct{}).(*model.User)
	userSaved, err := f.repo.SaveUser(user)
	if err != nil {
		f.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userSaved {
		f.logger.Print("New user saved to database")
		rw.WriteHeader(http.StatusCreated)
	} else {
		rw.WriteHeader(http.StatusConflict)
	}
}

func (f *FollowersHandler) UnfollowUser(rw http.ResponseWriter, h *http.Request) {
	userId1 := h.URL.Query().Get("followerId")
	userId2 := h.URL.Query().Get("followedId")
	err := f.repo.DeleteFollowing(userId1, userId2)
	if err != nil {
		f.logger.Print("Database exception: ", err)
		return
	}
	user := model.User{}
	jsonData, _ := json.Marshal(user)
	rw.Write(jsonData)
}

func (f *FollowersHandler) CreateFollowing(rw http.ResponseWriter, h *http.Request) {
	// Parse the request body
	decoder := json.NewDecoder(h.Body)
	defer h.Body.Close()

	// Define a variable to store the decoded data
	var users []model.User

	// Decode the JSON data into the users slice
	if err := decoder.Decode(&users); err != nil {
		// Handle decoding error
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Error decoding JSON data: " + err.Error()))
		return
	}

	// Ensure that the users slice contains exactly two users
	if len(users) != 2 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Expecting exactly two users"))
		return
	}

	// Now you have users[0] and users[1] containing the data sent from the front end
	// You can use this data to create new User objects or perform any other operations

	// Example:
	newUser1 := model.User{UserId: users[0].UserId, Username: users[0].Username, ProfileImage: users[0].ProfileImage}
	newUser2 := model.User{UserId: users[1].UserId, Username: users[1].Username, ProfileImage: users[1].ProfileImage}
	err := f.repo.SaveFollowing(&newUser1, &newUser2)
	if err != nil {
		f.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	newUser1 = model.User{}
	jsonData, _ := json.Marshal(newUser1)
	rw.Write(jsonData)
	// Continue with your logic...
}

/*
func (f *FollowersHandler) GetUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["userId"]
	user, err := f.repo.ReadUser(id)
	if err != nil {
		f.logger.Print("Database exception: ", err)
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		f.logger.Fatal("Unable to convert to json :", err)
		return
	}
}
*/

func (f *FollowersHandler) GetFollowingsForUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["userId"]
	users, err := f.repo.GetFollowingsForUser(id)
	if err != nil {
		f.logger.Print("Database exception: ", err)
	}
	if users == nil {
		users = model.Users{}
		jsonData, _ := json.Marshal(users)
		rw.Write(jsonData)
		return
	}
	err = users.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		f.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (f *FollowersHandler) GetFollowersForUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["userId"]
	users, err := f.repo.GetFollowersForUser(id)
	if err != nil {
		f.logger.Print("Database exception: ", err)
	}
	if users == nil {
		users = model.Users{}
		jsonData, _ := json.Marshal(users)
		rw.Write(jsonData)
		return
	}
	err = users.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		f.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (f *FollowersHandler) GetRecommendationsForUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["userId"]
	users, err := f.repo.GetRecommendationsForUser(id)
	if err != nil {
		f.logger.Print("Database exception: ", err)
	}
	if users == nil {
		users = model.Users{}
		jsonData, _ := json.Marshal(users)
		rw.Write(jsonData)
		return
	}
	err = users.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		f.logger.Fatal("Unable to convert to json :", err)
		return
	}
}
