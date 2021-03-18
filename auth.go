package main 

import (
	"net/http"
	"fmt"
	"context"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"strings"
	"log"
	"encoding/json"
	"time"
)

type AuthUserData struct {
	Username 		string		`json:"username,omitempty"`
	Role			string		`json:"role,omitempty"`
	ProjectID		string		`json:"projectid,omitempty"`
	BucketID		string		`json:"bucketid,omitempty"`
}

type Session struct {
	IsNew		bool	
	Id			string	
	DocId 		string
	Values		map[string]interface{}				
}

func AuthenticateUser(w http.ResponseWriter, r *http.Request, username, password string) (bool, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("AuthenticateUser firestore.NewClient() error: %v", err)
	}
	defer client.Close()

	iter := client.Collection(usersCollection).Where("username", "==", username).Limit(1).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return false, fmt.Errorf("AuthenticateUser iter.Next() error: %v", err)
		}
		
		if doc.Data()["hashpassword"] != nil {
			hashpassword, ok := doc.Data()["hashpassword"].(string)
			if ok {
				if checkPasswordHash(password, hashpassword) {
					aud := AuthUserData{}

					if doc.Data()["username"] != nil { aud.Username, _ = doc.Data()["username"].(string) }
					if doc.Data()["role"] != nil { aud.Role, _ = doc.Data()["role"].(string) }
					if doc.Data()["projectid"] != nil { aud.ProjectID, _ = doc.Data()["projectid"].(string) }
					if doc.Data()["bucketid"] != nil { aud.BucketID, _ = doc.Data()["bucketid"].(string) }
					// if doc.Data()["collectionid"] != nil { aud.CollectionID, _ = doc.Data()["collectionid"].(string) }

					session, err := setSession(w, r, &aud)
					if err != nil {
						return false, err
					} else {
						expire := time.Now().Add(100*time.Minute)
						cookie := http.Cookie{
							Name: COOKIE_NAME,
							Value: session.Id,
							Expires: expire,
							// Path: "",
							// Domain: "",
							// RawExpires "",
							// MaxAge=0 means no 'Max-Age' attribute specified.
							// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
							// MaxAge>0 means Max-Age attribute present and given in seconds
							// MaxAge   int
							// Secure   bool
							// HttpOnly bool
							// Raw      string
							// Unparsed []string // Raw text of unparsed attribute-value pairs
						}
						http.SetCookie(w, &cookie)	

						return true, nil
					}
				}
			}
		}
	}


	return false, fmt.Errorf("Could not Authenticate %s", username)
}

func IsAuthenticated(r *http.Request) (bool, *AuthUserData) {
	var aud *AuthUserData
	var ok bool

	session := getSession(r)
	if !session.IsNew {
		authUserData := AuthUserData{}
		
		authenticated, ok1 := session.Values["authenticated"].(bool)
		audS, ok2 := session.Values["aud"].(string)
		if ok2 && audS != "" {
			err := json.Unmarshal([]byte(audS), &authUserData)
			if err != nil {
				log.Printf("IsAuthenticated json.Unmarshal() error:  %v\n", err)
			} else {
				ok = authenticated && ok1 && (audS != "") && ok2
				aud = &authUserData
			}
		}
	}

	return ok, aud
}

func RemoveAuthentication(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)

	if !session.IsNew {
		docId := session.DocId
		ctx := context.Background()
		client, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			log.Printf("RemoveAuthentication firestore.NewClient() error: %v\n", err)
		} else {
			defer client.Close()

			_, err := client.Collection(sessionsCollection).Doc(docId).Delete(ctx)
			if err != nil {
				log.Printf("RemoveAuthentication Doc.Delete() %q error: %v\n", docId, err)
			}
		}
	}

	expire := time.Now().Add(-10*time.Minute)
	cookie := http.Cookie{
		Name: COOKIE_NAME,
		Value: "deleted",
		Expires: expire,
	}

	http.SetCookie(w, &cookie)
}

func setSession(w http.ResponseWriter, r *http.Request, aud *AuthUserData) (*Session, error) {
	session := getSession(r)

	audJson, err := json.Marshal(*aud)
	if err != nil {
		return session, fmt.Errorf("setAuthentication json.Marshal() error: %v", err)
	}

	session.Values["authenticated"] = true
	session.Values["aud"] = string(audJson)

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return session, fmt.Errorf("setAuthentication firestore.NewClient() error: %v", err)
	}
	defer client.Close()

	data := map[string]interface{}{}
	data["id"] = session.Id
	data["values"] = session.Values

	if session.IsNew {
		doc := client.Collection(sessionsCollection).NewDoc()
		docId := doc.ID
		data["docid"] = docId
		if _, err := doc.Set(ctx, data); err != nil {
			return session, fmt.Errorf("setAuthentication NewDoc.Set() error: %v", err)
		}	
	} else {
		docId := session.DocId
		data["docid"] = docId
		if _, err := client.Collection(sessionsCollection).Doc(docId).Set(ctx, data); err != nil {
			return session, fmt.Errorf("setAuthentication Doc.Set() error: %v", err)
		}
	}

	return session, nil
}

func getSession(r *http.Request) *Session {
	session := Session{
		IsNew: true,
		Id: uuid.New().String(),
		Values: map[string]interface{}{},
	}

	cookie, err := r.Cookie(COOKIE_NAME)
	if err != http.ErrNoCookie {
		
		values := strings.Split(cookie.String(), "=")
		var sessionId string 
		if len(values) > 1 {
			sessionId = values[1]
		}

		ctx := context.Background()
		client, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			log.Printf("getSession firestore.NewClient() error: %v\n", err)
		} else {
			defer client.Close()

			iter := client.Collection(sessionsCollection).Where("id", "==", sessionId).Limit(1).Documents(ctx)
			for {
				doc, err := iter.Next()
				
				if err == iterator.Done {
					break
				} else if err != nil {
					log.Printf("getSession iter.Next() error: %v\n", err)
					break					
				} else {
					if doc.Data()["id"] != nil {
						id, ok := doc.Data()["id"].(string)
						if ok {
							session.Id = id
							session.IsNew = false
						}
					}
					if doc.Data()["docid"] != nil {
						docid, ok := doc.Data()["docid"].(string)
						if ok {
							session.DocId = docid
						}
					}
					if doc.Data()["values"] != nil {
						values, ok := doc.Data()["values"].(map[string]interface{})
						if ok {
							session.Values = values
						}
					}

					break
				}
			}
		}
	}

	return &session
}


func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}