/*
Package filerepo defines a shortURL file repo.

GetBySlug, GetByURL, Collect, Save, Batch, Delete possible methods
for manipulating data:

	shortURL, err := repo.GetBySlug(ctx, "slug")
	...
	shortURL, err := repo.GetByURL(ctx, "url")
	...
	shortURLs, err := repo.CollectByUser(ctx, "userID")
	...
	err := repo.Save(ctx, shorturls.ShortURL{})
	...
	err := repo.Batch(ctx, []shorturls.ShortURL{})
	...
	err := repo.Delete("userID", []string)
	...

For creation a new FileRepo the caller should pass filepath.
If filepath is empty, then it will be set from env param:

	repo, err := New(filepath)

The caller must close the repository when finished with it:

	repo.Close()

# File reader and writer.

To be able to write and read to a file, interfaces are defined in various ways: writer, reader.

The current implementation uses the MessagePack format for encoding data.
*/
package filestorage
