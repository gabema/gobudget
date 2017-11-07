package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// go run main.go bucket.go bucketItem.go category.go errors.go template.go templateItem.go db.go utils.go
func main() {
	r := chi.NewRouter()

	r.Route("/bucketItems", func(r chi.Router) {
		r.Get("/", listBucketItems)
		r.Post("/", createBucketItem)       // POST /bucketItems
		r.Get("/search", searchBucketItems) // GET /bucketItems/search

		r.Route("/{bucketItemID}", func(r chi.Router) {
			r.Use(BucketItemCtx)            // Load the *BucketItem on the request context
			r.Get("/", getBucketItem)       // GET /bucketItems/123
			r.Put("/", updateBucketItem)    // PUT /bucketItems/123
			r.Delete("/", deleteBucketItem) // DELETE /bucketItems/123
		})

		// GET /bucketItems/whats-up
		r.With(BucketItemCtx).Get("/{articleSlug:[a-z-]+}", getBucketItem)
	})

	r.Route("/buckets", func(r chi.Router) {
		r.Get("/", listBuckets)
		r.Post("/", createBucket) // POST /buckets

		r.Route("/{bucketID}", func(r chi.Router) {
			r.Use(BucketCtx)            // Load the *Bucket on the request context
			r.Get("/", getBucket)       // GET /buckets/123
			r.Put("/", updateBucket)    // PUT /buckets/123
			r.Delete("/", deleteBucket) // DELETE /buckets/123
		})
	})

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", listCategories)
		r.Post("/", createCategory) // POST /categories

		r.Route("/{categoryID}", func(r chi.Router) {
			r.Use(CategoryCtx)            // Load the *Bucket on the request context
			r.Get("/", getCategory)       // GET /categories/123
			r.Put("/", updateCategory)    // PUT /categories/123
			r.Delete("/", deleteCategory) // DELETE /categories/123
		})
	})

	r.Route("/templates", func(r chi.Router) {
		r.Get("/", listTemplates)
		r.Post("/", createTemplate) // POST /templates

		r.Route("/{templateID}", func(r chi.Router) {
			r.Use(TemplateCtx)            // Load the *Template on the request context
			r.Get("/", getTemplate)       // GET /templates/123
			r.Put("/", updateTemplate)    // PUT /templates/123
			r.Delete("/", deleteTemplate) // DELETE /templates/123
		})
	})

	r.Route("/templateItems", func(r chi.Router) {
		r.Get("/", listTemplateItems)
		r.Post("/", createTemplateItem)       // POST /templateItems
		r.Get("/search", searchTemplateItems) // GET /templateItems/search

		r.Route("/{templateItemID}", func(r chi.Router) {
			r.Use(TemplateItemCtx)            // Load the *TemplateItem on the request context
			r.Get("/", getTemplateItem)       // GET /templateItems/123
			r.Put("/", updateTemplateItem)    // PUT /templateItems/123
			r.Delete("/", deleteTemplateItem) // DELETE /templateItems/123
		})

		// GET /templateItems/whats-up
		r.With(TemplateItemCtx).Get("/{articleSlug:[a-z-]+}", getTemplateItem)
	})

	r.Route("/db", func(r chi.Router) {
		r.Get("/create", func(w http.ResponseWriter, r *http.Request) {
			if err := dbCreateTables(); err != nil {
				render.Render(w, r, ErrInvalidRequest(err))
				return
			}
			render.Status(r, http.StatusCreated)
		})
		r.Get("/drop", func(w http.ResponseWriter, r *http.Request) {
			if err := dbDropTables(); err != nil {
				render.Render(w, r, ErrInvalidRequest(err))
				return
			}
			render.Status(r, http.StatusGone)
		})
		r.Get("/init", func(w http.ResponseWriter, r *http.Request) {
			dbInit()
			render.Status(r, http.StatusCreated)
		})
	})

	http.ListenAndServe(fmt.Sprintf(":%s", readEnvOrDefault("HTTP_PLATFORM_PORT", "3000")), r)
}
