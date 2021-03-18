package main 

import (
	"net/http"
	"encoding/json"
	"log"
)

func sendJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("sendJson json.Marshal() error: %v\n", err)
		log.Printf("data: %+v\n", data)
		internalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	n, err := w.Write(jsonData)
	if err != nil {
		log.Printf("sendJson w.Write() error: %+v\n", err)
		log.Printf("sendJson w.Write() %d of %d bytes written\n", n, len(jsonData))
	}
}

func sendData(w http.ResponseWriter, contentType string, name string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="  + name)
	
	n, err := w.Write(data)
	if err != nil {
		log.Printf("sendData w.Write() error: %v\n", err)
		log.Printf("sendData w.Write() %d of %d bytes written\n", n, len(data))
	}
}