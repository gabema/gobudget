package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type Category struct {
	Name string `db:"name" json:"name"`

	Id int `db:"id,omitempty" json:"id"`
}

// listCategories lists out all the Categories
func listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := dbGetCategories()
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.RenderList(w, r, newCategoryListResponse(categories)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CategoryCtx middleware is used to load an Category object from
// the URL parameters passed through as the request. In case
// the Category could not be found, we stop here and return a 404.
func CategoryCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var category *Category
		var err error

		if categoryStr := chi.URLParam(r, "categoryID"); categoryStr != "" {
			categoryID, _ := strconv.Atoi(categoryStr)
			category, err = dbGetCategory(categoryID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "category", category)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// createCategory persists the posted Category and returns it
// back to the client as an acknowledgement.
func createCategory(w http.ResponseWriter, r *http.Request) {
	data := &CategoryRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	category := data.Category
	dbNewCategory(category)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, newCategoryResponse(category))
}

// getCategory returns the specific Category. You'll notice it just
// fetches the Category right off the context, as its understood that
// if we made it this far, the Category must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func getCategory(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the category
	// context because this handler is a child of the CategoryCtx
	// middleware. The worst case, the recoverer middleware will save us.
	category := r.Context().Value("category").(*Category)

	if err := render.Render(w, r, newCategoryResponse(category)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// updateCategory updates an existing Category in our persistent store.
func updateCategory(w http.ResponseWriter, r *http.Request) {
	category := r.Context().Value("category").(*Category)

	data := &CategoryRequest{Category: category}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	category = data.Category
	dbUpdateCategory(category.Id, category)

	render.Render(w, r, newCategoryResponse(category))
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the category
	// context because this handler is a child of the CategoryCtx
	// middleware. The worst case, the recoverer middleware will save us.
	category := r.Context().Value("category").(*Category)

	err = dbRemoveCategory(category.Id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newCategoryResponse(category))
}

// CategoryRequest is the request payload for Category data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type CategoryRequest struct {
	*Category
	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *CategoryRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	a.ProtectedID = ""                                 // unset the protected ID
	a.Category.Name = strings.ToLower(a.Category.Name) // as an example, we down-case
	return nil
}

// CategoryResponse is the response payload for the Category data model.
// See NOTE above in CategoryRequest as well.
//
// In the CategoryResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type CategoryResponse struct {
	*Category
	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func newCategoryResponse(category *Category) *CategoryResponse {
	return &CategoryResponse{Category: category}
}

func (rd *CategoryResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

type CategoryListResponse []*CategoryResponse

func newCategoryListResponse(categories []*Category) []render.Renderer {
	list := []render.Renderer{}
	for _, category := range categories {
		list = append(list, newCategoryResponse(category))
	}
	return list
}
