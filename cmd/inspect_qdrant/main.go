package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdrant/go-client/qdrant"
)

func main() {
	host := "192.168.1.21"
	port := 6334
	collection := "default"
	limit := 20

	if len(os.Args) > 1 {
		collection = os.Args[1]
	}
	if len(os.Args) > 2 {
		host = os.Args[2]
	}

	client, err := qdrant.NewClient(&qdrant.Config{Host: host, Port: port})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	pointsClient := client.GetPointsClient()

	result, err := pointsClient.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: collection,
		Limit:          qdrant.PtrOf(uint32(limit)),
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		panic(err)
	}

	for i, point := range result.Result {
		content := ""
		userID := "?"
		source := "?"
		docID := "?"

		if v, ok := point.Payload["page_content"]; ok {
			if s := v.GetStringValue(); s != "" {
				content = s
			}
		}
		if v, ok := point.Payload["user_id"]; ok {
			userID = v.GetStringValue()
		}
		if v, ok := point.Payload["source"]; ok {
			source = v.GetStringValue()
		}
		if v, ok := point.Payload["document_id"]; ok {
			docID = v.GetStringValue()
		}

		trunc := content
		if len(trunc) > 120 {
			trunc = trunc[:120] + "..."
		}

		fmt.Printf("--- Point %d (id=%v) ---\n", i+1, point.Id)
		fmt.Printf("  user_id: %s  doc_id: %s  source: %s\n", userID, docID, source)
		fmt.Printf("  content: %s\n\n", trunc)
	}
	fmt.Printf("Total points returned: %d\n", len(result.Result))
}
