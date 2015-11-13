package gopredit

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

func init() {
	http.HandleFunc("/me/file/view", fileViewHandler)
	http.HandleFunc("/me/file/upload", uploadHandler)
}

type File struct {
	UserKey string
	Data    []byte
}

func fileViewHandler(w http.ResponseWriter, r *http.Request) {

	rtn, _ := getFileKey(r)

	tmpl, err := template.ParseFiles("./templates/me/file.tmpl")
	if err != nil {
		return
	}
	err = tmpl.Execute(w, rtn)
	if err != nil {
		return
	}
}

//change ajax access
func uploadHandler(w http.ResponseWriter, r *http.Request) {

	//ducapple name

	name := r.FormValue("fileName")
	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		//add error handling
		return
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
	}

	c := appengine.NewContext(r)

	// get user data
	u, err := getUser(r)
	if err != nil {
	}

	key := datastore.NewKey(c, "File", u.UserKey+"/"+name, 0, nil)
	f := File{
		UserKey: u.UserKey,
		Data:    b,
	}
	// add empty slide data
	_, err = datastore.Put(c, key, &f)
	if err != nil {
	}

	http.Redirect(w, r, "/me/file/view", 301)
}

func getFile(r *http.Request, name string) (*File, error) {
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, "File", name, 0, nil)
	rtn := File{}

	if err := datastore.Get(c, key, &rtn); err != nil {
		if err != datastore.ErrNoSuchEntity {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &rtn, nil
}

func getFileKey(r *http.Request) ([]string, error) {
	c := appengine.NewContext(r)
	u, err := getUser(r)
	if err != nil {
	}

	userKey := u.UserKey
	q := datastore.NewQuery("File").KeysOnly().Filter("UserKey = ", userKey)
	keys, err := q.GetAll(c, nil)
	if err != nil {
	}

	rtn := make([]string, len(keys))
	for idx, elm := range keys {
		rtn[idx] = strings.Replace(elm.StringID(), userKey, "file", 1)
	}
	return rtn, nil

}
