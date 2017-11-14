package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// listBucketItems lists out all the BucketItems
func listBucketItems(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	bucketID, _ := strconv.Atoi(qs.Get("bid"))
	dateStart := qs.Get("dstart")
	dateEnd := qs.Get("dend")
	inName := qs.Get("namePart")
	var pageSize, pageStart int
	var err error
	if pageSize, err = strconv.Atoi(qs.Get("ps")); err != nil {
		pageSize = 0
	}
	pageStart, _ = strconv.Atoi(qs.Get("po"))

	bucketItems, err := dbGetBucketItems(bucketID, dateStart, dateEnd, inName, pageSize, pageStart)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err = render.RenderList(w, r, newBucketItemListResponse(bucketItems)); err != nil {
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

		if bucketItemStr := chi.URLParam(r, "bucketItemID"); bucketItemStr != "" {
			var bucketItemID int
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

// createBucketItem persists the posted BucketItem and returns it
// back to the client as an acknowledgement.
func createBucketItem(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("batch") == "" {
		data := &BucketItemRequest{}
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		bucketItem := data.BucketItem
		if err := dbNewBucketItem(bucketItem); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		render.Status(r, http.StatusCreated)
		render.Render(w, r, newBucketItemResponse(bucketItem))
	} else {
		data := &BucketItemsRequest{}
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		bucketItems := data.Items
		if err := dbNewBucketItems(bucketItems); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		render.Status(r, http.StatusCreated)
		render.Render(w, r, &BucketItemsResponse{count: len(bucketItems)})
	}
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
	bucketItemID := bucketItem.ID
	bucketItem.ID = 0
	if err := dbUpdateBucketItem(bucketItemID, bucketItem); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newBucketItemResponse(bucketItem))
}

func deleteBucketItem(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the bucketItem
	// context because this handler is a child of the BucketItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	bucketItem := r.Context().Value("bucketItem").(*BucketItem)

	err = dbRemoveBucketItem(bucketItem.ID)
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
}

func (a *BucketItemRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	return nil
}

type BucketItemsRequest struct {
	Items []BucketItem `json:"items"`
}

func (a *BucketItemsRequest) Bind(r *http.Request) error {
	return nil
}

type BucketItemsResponse struct {
	count int
}

func (rd *BucketItemsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
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
}

func newBucketItemResponse(bucketItem *BucketItem) *BucketItemResponse {
	return &BucketItemResponse{BucketItem: bucketItem}
}

func (rd *BucketItemResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
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
