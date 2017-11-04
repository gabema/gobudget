package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

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
		r.Post("/", createBucket) // POST /bucketItems

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

	http.ListenAndServe(":3000", r)
}
