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

type Template struct {
	Name string `db:"name" json:"name"`
	Id   int    `db:"id,omitempty" json:"id"`
}

// listTemplates lists out all the Templates
func listTemplates(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, newTemplateListResponse(templates)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TemplateCtx middleware is used to load an Template object from
// the URL parameters passed through as the request. In case
// the Template could not be found, we stop here and return a 404.
func TemplateCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var template *Template
		var err error

		if templateStr := chi.URLParam(r, "templateID"); templateStr != "" {
			templateID, _ := strconv.Atoi(templateStr)
			template, err = dbGetTemplate(templateID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "template", template)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// createTemplate persists the posted Template and returns it
// back to the client as an acknowledgement.
func createTemplate(w http.ResponseWriter, r *http.Request) {
	data := &TemplateRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	template := data.Template
	dbNewTemplate(template)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, newTemplateResponse(template))
}

// getTemplate returns the specific Template. You'll notice it just
// fetches the Template right off the context, as its understood that
// if we made it this far, the Template must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func getTemplate(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the template
	// context because this handler is a child of the TemplateCtx
	// middleware. The worst case, the recoverer middleware will save us.
	template := r.Context().Value("template").(*Template)

	if err := render.Render(w, r, newTemplateResponse(template)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// updateTemplate updates an existing Template in our persistent store.
func updateTemplate(w http.ResponseWriter, r *http.Request) {
	template := r.Context().Value("template").(*Template)

	data := &TemplateRequest{Template: template}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	template = data.Template
	dbUpdateTemplate(template.Id, template)

	render.Render(w, r, newTemplateResponse(template))
}

func deleteTemplate(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the template
	// context because this handler is a child of the TemplateCtx
	// middleware. The worst case, the recoverer middleware will save us.
	template := r.Context().Value("template").(*Template)

	template, err = dbRemoveTemplate(template.Id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, newTemplateResponse(template))
}

// Template fixture data
var templates = []*Template{
	{Id: 1, Name: "Hi"},
	{Id: 2, Name: "sup"},
	{Id: 3, Name: "alo"},
	{Id: 4, Name: "bonjour"},
	{Id: 5, Name: "whats up"},
}

// TemplateRequest is the request payload for Template data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type TemplateRequest struct {
	*Template
	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *TemplateRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	a.ProtectedID = ""                                 // unset the protected ID
	a.Template.Name = strings.ToLower(a.Template.Name) // as an example, we down-case
	return nil
}

// TemplateResponse is the response payload for the Template data model.
// See NOTE above in TemplateRequest as well.
//
// In the TemplateResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type TemplateResponse struct {
	*Template
	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func newTemplateResponse(template *Template) *TemplateResponse {
	return &TemplateResponse{Template: template}
}

func (rd *TemplateResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

type TemplateListResponse []*TemplateResponse

func newTemplateListResponse(templates []*Template) []render.Renderer {
	list := []render.Renderer{}
	for _, template := range templates {
		list = append(list, newTemplateResponse(template))
	}
	return list
}

func dbNewTemplate(template *Template) (int, error) {
	template.Id = rand.Intn(100) + 10
	templates = append(templates, template)
	return template.Id, nil
}

func dbGetTemplate(id int) (*Template, error) {
	for _, a := range templates {
		if a.Id == id {
			return a, nil
		}
	}
	return nil, errors.New("template not found")
}

func dbUpdateTemplate(id int, template *Template) (*Template, error) {
	for i, a := range templates {
		if a.Id == id {
			templates[i] = template
			return template, nil
		}
	}
	return nil, errors.New("template not found")
}

func dbRemoveTemplate(id int) (*Template, error) {
	for i, a := range templates {
		if a.Id == id {
			templates = append((templates)[:i], (templates)[i+1:]...)
			return a, nil
		}
	}
	return nil, errors.New("template not found")
}
