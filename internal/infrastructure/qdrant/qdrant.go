package qdrant

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/schema"
)

type QdrantAdapter struct {
	client *qdrant.Client
}

func NewQdrantAdapter(host string, port int) (*QdrantAdapter, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating qdrant client: %w", err)
	}
	return &QdrantAdapter{client: client}, nil
}

func (qa *QdrantAdapter) CreateCollection(ctx context.Context, name string, size int) error {
	collections := qa.client.GetCollectionsClient()
	exists, err := collections.CollectionExists(ctx, &qdrant.CollectionExistsRequest{
		CollectionName: name,
	})
	if err != nil {
		return fmt.Errorf("error verifying collection: %w", err)
	}
	if exists.GetResult().GetExists() {
		return nil
	}

	_, err = collections.Create(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(size),
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("error creating collection: %w", err)
	}
	return nil
}

func (qa *QdrantAdapter) UpsertDocuments(
	ctx context.Context,
	collection string,
	docs []schema.Document,
	vectors [][]float32,
) error {
	points := make([]*qdrant.PointStruct, len(docs))
	for i, doc := range docs {
		id := doc.Metadata["_id"]
		var pointId uint64
		if id != nil {
			fmt.Sscanf(fmt.Sprintf("%v", id), "%d", &pointId)
		} else {
			pointId = uint64(i)
		}

		points[i] = &qdrant.PointStruct{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Num{
					Num: pointId,
				},
			},
			Payload: formatPayload(doc),
			Vectors: qdrant.NewVectorsDense(vectors[i]),
		}
	}

	pointsClient := qa.client.GetPointsClient()
	_, err := pointsClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("error haciendo upsert: %w", err)
	}
	return nil

}

func (qa *QdrantAdapter) Search(ctx context.Context, collection string, queryVector []float32, limit int) ([]schema.Document, error) {
	pointsClient := qa.client.GetPointsClient()
	searchResult, err := pointsClient.Search(ctx, &qdrant.SearchPoints{
		CollectionName: collection,
		Vector:         queryVector,
		Limit:          uint64(limit),
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error searching: %w", err)
	}

	docs := make([]schema.Document, len(searchResult.Result))
	for i, result := range searchResult.Result {
		docs[i] = schema.Document{
			PageContent: extractPageContent(result.Payload),
			Metadata:    extractMetadata(result.Payload),
		}
	}
	return docs, nil
}

func (qa *QdrantAdapter) GetPointsCount(ctx context.Context, collection string) (uint64, error) {
	collections := qa.client.GetCollectionsClient()
	info, err := collections.Get(ctx, &qdrant.GetCollectionInfoRequest{
		CollectionName: collection,
	})
	if err != nil {
		return 0, nil
	}
	if info.GetResult() == nil {
		return 0, nil
	}
	return info.GetResult().GetPointsCount(), nil
}

func formatPayload(doc schema.Document) map[string]*qdrant.Value {
	payload := make(map[string]*qdrant.Value)
	payload["page_content"] = &qdrant.Value{
		Kind: &qdrant.Value_StringValue{
			StringValue: doc.PageContent,
		},
	}
	for k, v := range doc.Metadata {
		payload[k] = &qdrant.Value{
			Kind: &qdrant.Value_StringValue{
				StringValue: fmt.Sprintf("%v", v),
			},
		}
	}
	return payload
}

func extractPageContent(payload map[string]*qdrant.Value) string {
	if v, ok := payload["page_content"]; ok {
		if str, ok := v.GetKind().(*qdrant.Value_StringValue); ok {
			return str.StringValue
		}
	}
	return ""
}

func extractMetadata(payload map[string]*qdrant.Value) map[string]any {
	metadata := make(map[string]any)
	for k, v := range payload {
		if k == "page_content" {
			continue
		}
		if str, ok := v.GetKind().(*qdrant.Value_StringValue); ok {
			metadata[k] = str.StringValue
		}
	}
	return metadata
}

func ptrOf[T any](v T) *T {
	return &v
}
