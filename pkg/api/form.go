package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/akinbezatoglu/survey-web-app/pkg/api/response"
	"github.com/akinbezatoglu/survey-web-app/pkg/model"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//	api/v1/f/{formid}/view	GET
//	Handles surveys being displayed on a page where they can be answered.
func (a *API) formViewGetHandler(w http.ResponseWriter, r *http.Request) {
	// https://gist.github.com/17twenty/c815680c9c585cd9c16e62cbee7317b6
	ipAddress := r.RemoteAddr
	fwdAddress := r.Header.Get("X-Forwarded-For") // capitalisation doesn't matter
	if fwdAddress != "" {
		// Got X-Forwarded-For
		ipAddress = fwdAddress // If it's a single IP, then awesome!

		// If we got an array... grab the first IP
		ips := strings.Split(fwdAddress, ", ")
		if len(ips) > 1 {
			ipAddress = ips[0]
		}
	}
	////////////////////////////////////////////////////////////////////////

	formid := strings.Split(r.URL.Path, "/")[4]
	f, err := a.db.GetForm(formid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if len(f.BlockList) != 0 {
		for i := 0; i < len(f.BlockList); i++ {
			if f.BlockList[i] == ipAddress {
				response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
				return
			}
		}
	}
	var q []model.Question
	for i := 0; i < len(f.Questions); i++ {
		ques, err := a.db.GetQuestion(f.Questions[i])
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		q = append(q, *ques)
	}
	var viewForm model.ViewForm
	mapstructure.Decode(f, &viewForm)
	viewForm.Questions = q
	response.Write(w, r, viewForm)
}

//	api/v1/f/{formid}/view	POST
// 	Handles survey responses
func (a *API) formViewPostHandler(w http.ResponseWriter, r *http.Request) {
	// https://gist.github.com/17twenty/c815680c9c585cd9c16e62cbee7317b6
	ipAddress := r.RemoteAddr
	fwdAddress := r.Header.Get("X-Forwarded-For") // capitalisation doesn't matter
	if fwdAddress != "" {
		// Got X-Forwarded-For
		ipAddress = fwdAddress // If it's a single IP, then awesome!

		// If we got an array... grab the first IP
		ips := strings.Split(fwdAddress, ", ")
		if len(ips) > 1 {
			ipAddress = ips[0]
		}
	}
	////////////////////////////////////////////////////////////////////////
	formid := strings.Split(r.URL.Path, "/")[4]
	f, err := a.db.GetForm(formid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	f.BlockList = append(f.BlockList, ipAddress)

	var allAnswers []model.Answer
	err = json.NewDecoder(r.Body).Decode(&allAnswers)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internall Server Error")
		return
	}
	for i := 0; i < len(allAnswers); i++ {
		q, err := a.db.GetQuestion(allAnswers[i].QuesID)
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internalll Server Error")
			return
		}
		answer := allAnswers[i]
		ansID, err := a.db.CreateAnswer(&answer)
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internallll Server Error")
			return
		}
		q.Answers = append(q.Answers, ansID)
		err = a.db.SaveQuestion(q)
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internalllll Server Error")
			return
		}
	}
	response.Write(w, r, nil)
}

