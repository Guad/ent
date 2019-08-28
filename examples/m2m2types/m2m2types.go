// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/facebookincubator/ent/examples/m2m2types/ent/user"

	"github.com/facebookincubator/ent/examples/m2m2types/ent"
	"github.com/facebookincubator/ent/examples/m2m2types/ent/group"

	"github.com/facebookincubator/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "file:o2o2types?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()
	client := ent.NewClient(ent.Driver(db))
	ctx := context.Background()
	// run the auto migration tool.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	if err := Do(ctx, client); err != nil {
		log.Fatal(err)
	}
}

func Do(ctx context.Context, client *ent.Client) error {
	// Unlike `Save`, `SaveX` panics if an error occurs.
	hub := client.Group.
		Create().
		SetName("GitHub").
		SaveX(ctx)
	lab := client.Group.
		Create().
		SetName("GitLab").
		SaveX(ctx)
	a8m := client.User.
		Create().
		SetAge(30).
		SetName("a8m").
		AddGroups(hub, lab).
		SaveX(ctx)
	nati := client.User.
		Create().
		SetAge(28).
		SetName("nati").
		AddGroups(hub).
		SaveX(ctx)

	// Query the edges.
	groups, err := a8m.
		QueryGroups().
		All(ctx)
	if err != nil {
		return fmt.Errorf("querying a8m groups: %v", err)
	}
	fmt.Println(groups)
	// Output: [Group(id=1, name=GitHub) Group(id=2, name=GitLab)]

	groups, err = nati.
		QueryGroups().
		All(ctx)
	if err != nil {
		return fmt.Errorf("querying nati groups: %v", err)
	}
	fmt.Println(groups)
	// Output: [Group(id=1, name=GitHub)]

	// Traverse the graph.
	users, err := a8m.
		QueryGroups().                                           // [hub, lab]
		Where(group.Not(group.HasUsersWith(user.Name("nati")))). // [lab]
		QueryUsers().                                            // [a8m]
		QueryGroups().                                           // [hub, lab]
		QueryUsers().                                            // [a8m, nati]
		All(ctx)
	if err != nil {
		return fmt.Errorf("traversing the graph: %v", err)
	}
	fmt.Println(users)
	// Output: [User(id=1, age=30, name=a8m) User(id=2, age=28, name=nati)]
	return nil
}