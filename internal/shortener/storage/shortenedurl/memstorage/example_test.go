package memstorage

import (
	"context"
	"fmt"

	"github.com/alukart32/shortener-url/internal/shortener/models"
)

func Example() {
	urls := []struct {
		userID string
		corrID string
		raw    string
		slug   string
	}{
		{
			userID: "1",
			corrID: "1",
			raw:    "http://demo.com",
			slug:   "slug1",
		},
		{
			userID: "2",
			corrID: "1",
			raw:    "http://demo.com",
			slug:   "slug2",
		},
		{
			userID: "3",
			corrID: "1",
			raw:    "http://demo.com",
			slug:   "slug3",
		},
	}

	repo := MemStorage()

	for _, v := range urls {
		shortenedURL := models.NewShortenedURL(
			v.userID,
			v.corrID,
			v.raw,
			v.slug,
			"http://localhost:8080/"+v.slug,
		)

		if err := repo.Save(context.TODO(), shortenedURL); err != nil {
			panic(err)
		}
	}

	collectedURLs, err := repo.CollectByUser(context.TODO(), "1")
	if err != nil {
		panic(err)
	}

	for _, v := range collectedURLs {
		fmt.Println(v.String())
	}

	// Output:
	// shortenedURL[userID: 1, corrID: 1, raw: http://demo.com, slug: slug1, value: http://localhost:8080/slug1, deleted: false]
}
