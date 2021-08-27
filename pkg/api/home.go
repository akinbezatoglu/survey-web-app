package api

import (
	"net/http"
	"strings"

	"github.com/akinbezatoglu/survey-web-app/pkg/api/response"
	"github.com/akinbezatoglu/survey-web-app/pkg/model"
)

//	api/v1/f/{userid}	GET
// 	Gets all forms that belong to the user.
func (a *API) GetAllFormsHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/f/{userid}
	userid := strings.Split(r.URL.Path, "/")[4]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var f []model.Form
	form_num := len(user.Forms)
	for i := 0; i < form_num; i++ {
		form, err := a.db.GetForm(user.Forms[i])
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		f = append(f, *form)
	}
	response.Write(w, r, f)
}

// 	api/v1/f/{userid}	POST
//	Creates a form that belongs to the user.
func (a *API) CreateFormHandler(w http.ResponseWriter, r *http.Request) {
	// /api/v1/f/{userid}
	userid := strings.Split(r.URL.Path, "/")[4]
	u, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var f model.Form
	f.Name = "Adsız Başlık"
	f.Desc = "Adsız açıklama"
	f.IsVisible = true

	var q model.Question
	q.Ques = "Soru"
	q.Type = 1 // 1: çoktan seçmeli // 2:Paragraf // 3: kutu işaretlemeli
	q.No = 1
	q.IsRequired = false

	var o model.Option
	o.Opt = "Seçenek"
	o.IsTrue = false
	o.Number = 1

	q.Options = append(q.Options, o)

	ques_id, err := a.db.CreateQuestion(&q)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	f.Questions = append(f.Questions, ques_id)
	formID, err := a.db.CreateForm(&f)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	u.Forms = append(u.Forms, formID)
	err = a.db.SaveUser(u)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	response.Write(w, r, formID)
}

//	api/v1/f/{userid}/less	GET
func (a *API) GetAllFormsWithLessValueHandler(w http.ResponseWriter, r *http.Request) {
	userid := strings.Split(r.URL.Path, "/")[4]
	user, err := a.db.GetUser(userid)
	if err != nil {
		response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	var f []model.FormSmall
	form_num := len(user.Forms)
	for i := 0; i < form_num; i++ {
		form, err := a.db.GetForm(user.Forms[i])
		if err != nil {
			response.Errorf(w, r, err, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		f_sm := &model.FormSmall{
			ID:   form.ID.Hex(),
			Name: form.Name,
		}
		f = append(f, *f_sm)
	}
	response.Write(w, r, f)
}