//	api/v1/f/{userid}/{formid} GET
//	Handles sending the user the survey that the user wants to edit.
func (a *API) formGetHandler(w http.ResponseWriter, r *http.Request) {
	// api/v1/f/{userid}/{formid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			f, err := a.db.GetForm(formid)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			var q []model.Question
			if len(f.Questions) != 0 {
				for i := 0; i < len(f.Questions); i++ {
					ques, err := a.db.GetQuestion(f.Questions[i])
					if err != nil {
						response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
						return
					}
					q = append(q, *ques)
				}
			}
			var viewForm model.ViewForm
			mapstructure.Decode(f, &viewForm)
			viewForm.Questions = q
			response.Write(w, r, viewForm)
		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/f/{userid}/{formid} GET
// 	Handles the changes that the users will have made in their surveys.
func (a *API) formPutHandler(w http.ResponseWriter, r *http.Request) {
	// api/v1/f/{userid}/{formid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			form, err := a.db.GetForm(formid)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//var f *model.Form
			err = json.NewDecoder(r.Body).Decode(&form)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//mapstructure.Decode(f, &form)
			err = a.db.SaveForm(form)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response.Write(w, r, nil)
		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/f/{userid}/{formid} GET
// 	Handles users deleting their surveys.
func (a *API) formDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// api/v1/f/{userid}/{formid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			succes := false
			for i := 0; i < len(user.Forms); i++ {
				if user.Forms[i] == formid {
					fID, err := primitive.ObjectIDFromHex(formid)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					err = a.db.DeleteForm(fID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					// Remove the element at index i from a.
					copy(user.Forms[i:], user.Forms[i+1:])      // Shift a[i+1:] left one index.
					user.Forms[len(user.Forms)-1] = ""          // Erase last element (write zero value).
					user.Forms = user.Forms[:len(user.Forms)-1] // Truncate slice.
					err = a.db.SaveUser(user)
					if err != nil {
						response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
						return
					}
					succes = true
					response.Write(w, r, nil)
					return
				}
			}
			if !succes {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/f/{userid}/{formid}/response	GET
// 	Handles the display of survey responses.
func (a *API) formResponseGetHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/f/{userid}/{formid}/response
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	for _, v := range user.Forms {
		if v == formid {
			owner = true
			break
		}
	}
	if owner {
		f, err := a.db.GetForm(formid)
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		var res []model.ViewResponse
		for i := 0; i < len(f.Questions); i++ {

			ques, err := a.db.GetQuestion(f.Questions[i])
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			var responseOfques model.ViewResponse
			responseOfques.Ques = ques.Ques
			for j := 0; j < len(ques.Answers); j++ {
				ans, err := a.db.GetAnswer(ques.Answers[j])
				if err != nil {
					response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
					return
				}
				responseOfques.Opt = append(responseOfques.Opt, ans.Opt)
			}
			res = append(res, responseOfques)
		}

		//	var q []model.Question
		//	for i := 0; i < len(f.Questions); i++ {
		//		ques, err := a.db.GetQuestion(f.Questions[i])
		//		if err != nil {
		//			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		//			return
		//		}
		//		var AllOpt []string
		//		for j := 0; j < len(ques.Answers); j++ {
		//			ans, err := a.db.GetAnswer(ques.Answers[j])
		//			if err != nil {
		//				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		//				return
		//			}
		//			AllOpt = append(AllOpt, ans.Opt)
		//		}
		//		ques.Answers = AllOpt
		//
		//		q = append(q, *ques)
		//	}
		response.Write(w, r, res)
	} else {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
	}
}

// 	api/v1/q/{userid}/{formid}
//	Handles creating new questions on the survey.
func (a *API) questionPostHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/q/{userid}/{formid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			f, err := a.db.GetForm(formid)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			numberOfQuestions := len(f.Questions)
			var q model.Question
			// Define question type
			//	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &q)
			//
			//	if err != nil {
			//		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
			//		return
			//	}

			err = json.NewDecoder(r.Body).Decode(&q)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}

			q.Ques = "Soru"
			q.No = numberOfQuestions + 1
			var o model.Option
			if q.Type == 1 {
				o.Opt = "Seçenek"

			} else if q.Type == 3 {
				o.Opt = "İşaretleyiniz"
			}

			o.IsTrue = false
			o.Number = 1

			q.Options = append(q.Options, o)
			qID, err := a.db.CreateQuestion(&q)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			f.Questions = append(f.Questions, qID)
			err = a.db.SaveForm(f)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			response.Write(w, r, qID)

		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

// 	api/v1/q/{userid}/{formid}/copy
//	Handles creating new questions on the survey.
func (a *API) questionPostCopyHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/q/{userid}/{formid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			f, err := a.db.GetForm(formid)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			numberOfQuestions := len(f.Questions)
			var q model.Question
			err = json.NewDecoder(r.Body).Decode(&q)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			q.No = numberOfQuestions + 1
			qID, err := a.db.CreateQuestion(&q)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			f.Questions = append(f.Questions, qID)
			err = a.db.SaveForm(f)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			response.Write(w, r, qID)
		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

// 	api/v1/q/{userid}/{formid}/{questionid}
//	Handles updating the question on the survey.
func (a *API) questionPutHandler(w http.ResponseWriter, r *http.Request) {
	//	/api/v1/q/{userid}/{formid}/{questionid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	qID := strings.Split(r.URL.Path, "/")[6]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			q, err := a.db.GetQuestion(qID)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			err = json.NewDecoder(r.Body).Decode(&q)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			err = a.db.SaveQuestion(q)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			response.Write(w, r, nil)

		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/q/{userid}/{formid}/{questionid}
//	Handles deleting the question on the survey.
func (a *API) questionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// api/v1/q/{userid}/{formid}/{questionid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	qID := strings.Split(r.URL.Path, "/")[6]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			quesID, err := primitive.ObjectIDFromHex(qID)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internallll Server Error")
				return
			}
			f, err := a.db.GetForm(formid)
			if err != nil {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internalll Server Error")
				return
			}
			succes := false
			for i := 0; i < len(f.Questions); i++ {
				if f.Questions[i] == qID {
					err = a.db.DeleteQuestion(quesID)
					if err != nil {
						response.Errorf(w, r, err, http.StatusInternalServerError, "Internall Server Error")
						return
					}
					// Remove the element at index i from a.
					copy(f.Questions[i:], f.Questions[i+1:])       // Shift a[i+1:] left one index.
					f.Questions[len(f.Questions)-1] = ""           // Erase last element (write zero value).
					f.Questions = f.Questions[:len(f.Questions)-1] // Truncate slice.
					err = a.db.SaveForm(f)
					if err != nil {
						response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
						return
					}
					for k := i; k < len(f.Questions); k++ {
						qq, err := a.db.GetQuestion(f.Questions[k])
						if err != nil {
							response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
							return
						}
						qq.No -= 1
						err = a.db.SaveQuestion(qq)
						if err != nil {
							response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
							return
						}
					}

					succes = true
					response.Write(w, r, nil)
				}
			}
			if !succes {
				response.Errorf(w, r, err, http.StatusInternalServerError, "Internaqqqlll Server Error")
			}
		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/o/{userid}/{formid}/{questionid}
//	Handles creating the option on the question on the survey.
func (a *API) optionPostHandler(w http.ResponseWriter, r *http.Request) {
	// api/v1/o/{userid}/{formid}/{questionid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	qID := strings.Split(r.URL.Path, "/")[6]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			q, err := a.db.GetQuestion(qID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if q.Type == 1 {
				newOpt := model.Option{
					Number: len(q.Options) + 1,
					Opt:    "Seçenek",
					IsTrue: false,
				}
				q.Options = append(q.Options, newOpt)
				err = a.db.SaveQuestion(q)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				response.Write(w, r, nil)
			} else if q.Type == 3 {
				newOpt := model.Option{
					Number: len(q.Options) + 1,
					Opt:    "İşaretleyiniz",
					IsTrue: false,
				}
				q.Options = append(q.Options, newOpt)
				err = a.db.SaveQuestion(q)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				response.Write(w, r, nil)
			} else {
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/o/{userid}/{formid}/{questionid}/{optionid}
//	Handles updating the option on the question on the survey.
func (a *API) optionPutHandler(w http.ResponseWriter, r *http.Request) {
	// api/v1/o/{userid}/{formid}/{questionid}/{optionid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	qID := strings.Split(r.URL.Path, "/")[6]
	oID := strings.Split(r.URL.Path, "/")[7]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			oID_int, err := strconv.Atoi(oID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			q, err := a.db.GetQuestion(qID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = json.NewDecoder(r.Body).Decode(&q.Options[oID_int-1])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = a.db.SaveQuestion(q)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response.Write(w, r, nil)

		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}

//	api/v1/o/{userid}/{formid}/{questionid}/{optionid}
//	Handles deleting the option on the question on the survey.
func (a *API) optionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/o/{userid}/{formid}/{questionid}/{optionid}
	owner := false
	userid := strings.Split(r.URL.Path, "/")[4]
	formid := strings.Split(r.URL.Path, "/")[5]
	qID := strings.Split(r.URL.Path, "/")[6]
	oID := strings.Split(r.URL.Path, "/")[7]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var u *model.User
	err = mapstructure.Decode(r.Context().Value(ContextJWTKey), &u)
	if err != nil {
		response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}
	u, err = a.db.GetUserByEmail(u.Email)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if u.ID == user.ID {
		for _, v := range user.Forms {
			if v == formid {
				owner = true
				break
			}
		}
		if owner {
			oID_int, err := strconv.Atoi(oID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			q, err := a.db.GetQuestion(qID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			deleteIndex := oID_int - 1
			// Remove the element at index i from a.
			copy(q.Options[deleteIndex:], q.Options[deleteIndex+1:]) // Shift a[i+1:] left one index.
			q.Options[len(q.Options)-1] = model.Option{}             // Erase last element (write zero value).
			q.Options = q.Options[:len(q.Options)-1]                 // Truncate slice.
			for i := deleteIndex; i < len(q.Options); i++ {
				q.Options[i].Number -= 1
			}
			err = a.db.SaveQuestion(q)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response.Write(w, r, nil)

		} else {
			response.Errorf(w, r, nil, http.StatusUnauthorized, "Unauthorized")
		}
	} else {
		response.Errorf(w, r, nil, http.StatusForbidden, "Forbidden")
	}
}
