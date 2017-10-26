package api

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mrevilme/promping/config"
	"io"
	"net/http"
	"strings"
)

var Router *mux.Router

func init() {
	Router = mux.NewRouter().PathPrefix("/api/").Subrouter()
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {

		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		qt, err := route.GetQueriesTemplates()
		// p will contain regular expression is compatible with regular expression in Perl, Python, and other languages.
		// for instance the regular expression for path '/articles/{id}' will be '^/articles/(?P<v0>[^/]+)$'
		p, err := route.GetPathRegexp()
		if err != nil {
			return err
		}
		// qr will contain a list of regular expressions with the same semantics as GetPathRegexp,
		// just applied to the Queries pairs instead, e.g., 'Queries("surname", "{surname}") will return
		// {"^surname=(?P<v0>.*)$}. Where each combined query pair will have an entry in the list.
		qr, err := route.GetQueriesRegexp()
		m, err := route.GetMethods()
		if err != nil {
			return err
		}

		io.WriteString(w, fmt.Sprintf("%s %s %s %s %s\n", strings.Join(m, ","), strings.Join(qt, ","), strings.Join(qr, ","), t, p))
		return nil
	})
}
func Run() {
	Router.HandleFunc("/", catchAllHandler).Methods("GET")
	http.Handle("/api/", Router)
}
