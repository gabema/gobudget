package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	// RESTy routes for "bucketItems" resource
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

	http.ListenAndServe(":3000", r)
}
