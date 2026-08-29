package main

import (
	"context"
	"log"
	"net/http"

	"github.com/tkdn/dynamodb-sort-uuid/dynamo"
	"github.com/tkdn/dynamodb-sort-uuid/handler"
)

func main() {
	ctx := context.Background()

	client, err := dynamo.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create dynamodb client: %v", err)
	}

	lister := dynamo.NewItemLister(client)
	itemsHandler := handler.NewItemsHandler(lister)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{user_id}/items", itemsHandler.List)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
