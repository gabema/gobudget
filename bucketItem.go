package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

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

		if articleID := chi.URLParam(r, "bucketItemID"); articleID != "" {
			bucketItem, err = dbGetBucketItem(articleID)
		} else if articleSlug := chi.URLParam(r, "articleSlug"); articleSlug != "" {
			bucketItem, err = dbGetBucketItemBySlug(articleSlug)
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

	article := data.BucketItem
	dbNewBucketItem(article)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, newBucketItemResponse(article))
}

// getBucketItem returns the specific BucketItem. You'll notice it just
// fetches the BucketItem right off the context, as its understood that
// if we made it this far, the BucketItem must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func getBucketItem(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the article
	// context because this handler is a child of the BucketItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	article := r.Context().Value("bucketItem").(*BucketItem)

	if err := render.Render(w, r, newBucketItemResponse(article)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// updateBucketItem updates an existing BucketItem in our persistent store.
func updateBucketItem(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("bucketItem").(*BucketItem)

	data := &BucketItemRequest{BucketItem: article}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	article = data.BucketItem
	dbUpdateBucketItem(article.ID, article)

	render.Render(w, r, newBucketItemResponse(article))
}

func deleteBucketItem(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the article
	// context because this handler is a child of the BucketItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	article := r.Context().Value("bucketItem").(*BucketItem)

	article, err = dbRemoveBucketItem(article.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newBucketItemResponse(article))
}

// BucketItem data model. I suggest looking at https://upper.io for an easy
// and powerful data persistence adapter.
type BucketItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// BucketItem fixture data
var bucketItems = []*BucketItem{
	{ID: "1", Title: "Hi", Slug: "hi"},
	{ID: "2", Title: "sup", Slug: "sup"},
	{ID: "3", Title: "alo", Slug: "alo"},
	{ID: "4", Title: "bonjour", Slug: "bonjour"},
	{ID: "5", Title: "whats up", Slug: "whats-up"},
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
	a.ProtectedID = ""                                       // unset the protected ID
	a.BucketItem.Title = strings.ToLower(a.BucketItem.Title) // as an example, we down-case
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

func newBucketItemResponse(article *BucketItem) *BucketItemResponse {
	return &BucketItemResponse{BucketItem: article}
}

func (rd *BucketItemResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

type BucketItemListResponse []*BucketItemResponse

func newBucketItemListResponse(bucketItems []*BucketItem) []render.Renderer {
	list := []render.Renderer{}
	for _, article := range bucketItems {
		list = append(list, newBucketItemResponse(article))
	}
	return list
}

func dbNewBucketItem(bucketItem *BucketItem) (string, error) {
	bucketItem.ID = fmt.Sprintf("%d", rand.Intn(100)+10)
	bucketItems = append(bucketItems, bucketItem)
	return bucketItem.ID, nil
}

func dbGetBucketItem(id string) (*BucketItem, error) {
	for _, a := range bucketItems {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbGetBucketItemBySlug(slug string) (*BucketItem, error) {
	for _, a := range bucketItems {
		if a.Slug == slug {
			return a, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbUpdateBucketItem(id string, article *BucketItem) (*BucketItem, error) {
	for i, a := range bucketItems {
		if a.ID == id {
			bucketItems[i] = article
			return article, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbRemoveBucketItem(id string) (*BucketItem, error) {
	for i, a := range bucketItems {
		if a.ID == id {
			bucketItems = append((bucketItems)[:i], (bucketItems)[i+1:]...)
			return a, nil
		}
	}
	return nil, errors.New("article not found")
}
