package main 

import (
	"net/http"
	"log"
	"strings"

	"fmt"
)

type LoginMessage struct {
		Result		string	`json:"result"`
		Message		string	`json:"message"`
		User		string	`json:"user,omitempty"`
		Role		string	`json:"role,omitempty"`
		GoTo		string	`json:"goto,omitempty"`
}

var	messages = []LoginMessage{
	{
		Result: "ok",
		Message: "",
		User: "",
		Role: "Administrator",
		GoTo: "/app",
	},
	{
		Result: "failed",
		Message: "",
	},
}


func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("NOT FOUND", r.Method,r.URL.Path, r.RemoteAddr)
	http.Error(w, "Not Found", http.StatusNotFound)	
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("METHOD NOT ALLOWED", r.Method,r.URL.Path, r.RemoteAddr)
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func unauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UNAUTHORIZED", r.Method,r.URL.Path, r.RemoteAddr)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func internalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("INTERNAL SERVER ERROR", r.Method, r.URL.Path, r.RemoteAddr)
	// Dump Api Request r.Context()
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	} else if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		notFoundHandler(w, r)
		return
	}

	http.ServeFile(w, r, "./static/index.html")
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	}

	http.ServeFile(w, r, "./static/styles.css")
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	}

	http.ServeFile(w, r, "./static/script.js")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowedHandler(w, r)
		return
	} else if err := r.ParseForm(); err != nil {
		log.Printf("loginHandler r.ParseForm() error: %v\n", err)
		log.Printf("r.Form : %+v\n", r.Form)
		log.Printf("r.PostForm : %+v\n", r.PostForm)
		internalServerErrorHandler(w, r)
		return
	}

	postFormValues := r.Form

	var username string
	val := postFormValues["username"]
	if len(val) > 0 {
		username = val[0]
	}

	var password string
	val = postFormValues["password"]
	if len(val) > 0 {
		password = val[0]
	}

	var message LoginMessage
	if authOk, err := AuthenticateUser(w, r, username, password); !authOk || err != nil {
		log.Printf("loginHandler AuthenticateUser error: %v\n", err)
		message = messages[1]

	} else {
		log.Printf("loginHandler User Authenticated: %s | remoteAddr: %s\n", username, r.RemoteAddr)
		message = messages[0]
	}

	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	sendJson(w, r, message)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	}

	RemoveAuthentication(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
	
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	http.ServeFile(w, r, "./static/appindex.html")
}

func appCssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	}

	http.ServeFile(w, r, "./static/appstyles.css")
}

func appJsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w, r)
		return
	}

	http.ServeFile(w, r, "./static/appscript.js")
}

func appApiHandler(w http.ResponseWriter, r *http.Request) {
	apiReq := ApiRequest {
		Aud: &AuthUserData{},
		W: w,
		R: r,
		Method: r.Method,
	}

	aud, ok := r.Context().Value("aud").(*AuthUserData) 
	if ok {
		apiReq.Aud = aud
	} else {
		fmt.Printf("aud error: %+v\n", *aud)
		internalServerErrorHandler(w, r)
		return
	}

	// Build API Request from HTTP Request
	path := r.URL.Path
	if len(path) > 0 {
		if path[0] == '/' {
			path = path[1:]
		}
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	}

	pathElements := strings.Split(path, "/")
	// app			- 0
	// api			- 1
	// apiType		- 2
	// groupName	- 3
	// objectName	- 4
	if len(pathElements) < 3 { // NOT FOUND if apiType not present
		notFoundHandler(w, r)
		return
	} else if len(pathElements) == 3 { // List Entities (Folders or Collections)
		apiReq.ApiType = pathElements[2]
	} else if len(pathElements) == 4 { // List Entity Contents (Folder Objects or Collection Objects)
		apiReq.ApiType = pathElements[2]
		apiReq.GroupName = pathElements[3]
	} else if len(pathElements) == 5 { // Entity Object (Bukcet Object or Collection Document)
		apiReq.ApiType = pathElements[2]
		apiReq.GroupName = pathElements[3]
		apiReq.ObjectName = pathElements[4]
	} else if len(pathElements) > 4 {
		notFoundHandler(w, r)
		return
	}

	apiReq.doApiRequest()
}