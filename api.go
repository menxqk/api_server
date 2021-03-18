package main 

import (
	"net/http"
	"reflect"
)

const (
	API_GET_ALL = iota + 1
	API_GET_ONE
	API_GET_OBJECT
	API_DELETE
	API_POST
	API_PUT
)

type ApiRequest struct {
	Aud 		*AuthUserData 
	W 			http.ResponseWriter
	R 			*http.Request
	Method 		string
	ApiType		string 
	GroupName	string 
	ObjectName	string
}

var (
	apiFuncs = map[string]func(int, *ApiRequest){
		"buckets": BucketDoRequest,
		"collections": CollectionDoRequest,
	}
)

func (apiReq *ApiRequest) doApiRequest() {
	var reqType int

	if apiReq.Method == http.MethodGet { // GET
		if apiReq.GroupName == "" && apiReq.ObjectName == "" { // GET ALL (Folders or Collections)
			reqType = API_GET_ALL
		}  else if apiReq.GroupName != "" && apiReq.ObjectName == "" { // GET ONE (contents)
			reqType = API_GET_ONE
		} else if apiReq.GroupName != "" && apiReq.ObjectName != "" { // GET OBJECT
			reqType = API_GET_OBJECT
		}
	} else if apiReq.Method == http.MethodDelete { // DELETE
		reqType = API_DELETE
	} else if apiReq.Method == http.MethodPost  { // POST
		reqType = API_POST
	} else if  apiReq.Method == http.MethodPut {
		reqType = API_PUT
	}

	f := apiFuncs[apiReq.ApiType]
	if f != nil && reflect.TypeOf(f).Kind() == reflect.Func {
		f(reqType, apiReq)
	}
}