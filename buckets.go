package main 

import (
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"context"
	"time"
	"log"
	"net/http"
	"io/ioutil"
	"strconv"
	"strings"
)

type BucketObject struct {
	FolderName string 
	ObjectName string
	Size int64
	ContentType string
	Data []byte
}


func BucketDoRequest(reqType int, apiReq *ApiRequest) {
	switch reqType {
	case API_GET_ALL:
		BucketGetAll(apiReq)
	case API_GET_ONE:
		BucketGetOne(apiReq)
	case API_GET_OBJECT:
		BucketGetObject(apiReq)
	case API_DELETE:
		BucketDeleteObject(apiReq)
	case API_POST:
		BucketPostObject(apiReq)
	case API_PUT:
		BucketPostObject(apiReq)
	}
}

func BucketGetAll(apiReq *ApiRequest) {
	bucketFolders := []string{}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("BucketAll storage.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	query := &storage.Query{
		// Prefix: "",
		Delimiter: ".",
		// StartOffset: "bar/",  // Only list objects lexicographically >= "bar/"
		// EndOffset: "foo/",    // Only list objects lexicographically < "foo/"
	}

	iter := client.Bucket(apiReq.Aud.BucketID).Objects(ctx, query)
	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("getContentsFromBucket iter.Next() error %v\n", err)
			return 
		}

		if strings.TrimSpace(strings.ReplaceAll(attrs.Name, "/", "")) != "" {
			bucketFolders = append(bucketFolders, strings.TrimSpace(strings.ReplaceAll(attrs.Name, "/", "")))
		}
	}

	sendJson(apiReq.W, apiReq.R, bucketFolders)
}

func BucketGetOne(apiReq *ApiRequest) {
	bucketObjects := []string{}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("BucketGetone storage.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	query := &storage.Query{
		Prefix: apiReq.GroupName + "/",
		// Delimiter: ".",
		// StartOffset: "bar/",  // Only list objects lexicographically >= "bar/"
		// EndOffset: "foo/",    // Only list objects lexicographically < "foo/"
	}

	iter := client.Bucket(apiReq.Aud.BucketID).Objects(ctx, query)
	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("BucketGetOne iter.Next() error %v\n", err)
			return 
		}

		attrs.Name = strings.ReplaceAll(attrs.Name, apiReq.GroupName + "/", "")

		if strings.TrimSpace(strings.ReplaceAll(attrs.Name, apiReq.GroupName + "/", "")) != "" {
			bucketObjects = append(bucketObjects, strings.TrimSpace(strings.ReplaceAll(attrs.Name, apiReq.GroupName + "/", "")))
		}
	}

	sendJson(apiReq.W, apiReq.R, bucketObjects)
}

func BucketGetObject(apiReq *ApiRequest) {
	bObj := BucketObject{
		FolderName: apiReq.GroupName, 
		ObjectName: apiReq.ObjectName,
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("BucketGetObject storage.NewClient() error: %v", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	obj := client.Bucket(apiReq.Aud.BucketID).Object(bObj.FolderName + "/" + bObj.ObjectName)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		log.Printf("BucketGetObject obj.Attrs() error: %v", err)
		return 
	}

	bObj.ObjectName = attrs.Name // FOLDER NAME + OBJECT NAME
	bObj.Size = attrs.Size
	bObj.ContentType = attrs.ContentType

	rc, err := client.Bucket(apiReq.Aud.BucketID).Object(bObj.ObjectName).NewReader(ctx)
	if err != nil {
		log.Printf("BucketGetObject rc.NewReader() error: %v", err)
		return
	}
	defer rc.Close()

	bObj.Data, err = ioutil.ReadAll(rc)
	if err != nil {
		log.Printf("BucketGetObject ioutil.ReadAll() error: %v", err)
		return 
	}
	
	sendData(apiReq.W, bObj.ContentType, bObj.ObjectName, bObj.Data)
}

func BucketDeleteObject(apiReq *ApiRequest) {
	bObj := BucketObject{
		FolderName: apiReq.GroupName, 
		ObjectName: apiReq.ObjectName,
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("BucketDeleteObject storage.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	obj := client.Bucket(apiReq.Aud.BucketID).Object(bObj.FolderName + "/" + bObj.ObjectName)
	if err = obj.Delete(ctx); err != nil {
		log.Printf("BucketDeleteObject obj.NewClient() error: %v\n", err)
		return 
	}
}

func BucketPostObject(apiReq *ApiRequest) {
	bObj := BucketObject{
		FolderName: apiReq.GroupName,
	}
	doFileUpload(apiReq.R, &bObj)
	if len(bObj.Data) < 1 {
		return
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("BucketPutObject storage.NewClient() error: %v\n", err)
		return 
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	obj := client.Bucket(apiReq.Aud.BucketID).Object(bObj.FolderName + "/" + bObj.ObjectName)
	wc := obj.NewWriter(ctx)
	if _, err = wc.Write(bObj.Data); err != nil {
		log.Printf("BucketPutObject wc.Write() error: %v\n", err)
		return 
	}
	if err = wc.Close(); err != nil {
		log.Printf("BucketPutObject wc.Close() error: %v\n", err)
		return 
	}

	uAttrs := storage.ObjectAttrsToUpdate {
		ContentType: bObj.ContentType,
	}
	if _, err = obj.Update(ctx, uAttrs); err != nil {
		log.Printf("BucketPutObject obj.Update() error: %v\n", err)
		return 
	}
}

func doFileUpload(r *http.Request, bObj *BucketObject) {
	r.ParseMultipartForm(10 << 20)

	var uploadName string
	if len(r.MultipartForm.Value["upload-name"]) > 0 {
		uploadName = r.MultipartForm.Value["upload-name"][0]
	}
	var uploadSize int64
	if len(r.MultipartForm.Value["upload-size"]) > 0 {
		var err error
		uploadSize, err = strconv.ParseInt(r.MultipartForm.Value["upload-size"][0], 10, 64)
		if err != nil {
			uploadSize = -1;
		}
	}
	var uploadType string
	if len(r.MultipartForm.Value["upload-type"]) > 0 {
		uploadType = r.MultipartForm.Value["upload-type"][0]
	}

	multipartFile, handler, err := r.FormFile("upload-file")
	if err != nil {
		log.Printf("doFileUpload r.FormFile() error %s: %v\n", handler.Filename, err)
		return
	}
	defer multipartFile.Close()

	log.Println("uploadName", uploadName)
	log.Println("uploadSize", uploadSize)
	log.Println("uploadType", uploadType)

	
	bObj.ObjectName = uploadName
	bObj.Size = uploadSize
	bObj.ContentType = uploadType
	bObj.Data, err = ioutil.ReadAll(multipartFile)
	if err != nil {
		log.Printf("doFileUpload ioutil.ReadAll() error: %v\n", err)
	}
}