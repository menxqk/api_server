package main 

import (
	"fmt"
	"os"
	"flag"
	"net/http"
	"log"
	"context"
)

var (
	projectID = "robust-service-285202"
	sessionsCollection = "sessions"
	usersCollection = "users"
	COOKIE_NAME = "SESSIONID"
)

func main() {
	fmt.Println("ApiServer v0.02")

	// // if len(os.Args) < 7 {
	// // 	printUsageInformation()
	// // 	return
	// // }

	// parseFlags()

	var port string
	if os.Getenv("PORT") == "" {
		port = "8080"
	} else {
		port = os.Getenv("PORT")
	}

	router := http.NewServeMux()
	setRoutes(router)
	http.Handle("/", router)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		fmt.Printf("ApiServer ListenAndServe() error: %v\n", err)
		os.Exit(1)
	}
}

func printUsageInformation() {
	fmt.Println("Usage: api_server -project PROJECT_ID -sessions SESSIONS_COLLECTION -users USERS_COLLECTION")
}

func parseFlags() {
	project := flag.String("project", "", "Project_ID")
	sessions := flag.String("sessions", "", "Sessions Collection")
	users := flag.String("users", "", "Users Collection")
	flag.Parse()

	projectID = *project
	sessionsCollection = *sessions
	usersCollection = *users

	fmt.Println("projectID:", projectID)
	fmt.Println("sessionsCollection:", sessionsCollection)
	fmt.Println("usersCollection:", usersCollection)
}

func setRoutes(router *http.ServeMux) {
	
	router.HandleFunc("/", makeLogHandler(indexHandler))
	router.HandleFunc("/styles.css", makeLogHandler(cssHandler))
	router.HandleFunc("/script.js", makeLogHandler(jsHandler))

	router.HandleFunc("/login", makeLogHandler(loginHandler))
	router.HandleFunc("/logout", makeLogHandler(logoutHandler))

	router.HandleFunc("/app", makeLogHandler(makeAuthHandler(appHandler)))
	router.HandleFunc("/app/styles.css", makeLogHandler(makeAuthHandler(appCssHandler)))
	router.HandleFunc("/app/script.js", makeLogHandler(makeAuthHandler(appJsHandler)))
	router.HandleFunc("/app/api/", makeLogHandler(makeAuthHandler(appApiHandler)))

}

func makeLogHandler(f http.HandlerFunc) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path, r.RemoteAddr)
		f(w, r)
	}
}

func makeAuthHandler(f http.HandlerFunc) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		authOk, aud := IsAuthenticated(r)
		if authOk && aud != nil {
			newReq := r.WithContext(context.WithValue(r.Context(), "aud", aud))
			f(w, newReq)

		} else {
			unauthorizedHandler(w, r)
		}
	}
}

