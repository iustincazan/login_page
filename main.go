package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Alert struct {
	Level   string
	Message string
}

type loginHandler struct {
	next http.Handler
}

type home struct{}

type authenticated_user struct {
	Name string
}

type loggedInUser string

func main() {

	// we must create an http FileServer to serve static files, such as images, js and styles
	// the webserver path is Dir path + Route patch
	// in this case ./static + /
	// so the files will need to be in the static folder in the project path
	dir := http.Dir("./static")

	fs := http.FileServer(dir)

	mux := http.NewServeMux()

	mux.Handle("/", fs)

	var l loginHandler
	var h home
	var au authenticated_user
	// setting the next chained handler as authenticated_user, to account for the case when auth was successful
	l.next = au

	mux.Handle("/login", l)
	mux.Handle("/home", h)

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}

}

func (l loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		// Parse the form and perform authentication logic

		au, err := auth(r.FormValue("email"), r.FormValue("password"))
		if err != nil {

			alert := Alert{
				Level:   "danger",
				Message: err.Error(),
			}
			// display login page
			t, err := template.ParseFiles("templates/login.gohtml", "templates/alert.gohtml")

			if err != nil {
				panic(err)
			}

			err = t.Execute(w, alert)

			if err != nil {
				panic(err)
			}

			w.Header().Add("Content-Type", "application/x-www-form-urlencoded")
			return

		}

		// if credentials are good, redirect to home page
		// passing the user name as a context key
		// key type is defined as a custom string type, to avoid type conflicts
		c := context.WithValue(r.Context(), loggedInUser("user"), au.Name)
		r := r.WithContext(c)
		l.next.ServeHTTP(w, r)

	} else if r.Method == "GET" {
		// display login page
		t, err := template.ParseFiles("templates/login.gohtml", "templates/alert.gohtml")

		if err != nil {
			panic(err)
		}

		err = t.Execute(w, nil)

		if err != nil {
			panic(err)
		}

		w.Header().Add("Content-Type", "application/x-www-form-urlencoded")

	} else {
		// display "Unsupported Method HTTP error"
		var err string = fmt.Sprintf("%s method is not supported", r.Method)
		http.Error(w, err, http.StatusMethodNotAllowed)
	}

}

func (h home) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/home.gohtml")

	if err != nil {
		panic(err)
	}

	err = t.Execute(w, nil)

	if err != nil {
		panic(err)
	}

}

func (au authenticated_user) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/authenticated_user.gohtml")

	if err != nil {
		panic(err)
	}

	type currentUser struct {
		Name string
	}

	var cu currentUser
	cu.Name = r.Context().Value(loggedInUser("user")).(string)

	err = t.Execute(w, cu)

	if err != nil {
		panic(err)
	}

}

func auth(email string, password string) (authenticated_user, error) {

	if email == "jdoe@mail.com" && password == "correctpassword" {
		return authenticated_user{Name: "Iustin"}, nil
	}
	return authenticated_user{}, errors.New("No match found!")
}
