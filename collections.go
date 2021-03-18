package main 

import (
	"log"
	"context"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type CollectionDocument struct {
	CollectionID string
	ID string
	Data map[string]interface{}
}

func CollectionDoRequest(reqType int, apiReq *ApiRequest) {
	switch reqType {
	case API_GET_ALL:
		CollectionGetAll(apiReq)
	case API_GET_ONE:
		CollectionGetOne(apiReq)
	case API_GET_OBJECT:
		CollectionGetObject(apiReq)
	case API_DELETE:
		CollectionDeleteObject(apiReq)
	case API_POST:
		CollectionPostObject(apiReq)
	case API_PUT:
		CollectionPutObject(apiReq)
	}
}

func CollectionGetAll(apiReq *ApiRequest) {
	collections := []string{}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, apiReq.Aud.ProjectID)
	if err != nil {
		log.Printf("CollectionGetOne firestore.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	iter := client.Collections(ctx)
	for {
		coll, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("CollectionGetOne iter.Next() error: %v\n", err)
			return 
		}

		collections = append(collections, coll.ID)
	}

	sendJson(apiReq.W, apiReq.R, collections)
}

func CollectionGetOne(apiReq *ApiRequest) {
	collectionDocuments := []string{}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, apiReq.Aud.ProjectID)
	if err != nil {
		log.Printf("CollectionGetOne firestore.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	iter := client.Collection(apiReq.GroupName).Documents(ctx)
	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("CollectionGetOne iter.Next() error: %v\n", err)
			return 
		}

		collectionDocuments = append(collectionDocuments, attrs.Ref.ID)
	}

	sendJson(apiReq.W, apiReq.R, collectionDocuments)
}

func CollectionGetObject(apiReq *ApiRequest) {
	cDoc := CollectionDocument{
		CollectionID: apiReq.GroupName, 
		ID: apiReq.ObjectName,
		Data: map[string]interface{}{},
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, apiReq.Aud.ProjectID)
	if err != nil {
		log.Printf("CollectionGetObject firestore.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	dsnap, err := client.Collection(cDoc.CollectionID).Doc(cDoc.ID).Get(ctx)
	if err != nil {
		log.Printf("CollectionGetObject client.Get() error: %v\n", err)
		return
	}

	cDoc.Data = dsnap.Data()

	sendJson(apiReq.W, apiReq.R, cDoc.Data)
}

func CollectionDeleteObject(apiReq *ApiRequest) {
	cDoc := CollectionDocument{
		CollectionID: apiReq.GroupName, 
		ID: apiReq.ObjectName,
		Data: map[string]interface{}{},
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, apiReq.Aud.ProjectID)
	if err != nil {
		log.Printf("CollectionDeleteObject firestore.NewClient() error: %v\n", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	doc := client.Collection(cDoc.CollectionID).Doc(cDoc.ID)
	if _, err = doc.Delete(ctx); err != nil {
		log.Printf("CollectionDeleteObject doc.Delete() error: %v\n", err)
	}
}

func CollectionPostObject(apiReq *ApiRequest) {
	cDoc := CollectionDocument{
		CollectionID: apiReq.GroupName, 
	}
	doDocumentUpload(apiReq.R, &cDoc)
	if len(cDoc.Data) < 1 {
		return
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, apiReq.Aud.ProjectID)
	if err != nil {
		log.Printf("CollectionPutObject firestore.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	if _, err = client.Collection(cDoc.CollectionID).NewDoc().Set(ctx, cDoc.Data); err != nil {
		log.Printf("CollectionPostObject client.NewDoc().Set() error: %v\n", err)		
	}
}

func CollectionPutObject(apiReq *ApiRequest) {
	cDoc := CollectionDocument{
		CollectionID: apiReq.GroupName, 
		ID: apiReq.ObjectName,
	}
	doDocumentUpload(apiReq.R, &cDoc)
	if len(cDoc.Data) < 1 {
		return
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, apiReq.Aud.ProjectID)
	if err != nil {
		log.Printf("CollectionPutObject firestore.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	updates := []firestore.Update{}
	for key, value := range cDoc.Data {
		update := firestore.Update{
			Path: key,
			Value: value,
		}
		updates = append(updates, update)
	}

	if _, err = client.Collection(cDoc.CollectionID).Doc(cDoc.ID).Update(ctx, updates); err != nil {
		log.Printf("CollectionPutObject client.NewDoc().Set() error: %v\n", err)		
	}
}

func doDocumentUpload(r *http.Request, cDoc *CollectionDocument) {
	data := map[string]interface{}{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("doDocumentUpload ioutil.ReadAll() error: %v\n", err)
		return
	}
	
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("doDocumentUpload json.Unmarshal() error: %v\n", err)
		return
	}

	cDoc.Data = data
}