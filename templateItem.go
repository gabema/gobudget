package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// listTemplateItems lists out all the TemplateItems
func listTemplateItems(w http.ResponseWriter, r *http.Request) {
	templateItems, err := dbGetTemplateItems()
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.RenderList(w, r, newTemplateItemListResponse(templateItems)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TemplateItemCtx middleware is used to load an TemplateItem object from
// the URL parameters passed through as the request. In case
// the TemplateItem could not be found, we stop here and return a 404.
func TemplateItemCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var templateItem *TemplateItem
		var err error
		var templateItemID int

		if templateItemStr := chi.URLParam(r, "templateItemID"); templateItemStr != "" {
			templateItemID, err = strconv.Atoi(templateItemStr)
			templateItem, err = dbGetTemplateItem(templateItemID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "templateItem", templateItem)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func searchTemplateItems(w http.ResponseWriter, r *http.Request) {
	templateItems, _ := dbGetTemplateItems()
	render.RenderList(w, r, newTemplateItemListResponse(templateItems))
}

// createTemplateItem persists the posted TemplateItem and returns it
// back to the client as an acknowledgement.
func createTemplateItem(w http.ResponseWriter, r *http.Request) {
	data := &TemplateItemRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	templateItem := data.TemplateItem
	dbNewTemplateItem(templateItem)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, newTemplateItemResponse(templateItem))
}

// getTemplateItem returns the specific TemplateItem. You'll notice it just
// fetches the TemplateItem right off the context, as its understood that
// if we made it this far, the TemplateItem must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func getTemplateItem(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the templateItem
	// context because this handler is a child of the TemplateItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	templateItem := r.Context().Value("templateItem").(*TemplateItem)

	if err := render.Render(w, r, newTemplateItemResponse(templateItem)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// updateTemplateItem updates an existing TemplateItem in our persistent store.
func updateTemplateItem(w http.ResponseWriter, r *http.Request) {
	templateItem := r.Context().Value("templateItem").(*TemplateItem)

	data := &TemplateItemRequest{TemplateItem: templateItem}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	templateItem = data.TemplateItem
	dbUpdateTemplateItem(templateItem.ID, templateItem)

	render.Render(w, r, newTemplateItemResponse(templateItem))
}

func deleteTemplateItem(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the templateItem
	// context because this handler is a child of the TemplateItemCtx
	// middleware. The worst case, the recoverer middleware will save us.
	templateItem := r.Context().Value("templateItem").(*TemplateItem)

	err = dbRemoveTemplateItem(templateItem.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newTemplateItemResponse(templateItem))
}

type TemplateItem struct {
	ID         int     `db:"id,omitempty" json:"id"`
	TemplateID int     `db:"templateID" json:"tid"`
	BucketID   int     `db:"bucketID" json:"bid"`
	Name       string  `db:"name" json:"name"`
	Deposit    float32 `db:"deposit" json:"d"`
	Withdraw   float32 `db:"withdraw" json:"w"`
}

// TemplateItemRequest is the request payload for TemplateItem data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type TemplateItemRequest struct {
	*TemplateItem
	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *TemplateItemRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	a.ProtectedID = ""                                         // unset the protected ID
	a.TemplateItem.Name = strings.ToLower(a.TemplateItem.Name) // as an example, we down-case
	return nil
}

// TemplateItemResponse is the response payload for the TemplateItem data model.
// See NOTE above in TemplateItemRequest as well.
//
// In the TemplateItemResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type TemplateItemResponse struct {
	*TemplateItem
	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func newTemplateItemResponse(templateItem *TemplateItem) *TemplateItemResponse {
	return &TemplateItemResponse{TemplateItem: templateItem}
}

func (rd *TemplateItemResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

type TemplateItemListResponse []*TemplateItemResponse

func newTemplateItemListResponse(templateItems []*TemplateItem) []render.Renderer {
	list := []render.Renderer{}
	for _, templateItem := range templateItems {
		list = append(list, newTemplateItemResponse(templateItem))
	}
	return list
}
