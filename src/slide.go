package gopredit

import (
	_ "golang.org/x/tools/playground"
	"golang.org/x/tools/present"
	"html/template"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pborman/uuid"

	"bufio"
	"bytes"
	"time"
)

type Slide struct {
	UserKey   string
	Title     string
	SubTitle  string
	SpeakDate string
	Tags      string
	Markdown  string
	Date      time.Time
}

func init() {
	http.HandleFunc("/me/slide/create", createHandler)
	http.HandleFunc("/me/slide/edit/", editHandler)
	http.HandleFunc("/me/slide/view/", viewHandler)
	http.HandleFunc("/me/slide/delete/", deleteHandler)
}

func createHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	// get user data
	u, _ := getUser(r)
	slide := Slide{
		UserKey:   u.UserKey,
		Title:     "Title",
		SubTitle:  "SubTitle",
		SpeakDate: time.Now().Format("_2 Jan 2006"),
		Tags:      "golang",
		Markdown:  "* Welcome GoPreEdit",
	}

	k := datastore.NewKey(c, "Slide", uuid.New(), 0, nil)
	// add empty slide data
	key, err := datastore.Put(c, k, &slide)
	if err != nil {
		errorPage(w, "InternalServerError", err.Error(), "Put Error", 500)
	} else {
		http.Redirect(w, r, "/me/slide/edit/"+key.StringID(), 301)
	}

}

func createFormSlide(r *http.Request) (*Slide, error) {
	r.ParseForm()
	slide := Slide{
		UserKey:   r.FormValue("UserKey"),
		Title:     r.FormValue("Title"),
		SubTitle:  r.FormValue("SubTitle"),
		SpeakDate: r.FormValue("SpeakDate"),
		Tags:      r.FormValue("Tags"),
		Markdown:  r.FormValue("Markdown"),
		Date:      time.Now(),
	}
	return &slide, nil
}

func putSlide(r *http.Request, key string) (*Slide, error) {
	c := appengine.NewContext(r)
	slide, err := createFormSlide(r)
	if err != nil {
		return nil, err
	}

	k := createKey(c, "Slide", key)

	_, err = datastore.Put(c, k, slide)
	if err != nil {
		return nil, err
	}
	return slide, nil
}

func getSlide(r *http.Request, key string) (*Slide, error) {

	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "Slide", key, 0, nil)
	rtn := Slide{}
	if err := datastore.Get(c, k, &rtn); err != nil {
		if err != datastore.ErrNoSuchEntity {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &rtn, nil
}

// Slide EditView
func editHandler(w http.ResponseWriter, r *http.Request) {

	urls := strings.Split(r.URL.Path, "/")
	keyId := urls[4]

	var s *Slide
	var err error
	if r.Method == "POST" {
		s, err = putSlide(r, keyId)
	} else {
		s, err = getSlide(r, keyId)
	}

	if err != nil {
		errorPage(w, "Slide Error", err.Error(), "Get ? Put ?", 500)
		return
	}

	rtn := struct {
		Key  string
		Data *Slide
	}{keyId, s}

	meRender(w, "./templates/me/edit.tmpl", rtn)
}

// Slide View
func viewHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	u, err := getUser(r)
	if err != nil {
		errorPage(w, "Not Found", "User Not Found", err.Error(), 404)
		return
	}

	urls := strings.Split(r.URL.Path, "/")
	keyId := urls[4]
	s, err := getSlide(r, keyId)
	if err != nil {
		errorPage(w, "Slide Error", "Slide Get", err.Error(), 404)
		return
	}

	data := Who{
		author:  u.UserKey,
		request: r,
	}

	b, err := createSlide(u, s, &data)
	if err != nil {
		log.Infof(c, err.Error())
		errorPage(w, "Slide Error", "Slide Create", err.Error(), 500)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		errorPage(w, "Slide Error", "Slide Write", err.Error(), 500)
	}
}

func addLine(orgData, data, prefix string) string {
	if data != "" {
		if prefix != "" {
			orgData += prefix + data + "\n"
		} else {
			orgData += data + "\n"
		}
	}
	return orgData
}

func createSlide(u *User, s *Slide, w *Who) ([]byte, error) {

	//c := appengine.NewContext(w.request)

	// create space data
	slideTxt := ""
	slideTxt = addLine(slideTxt, s.Title, "")
	slideTxt = addLine(slideTxt, s.SubTitle, "")
	slideTxt = addLine(slideTxt, s.SpeakDate, "")
	slideTxt = addLine(slideTxt, s.Tags, "Tags:")
	slideTxt += "\n"

	slideTxt = addLine(slideTxt, u.Name, "")
	slideTxt = addLine(slideTxt, u.Job, "")
	slideTxt = addLine(slideTxt, u.Url, "")
	slideTxt = addLine(slideTxt, u.TwitterId, "@")

	slideTxt += "\n"
	slideTxt += s.Markdown

	//
	//Golang Present Tools Editor
	//15 Aug 2015
	//Tags: golang shizuoka_go
	//
	//secondarykey
	//Programer
	//http://github.com/shizuokago/52md
	//@secondarykey
	//
	//* This Service Alpha

	ctx := present.Context{ReadFile: w.AttributeFile}
	reader := strings.NewReader(slideTxt)
	doc, err := ctx.Parse(reader, "tour.slide", 0)
	if err != nil {
		return nil, err
	}

	tmpl, err := createTemplate()
	if err != nil {
		return nil, err
	}

	if u.LastWord == "" {
		u.LastWord = "Thank you"
	}

	rtn := struct {
		*present.Doc
		Template    *template.Template
		PlayEnabled bool
		LastWord    string
	}{doc, tmpl, true, u.LastWord}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err = tmpl.ExecuteTemplate(writer, "root", rtn)
	if err != nil {
		return nil, err
	}
	writer.Flush()

	return b.Bytes(), nil
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	urls := strings.Split(r.URL.Path, "/")
	keyId := urls[4]

	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "Slide", keyId, 0, nil)

	//err
	err := datastore.Delete(c, k)
	if err != nil {
		errorPage(w, "Delete Error", "Slide Delete", err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/me/", 301)
	return
}
