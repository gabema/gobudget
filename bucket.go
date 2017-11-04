package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

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

// listBuckets lists out all the Buckets
func listBuckets(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, newBucketListResponse(buckets)); err != nil {
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

// updateBucket updates an existing Bucket in our persistent store.
func updateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.Context().Value("bucket").(*Bucket)

	data := &BucketRequest{Bucket: bucket}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	bucket = data.Bucket
	dbUpdateBucket(bucket.Id, bucket)

	render.Render(w, r, newBucketResponse(bucket))
}

func deleteBucket(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the bucket
	// context because this handler is a child of the BucketCtx
	// middleware. The worst case, the recoverer middleware will save us.
	bucket := r.Context().Value("bucket").(*Bucket)

	bucket, err = dbRemoveBucket(bucket.Id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newBucketResponse(bucket))
}

// Bucket fixture data
var buckets = []*Bucket{
	{Id: 1, Name: "Hi", CategoryID: 1, IsLiquid: true},
	{Id: 2, Name: "sup", CategoryID: 2, IsLiquid: true},
	{Id: 3, Name: "alo", CategoryID: 1, IsLiquid: true},
	{Id: 4, Name: "bonjour", CategoryID: 1, IsLiquid: true},
	{Id: 5, Name: "whats up", CategoryID: 2, IsLiquid: true},
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
	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *BucketRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	a.ProtectedID = ""                             // unset the protected ID
	a.Bucket.Name = strings.ToLower(a.Bucket.Name) // as an example, we down-case
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
	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func newBucketResponse(bucket *Bucket) *BucketResponse {
	return &BucketResponse{Bucket: bucket}
}

func (rd *BucketResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

type BucketListResponse []*BucketResponse

func newBucketListResponse(buckets []*Bucket) []render.Renderer {
	list := []render.Renderer{}
	for _, bucket := range buckets {
		list = append(list, newBucketResponse(bucket))
	}
	return list
}

func dbNewBucket(bucket *Bucket) (int, error) {
	bucket.Id = rand.Intn(100) + 10
	buckets = append(buckets, bucket)
	return bucket.Id, nil
}

func dbGetBucket(id int) (*Bucket, error) {
	for _, a := range buckets {
		if a.Id == id {
			return a, nil
		}
	}
	return nil, errors.New("bucket not found")
}

func dbUpdateBucket(id int, bucket *Bucket) (*Bucket, error) {
	for i, a := range buckets {
		if a.Id == id {
			buckets[i] = bucket
			return bucket, nil
		}
	}
	return nil, errors.New("bucket not found")
}

func dbRemoveBucket(id int) (*Bucket, error) {
	for i, a := range buckets {
		if a.Id == id {
			buckets = append((buckets)[:i], (buckets)[i+1:]...)
			return a, nil
		}
	}
	return nil, errors.New("bucket not found")
}
