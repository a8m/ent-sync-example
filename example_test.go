package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/a8m/ent-sync-example/ent"
	_ "github.com/a8m/ent-sync-example/ent/runtime"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

const bucketURL = "mem://photos/"

func Example_SyncCreate() {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		log.Fatal("failed opening bucket:", err)
	}
	client, err := ent.Open(
		dialect.SQLite,
		"file:ent?mode=memory&cache=shared&_fk=1",
		ent.Bucket(bucket),
	)
	if err != nil {
		log.Fatal("failed opening connection to sqlite:", err)
	}
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatal("failed creating schema resources:", err)
	}
	if err := client.User.Create().SetName("a8m").SetAvatarURL("a8m.png").Exec(ctx); err == nil {
		log.Fatal("expect user creation to fail because the image does not exist in the bucket")
	}
	if err := bucket.WriteAll(ctx, "a8m.png", []byte{255, 255, 255}, nil); err != nil {
		log.Fatalf("failed uploading image to the bucket: %v", err)
	}
	fmt.Printf("%q\n", keys(ctx, bucket))

	// User creation should pass as image was uploaded to the bucket.
	u := client.User.Create().SetName("a8m").SetAvatarURL("a8m.png").SaveX(ctx)

	// Deleting a user, should delete also its image from the bucket.
	client.User.DeleteOne(u).ExecX(ctx)
	fmt.Printf("%q\n", keys(ctx, bucket))

	// Output:
	// ["a8m.png"]
	// []
}

func keys(ctx context.Context, bucket *blob.Bucket) []string {
	var (
		keys []string
		iter = bucket.List(nil)
	)
	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		keys = append(keys, obj.Key)
	}
	return keys
}
