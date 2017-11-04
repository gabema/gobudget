package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// listBucketItems lists out all the BucketItems
func listBucketItems(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, newBucketItemListResponse(bucketItems)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// BucketItemCtx middleware is used to load an BucketItem object from
// the URL parameters passed through as the request. In case
// the BucketItem could not be found, we stop here and return a 404.
func BucketItemCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bucketItem *BucketItem
		var err error
		var bucketItemID int

		if bucketItemStr := chi.URLParam(r, "bucketItemID"); bucketItemStr != "" {
			bucketItemID, err = strconv.Atoi(bucketItemStr)
			bucketItem, err = dbGetBucketItem(bucketItemID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "bucketItem", bucketItem)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func searchBucketItems(w http.ResponseWriter, r *http.Request) {
	render.RenderList(w, r, newBucketItemListResponse(bucketItems))
}

// createBucketItem persists the posted BucketItem and returns it
// back to the client as an acknowledgement.
func createBucketItem(w http.ResponseWriter, r *http.Request) {
	data := &BucketItemRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	bucketItem := data.BucketItem
	dbNewBucketItem(bucketItem)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, newBucketItemResponse(bucketItem))
}

// getBucketItem returns the specific BucketItem. You'll notice it just
// fetches the BucketItem right off the context, as its understood that
// if we made it this far, the BucketItem must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func getBucketItem(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the bucketItem
	// context because this handler is a child of the BucketItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	bucketItem := r.Context().Value("bucketItem").(*BucketItem)

	if err := render.Render(w, r, newBucketItemResponse(bucketItem)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// updateBucketItem updates an existing BucketItem in our persistent store.
func updateBucketItem(w http.ResponseWriter, r *http.Request) {
	bucketItem := r.Context().Value("bucketItem").(*BucketItem)

	data := &BucketItemRequest{BucketItem: bucketItem}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	bucketItem = data.BucketItem
	dbUpdateBucketItem(bucketItem.ID, bucketItem)

	render.Render(w, r, newBucketItemResponse(bucketItem))
}

func deleteBucketItem(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the bucketItem
	// context because this handler is a child of the BucketItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	bucketItem := r.Context().Value("bucketItem").(*BucketItem)

	bucketItem, err = dbRemoveBucketItem(bucketItem.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newBucketItemResponse(bucketItem))
}

type BucketItem struct {
	ID          int       `db:"id,omitempty" json:"id"`
	BucketID    int       `db:"bucketID" json:"bucketID"`
	Name        string    `db:"name" json:"name"`
	Transaction time.Time `db:"transaction" json:"transaction"`
	Deposit     float32   `db:"deposit" json:"d"`
	Withdraw    float32   `db:"withdrawl" json:"w"`
}

// BucketItem fixture data
var bucketItems = []*BucketItem{
	{ID: 5, BucketID: 1, Name: "whats-up", Transaction: time.Date(1959, time.March, 21, 0, 0, 0, 0, time.UTC), Deposit: 14.99, Withdraw: 0.0},
}

// BucketItemRequest is the request payload for BucketItem data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type BucketItemRequest struct {
	*BucketItem
	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *BucketItemRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	a.ProtectedID = ""                                     // unset the protected ID
	a.BucketItem.Name = strings.ToLower(a.BucketItem.Name) // as an example, we down-case
	return nil
}

// BucketItemResponse is the response payload for the BucketItem data model.
// See NOTE above in BucketItemRequest as well.
//
// In the BucketItemResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type BucketItemResponse struct {
	*BucketItem
	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func newBucketItemResponse(bucketItem *BucketItem) *BucketItemResponse {
	return &BucketItemResponse{BucketItem: bucketItem}
}

func (rd *BucketItemResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

type BucketItemListResponse []*BucketItemResponse

func newBucketItemListResponse(bucketItems []*BucketItem) []render.Renderer {
	list := []render.Renderer{}
	for _, bucketItem := range bucketItems {
		list = append(list, newBucketItemResponse(bucketItem))
	}
	return list
}

func dbNewBucketItem(bucketItem *BucketItem) (int, error) {
	bucketItem.ID = rand.Intn(100) + 10
	bucketItems = append(bucketItems, bucketItem)
	return bucketItem.ID, nil
}

func dbGetBucketItem(id int) (*BucketItem, error) {
	for _, a := range bucketItems {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("bucketItem not found")
}

func dbUpdateBucketItem(id int, bucketItem *BucketItem) (*BucketItem, error) {
	for i, a := range bucketItems {
		if a.ID == id {
			bucketItems[i] = bucketItem
			return bucketItem, nil
		}
	}
	return nil, errors.New("bucketItem not found")
}

func dbRemoveBucketItem(id int) (*BucketItem, error) {
	for i, a := range bucketItems {
		if a.ID == id {
			bucketItems = append((bucketItems)[:i], (bucketItems)[i+1:]...)
			return a, nil
		}
	}
	return nil, errors.New("bucketItem not found")
}
