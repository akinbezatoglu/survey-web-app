package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/akinbezatoglu/survey-web-app/pkg/api/response"
	"github.com/akinbezatoglu/survey-web-app/pkg/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
)

//	api/v1/auth	GET
// 	Checks if the user is authenticated
func (a *API) UserValidate(w http.ResponseWriter, r *http.Request) {
	var u model.User
	err := mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user, err := a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	resp := &model.ViewUser{
		ID:       user.ID.Hex(),
		Name:     user.Name,
		Lastname: user.Lastname,
		Forms:    user.Forms,
	}
	response.Write(w, r, resp)
}

// 	api/v1/auth/login	POST
// 	Handles user login
func (a *API) userLoginPostHandler(w http.ResponseWriter, r *http.Request) {
	// Validate user input
	var u model.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if u.Email == "" {
		response.Errorf(w, r, nil, http.StatusBadRequest, "Email address is missing")
		return
	} else if u.Password == "" {
		response.Errorf(w, r, nil, http.StatusBadRequest, "Password is missing")
		return
	}
	// Always use the lower case email address
	u.Email = strings.ToLower(u.Email)
	// Get the user database entry
	user, err := a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	} else if user == nil {
		response.Errorf(w, r, err, http.StatusBadRequest, "Invalid email address or password")
		return
	}
	// Check the password
	if !user.MatchPassword(u.Password) {
		response.Errorf(w, r, err, http.StatusBadRequest, "Invalid email address or password")
		return
	}
	// Create jwt token
	user.Token, err = a.createJWT(jwt.MapClaims{
		"email":    user.Email,
		"name":     user.Name,
		"lastname": user.Lastname,
		"password": user.Password,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	//	http.SetCookie(w, &http.Cookie{
	//		Name:  "token",
	//		Value: u.Token,
	//		Path:  "/",
	//	})
	response.Write(w, r, user)
}

// 	api/v1/auth/signup	POST
// 	Handles user registration
func (a *API) userSignupPostHandler(w http.ResponseWriter, r *http.Request) {
	// Validate user input
	var u model.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	} else if u.Name == "" {
		response.Errorf(w, r, nil, http.StatusBadRequest, "Name is missing")
		return
	} else if u.Lastname == "" {
		response.Errorf(w, r, nil, http.StatusBadRequest, "Lastname is missing")
		return
	} else if u.Email == "" {
		response.Errorf(w, r, nil, http.StatusBadRequest, "Email address is missing")
		return
	} else if u.Password == "" {
		response.Errorf(w, r, nil, http.StatusBadRequest, "Password is missing")
		return
	}
	// Always use the lower case email address
	u.Email = strings.ToLower(u.Email)

	// Create jwt token
	u.Token, err = a.createJWT(jwt.MapClaims{
		"email":    u.Email,
		"name":     u.Name,
		"lastname": u.Lastname,
		"password": u.Password,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Hash the user password
	err = u.HashPassword()
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	// Save user to database
	InsertedUserID, err := a.db.CreateUser(&u)
	if err != nil {
		if err.Error() == "email_address_already_exists" {
			response.Errorf(w, r, err, http.StatusBadRequest, "Email address already exists")
			return
		}
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	returnUser := model.ViewUser{
		ID:       InsertedUserID,
		Name:     u.Name,
		Lastname: u.Lastname,
		Email:    u.Email,
		Password: u.Password,
		Token:    u.Token,
	}
	//	http.SetCookie(w, &http.Cookie{
	//		Name:  "token",
	//		Value: u.Token,
	//		Path:  "/",
	//	})

	// Omit password
	//	u.ID, err = primitive.ObjectIDFromHex(InsertedUserID)
	//	if err != nil {
	//		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
	//		return
	//	}
	//	u.Password = ""
	response.Write(w, r, returnUser)
}

// 	api/v1/profile/{userid}	GET
// 	Retrieves the user's information
func (a *API) userProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/profile/{userid}
	//userid := strings.Split(r.URL.Path, "/")[4]
	var u *model.User
	err := mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	//	user, err := a.db.GetUser(userid)
	//	if err != nil {
	//		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
	//		return
	//	}
	response.Write(w, r, u)
}

// 	api/v1/profile/{userid}	PUT
// 	Updates the user's information
func (a *API) userProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/profile/{userid}
	userid := strings.Split(r.URL.Path, "/")[4]
	var u *model.User
	err := mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if u.ID.Hex() == userid {
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		err = a.db.SaveUser(u)
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		response.Write(w, r, u)
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}
