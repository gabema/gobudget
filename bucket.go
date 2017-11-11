package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type Bucket struct {
	Name string `db:"name" json:"name"`

	Id int `db:"id,omitempty" json:"id"`

	CategoryID int `db:"categoryID"  json:"categoryID"`

	Description string `db:"description"  json:"desc"`

	IsLiquid bool `db:"isLiquid"  json:"liq"`
}

type BucketSummary struct {
	BucketID     int     `db:"bucketID" json:"bid"`
	CategoryName string  `db:"categoryName" json:"cn"`
	BucketName   string  `db:"bucketName" json:"bn"`
	Total        float32 `db:"total" json:"t"`
	IsLiquid     bool    `db:"isLiquid" json:"l"`
}

// listBuckets lists out all the Buckets
func listBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := dbGetBuckets()
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err = render.RenderList(w, r, newBucketListResponse(buckets)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// BucketCtx middleware is used to load an Bucket object from
// the URL parameters passed through as the request. In case
// the Bucket could not be found, we stop here and return a 404.
func BucketCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bucket *Bucket
		var err error

		if bucketStr := chi.URLParam(r, "bucketID"); bucketStr != "" {
			bucketID, _ := strconv.Atoi(bucketStr)
			bucket, err = dbGetBucket(bucketID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "bucket", bucket)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// createBucket persists the posted Bucket and returns it
// back to the client as an acknowledgement.
func createBucket(w http.ResponseWriter, r *http.Request) {
	data := &BucketRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	bucket := data.Bucket
	dbNewBucket(bucket)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, newBucketResponse(bucket))
}

// getBucket returns the specific Bucket. You'll notice it just
// fetches the Bucket right off the context, as its understood that
// if we made it this far, the Bucket must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func getBucket(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the bucket
	// context because this handler is a child of the BucketCtx
	// middleware. The worst case, the recoverer middleware will save us.
	bucket := r.Context().Value("bucket").(*Bucket)

	if err := render.Render(w, r, newBucketResponse(bucket)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func summarizeBuckets(w http.ResponseWriter, r *http.Request) {
	if bucketSummaries, err := dbSummarizeBuckets(); err == nil {
		render.RenderList(w, r, newBucketSummaryResponse(bucketSummaries))
	} else {
		render.Render(w, r, ErrRender(err))
	}
}

// updateBucket updates an existing Bucket in our persistent store.
func updateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.Context().Value("bucket").(*Bucket)

	data := &BucketRequest{Bucket: bucket}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	bucket = data.Bucket
	bucketID := bucket.Id
	bucket.Id = 0

	if err := dbUpdateBucket(bucketID, bucket); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newBucketResponse(bucket))
}

func deleteBucket(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the bucket
	// context because this handler is a child of the BucketCtx
	// middleware. The worst case, the recoverer middleware will save us.
	bucket := r.Context().Value("bucket").(*Bucket)

	err = dbRemoveBucket(bucket.Id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newBucketResponse(bucket))
}

// BucketRequest is the request payload for Bucket data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type BucketRequest struct {
	*Bucket
	// ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *BucketRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	// a.ProtectedID = "" // unset the protected ID
	return nil
}

// BucketResponse is the response payload for the Bucket data model.
// See NOTE above in BucketRequest as well.
//
// In the BucketResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type BucketResponse struct {
	*Bucket
}

type BucketSummaryResponse struct {
	*BucketSummary
}

func (rd *BucketSummary) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

func newBucketResponse(bucket *Bucket) *BucketResponse {
	return &BucketResponse{Bucket: bucket}
}

func (rd *BucketResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

func newBucketSummaryResponse(bucketSummaries []BucketSummary) []render.Renderer {
	list := []render.Renderer{}
	for _, bucketSummary := range bucketSummaries {
		bs := bucketSummary
		list = append(list, &bs)
	}
	return list
}

func newBucketListResponse(buckets []*Bucket) []render.Renderer {
	list := []render.Renderer{}
	for _, bucket := range buckets {
		list = append(list, newBucketResponse(bucket))
	}
	return list
}
