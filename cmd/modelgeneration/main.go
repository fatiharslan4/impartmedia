package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/jsonschema"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/tags"
)

var topLevelModels = map[string]interface{}{
	"Profile":               models.Profile{},
	"Hive":                  models.Hive{},
	"Hives":                 models.Hives{},
	"Post":                  models.Post{},
	"Posts":                 models.Posts{},
	"Comment":               models.Comment{},
	"Comments":              models.Comments{},
	"PostCommentTrack":      models.PostCommentTrack{},
	"PagedPostsResponse":    models.PagedPostsResponse{},
	"PagedCommentsResponse": models.PagedCommentsResponse{},
	"Tags":                  tags.Tags{},
	"TagComparisons":        tags.TagComparisons{},
}

func main() {

	var serializedSchema []byte
	var err error
	for k, v := range topLevelModels {
		fmt.Printf("generating json schema for %s\n", k)
		schema := jsonschema.Reflect(v)
		schema.Title = k

		serializedSchema, err = json.MarshalIndent(schema, "", "\t")
		if err != nil {
			panic(err)
		}

		func() {
			fileName := fmt.Sprintf("./schemas/json/%s.json", k)
			f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			fmt.Println("writing ", fileName)
			_, err = f.Write(serializedSchema)
			if err != nil {
				panic(err)
			}
		}()
	}

}
