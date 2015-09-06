package go2md

import (
	_ "golang.org/x/tools/playground"
	"golang.org/x/tools/present"
	"html/template"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/pborman/uuid"
)

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
		Title:     "EmptyTitle",
		SubTitle:  "EmptySubTitle",
		SpeakDate: "1 Sep 2015",
		Tags:      "golang,present",
		Markdown:  "* Page 1",
	}

	// add empty slide data
	key, _ := datastore.Put(c, datastore.NewKey(c, "Slide", uuid.New(), 0, nil), &slide)
	http.Redirect(w, r, "/me/slide/edit/"+key.StringID(), 301)
}

func putSlide(r *http.Request, key string) (*Slide, error) {
	c := appengine.NewContext(r)
	r.ParseForm()

	slide := Slide{
		UserKey:   r.FormValue("UserKey"),
		Title:     r.FormValue("Title"),
		SubTitle:  r.FormValue("SubTitle"),
		SpeakDate: r.FormValue("SpeakDate"),
		Tags:      r.FormValue("Tags"),
		Markdown:  r.FormValue("Markdown"),
	}
	datastore.Put(c, datastore.NewKey(c, "Slide", key, 0, nil), &slide)
	return &slide, nil
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

func editHandler(w http.ResponseWriter, r *http.Request) {

	urls := strings.Split(r.URL.Path, "/")
	keyId := urls[4]

	var s *Slide
	if r.Method == "POST" {
		s, _ = putSlide(r, keyId)
	} else {
		s, _ = getSlide(r, keyId)
	}
	rtn := struct {
		Key  string
		Data *Slide
	}{keyId, s}
	meRender(w, "./templates/me/edit.tmpl", rtn)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	urls := strings.Split(r.URL.Path, "/")
	keyId := urls[4]

	u, _ := getUser(r)
	s, _ := getSlide(r, keyId)

	data := Who{
		author: "secondarykey",
		id:     "1",
	}

	slideTxt := ""
	slideTxt += s.Title + "\n"
	slideTxt += s.SubTitle + "\n"
	slideTxt += s.SpeakDate + "\n"
	slideTxt += "Tags:" + s.Tags + "\n"
	slideTxt += "\n"
	slideTxt += u.Name + "\n"
	slideTxt += u.Job + "\n"
	slideTxt += u.Url + "\n"
	slideTxt += "@" + u.TwitterId + "\n"
	slideTxt += "\n"
	slideTxt += s.Markdown

	//52md
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

	ctx := present.Context{ReadFile: data.AttributeFile}
	reader := strings.NewReader(slideTxt)
	doc, err := ctx.Parse(reader, "tour.slide", 0)
	if err != nil {
		panic(err)
	}

	tmpl, err := createTemplate()
	if err != nil {
		panic(err)
	}

	//doc.Render(w, tmpl)
	rtn := struct {
		*present.Doc
		Template    *template.Template
		PlayEnabled bool
		LastWord    string
	}{doc, tmpl, true, u.LastWord}
	err = tmpl.ExecuteTemplate(w, "root", rtn)
	if err != nil {
		panic(err)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	urls := strings.Split(r.URL.Path, "/")
	keyId := urls[4]

	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "Slide", keyId, 0, nil)

	//err
	datastore.Delete(c, k)

	http.Redirect(w, r, "/me/", 301)
	return
}
