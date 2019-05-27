package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/rules"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// errNotFound implies the error for document not found.
var errNotFound = errors.New("not found")

// dataSnap defines the response for retrieving data.
type dataSnap struct {
	val     interface{}
	noIndex bool
}

type client struct {
	*mongo.Client
	rules     rules.Rules
	database  string
	collTable string
}

func (c *client) Close() error {
	return c.Disconnect(context.Background())
}

// Database gets the database in mongodb.
func (c *client) Database() *mongo.Database {
	return c.Client.Database(c.database)
}

// CollectionTable gets the collection mapping table in mongodb.
func (c *client) CollectionTable() *mongo.Collection {
	return c.Database().Collection(c.collTable)
}

func (c *client) Set(ctx context.Context, ref string, data interface{}) error {
	// try to update the document field.
	if err := c.updateAncestor(ctx, ref, data); err == errNotFound {
		// fallthrough.
	} else if err != nil {
		return fmt.Errorf("failed to update ancestor to %s: %v", ref, err)
	} else {
		return nil
	}

	// split and validate the ref.
	paths := strings.Split(ref, "/")
	lenPaths := len(paths)
	if lenPaths < 3 {
		return fmt.Errorf("invalid ref %s, should be /coll/id/{field}", ref)
	}

	// detect the data's type, use strings.Join to guarantee first slash.
	coll, id := strings.Join(paths[:lenPaths-1], "/"), paths[lenPaths-1]

	if v, ok := data.(map[string]interface{}); ok && c.canInsert(coll, id) {
		// insert or replace a document.
		if err := c.insertDocument(ctx, coll, id, v); err != nil {
			return fmt.Errorf("failed to insert document to collection %s: %v", coll, err)
		}
	} else if data == nil {
		// delete the document.
		if _, err := c.Database().Collection(hash(coll)).DeleteOne(ctx, bson.M{"_id": id}); err != nil {
			return fmt.Errorf("failed to delete document %s in collections %s: %v", id, coll, err)
		}
	} else {
		// try to insert a primary into a document..
		if lenPaths < 4 {
			return fmt.Errorf("cannot insert document in %s with primary %+v", ref, v)
		}
		// use strings.Join to guarantee first slash.
		coll, id, field := strings.Join(paths[:lenPaths-2], "/"), paths[lenPaths-2], paths[lenPaths-1]

		// check security rule.
		if !c.canInsert(coll, id) {
			return fmt.Errorf("cannot insert to %s", coll)
		}

		// insert a parent document to contain the primary field.
		document := map[string]interface{}{field: data}
		if err := c.insertDocument(ctx, coll, id, document); err != nil {
			return fmt.Errorf("failed to insert document to collection %s: %v", coll, err)
		}
	}

	return nil
}

func (c *client) Get(ctx context.Context, ref string, query data.Query) (interface{}, error) {
	// find exactly the same collection first.
	if resp, err := c.getFromRef(ctx, ref, query); err == errNotFound {
		// fallthrough
	} else if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", ref, err)
	} else if resp != nil {
		return resp.val, nil
	}

	// find from ancestor collections.
	if resp, err := c.getFromAncestor(ctx, ref, query); err == errNotFound {
		// fallthrough
	} else if err != nil {
		return nil, fmt.Errorf("failed to get %s from ancestor: %v", ref, err)
	} else if resp != nil {
		return resp.val, nil
	}

	// find from descendant collections.
	resp, err := c.getFromDescendant(ctx, ref, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s from descendant: %v", ref, err)
	} else if resp != nil {
		return resp.val, nil
	}
	return nil, nil
}

func (c *client) Reset(ctx context.Context) error {
	return c.Database().Drop(ctx)
}

// updateAncestor tries to update the field of existed document,
// return errNotFound if no matched document.
func (c *client) updateAncestor(ctx context.Context, ref string, data interface{}) error {
	// generate possible ancestor collection paths.
	subPath, subPaths := "/", []string{}
	for _, p := range strings.Split(ref, "/") {
		subPath = path.Join(subPath, p)
		subPaths = append(subPaths, subPath)
	}

	// try find one ancestor collection.
	var collection bson.M
	if err := c.Database().Collection(c.collTable).FindOne(ctx, bson.M{"_id": bson.M{"$in": subPaths}}).Decode(&collection); err == mongo.ErrNoDocuments {
		// ancestor collection not found.
		return errNotFound
	} else if err != nil {
		return fmt.Errorf("failed to find ancestor collection %s: %v", ref, err)
	}

	// ancestor collection found.
	coll, hash := collection["_id"].(string), collection["hash"].(string)
	rel, err := filepath.Rel(coll, ref)
	if err != nil {
		return fmt.Errorf("failed to find relative path of %s and %s", coll, ref)
	}
	paths := strings.Split(rel, "/")
	id := paths[0]
	field := strings.Join(paths[1:], ".")
	if field == "" {
		// nothing to update, skip.
		return errNotFound
	}

	// compose the update by relative path.
	update := bson.M{"$set": bson.M{field: data}}
	if data == nil {
		update = bson.M{"$unset": bson.M{field: ""}}
	}

	// update the field in ancestor document.
	result, err := c.Database().Collection(hash).UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update document %s in collection %s: %v", id, coll, err)
	} else if result.MatchedCount == 0 {
		return errNotFound
	}

	return nil
}

func (c *client) insertDocument(ctx context.Context, coll, id string, data interface{}) error {
	// insert the document to collection.
	hash := hash(coll)
	if _, err := c.Database().Collection(hash).
		ReplaceOne(ctx, bson.M{"_id": id}, data, options.Replace().SetUpsert(true)); err != nil {
		return fmt.Errorf("failed to replace document %s in collection %s: %v", id, coll, err)
	}

	// insert the collection entry in collection table.
	result, err := c.CollectionTable().ReplaceOne(ctx, bson.M{"_id": coll}, bson.M{"hash": hash}, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to update collection table %s: %v", coll, err)
	}

	// if the collection is inserted first time, create index for the collection.
	if indexes := c.rules.Indexes(); len(indexes) > 0 && result.UpsertedCount > 0 {
		go c.createIndex(context.Background(), coll, indexes)
	}

	return nil
}

func (c *client) getFromRef(ctx context.Context, ref string, query data.Query) (*dataSnap, error) {
	var collection bson.M
	if err := c.CollectionTable().FindOne(ctx, bson.M{"_id": ref}).Decode(&collection); err == mongo.ErrNoDocuments {
		// ancestor collection not found.
		return nil, errNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to find ref collection %s: %v", ref, err)
	}

	return c.getWithQuery(ctx, ref, query)
}

func (c *client) getFromAncestor(ctx context.Context, ref string, query data.Query) (*dataSnap, error) {
	// generate possible ancestor collection paths.
	subPath, subPaths := "/", []string{}
	for _, p := range strings.Split(ref, "/") {
		subPath = path.Join(subPath, p)
		subPaths = append(subPaths, subPath)
	}

	// try find one ancestor collection.
	var collection bson.M
	if err := c.CollectionTable().FindOne(ctx, bson.M{"_id": bson.M{"$in": subPaths}}).Decode(&collection); err == mongo.ErrNoDocuments {
		// ancestor collection not found.
		return nil, errNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to find ancestor collection %s: %v", ref, err)
	}

	// ancestor collection found.
	coll := collection["_id"].(string)
	rel, err := filepath.Rel(coll, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to find relative path of %s and %s", coll, ref)
	}
	paths := strings.Split(rel, "/")
	id := paths[0]

	// TODO: apply query to avoid full document search.
	// get the document and put into response
	data, err := c.getWithQuery(ctx, coll, data.Query{
		OrderBy: ".key",
		StartAt: id,
		EndAt:   id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get document %s in collection %s: %v", id, coll, err)
	} else if data.val == nil {
		return nil, errNotFound
	}

	m, ok := data.val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert ancestor document to a map: %+v", data)
	}

	// iterate the relative paths to compose the response.
	for i, p := range paths {
		if p == "" {
			// leading space.
			continue
		}
		if i == len(paths)-1 {
			// trailing branch.
			return &dataSnap{val: m[p]}, nil
		}

		// go deeper to search.
		if child, ok := m[p].(map[string]interface{}); ok {
			m = child
		} else {
			break
		}
	}
	return &dataSnap{val: m}, nil
}

func (c *client) getFromDescendant(ctx context.Context, ref string, query data.Query) (*dataSnap, error) {
	// generate the filter by reference and query.
	regex := fmt.Sprintf("^%s/", ref)
	if ref == "/" {
		regex = "^/"
	}
	if query.Shallow {
		regex = regex + ".+$"
	}

	// find matched collections.
	cursor, err := c.CollectionTable().Find(ctx, bson.M{"_id": bson.M{"$regex": regex}})
	if err != nil {
		return nil, fmt.Errorf("failed to find descendant collections for %s: %v", ref, err)
	}
	defer cursor.Close(ctx)

	// iterate all collections.
	var collections []string
	for cursor.Next(ctx) {
		var collection bson.M
		if err := cursor.Decode(&collection); err != nil {
			return nil, fmt.Errorf("failed to decode collection: %v", err)
		}
		collections = append(collections, collection["_id"].(string))
	}

	// return if no collection found.
	if len(collections) == 0 {
		return &dataSnap{}, nil
	}

	// directly return if shallow query.
	if query.Shallow {
		resp := map[string]interface{}{}
		for _, collection := range collections {
			rel, err := filepath.Rel(ref, collection)
			if err != nil {
				return nil, fmt.Errorf("failed to find relative path for %s and %s: %v", ref, collection, err)
			}
			resp[strings.Split(rel, "/")[0]] = true
		}
		return &dataSnap{val: resp}, nil
	}

	// get all documents from each collections.
	// TODO: get data in parallel.
	resp := map[string]interface{}{}
	for _, collection := range collections {
		// gets data from each collection.
		data, err := c.getWithQuery(ctx, collection, data.Query{})
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %v", collection, err)
		}

		// update the final resp with fetched data.
		rel, err := filepath.Rel(ref, collection)
		if err != nil {
			return nil, fmt.Errorf("failed to find relative path of %s and %s", ref, collection)
		}
		paths := strings.Split(rel, "/")
		m := resp
		for i, p := range paths {
			if p == "" {
				// leading space.
				continue
			}

			// trailing branch.
			if i == len(paths)-1 {
				m[p] = data.val
				break
			}

			// handle if branch not exists.
			if _, ok := m[p].(map[string]interface{}); !ok {
				m[p] = map[string]interface{}{}
			}

			// move the pointer to child.
			m = m[p].(map[string]interface{})
		}
	}

	return &dataSnap{val: resp}, nil
}

func (c *client) getWithQuery(ctx context.Context, ref string, query data.Query) (*dataSnap, error) {
	// compose a filter based on query.
	filter, index := bson.M{}, ""
	switch query.OrderBy {
	case "$key", ".key":
		index = "_id"
	case "$value", ".value":
		// TODO: not supported now.
	case "":
	default:
		// for nested query, replace the path divider by delimiter.
		index = strings.Replace(query.OrderBy, "/", ".", -1)
	}
	if index != "" {
		param, param2 := bson.M{}, bson.M{}
		if query.StartAt != nil {
			param["$gte"] = query.StartAt
		}
		if query.EndAt != nil {
			param["$lte"] = query.EndAt
		}
		if query.StartKey != "" {
			param2["$gte"] = query.StartKey
		}
		if query.EndKey != "" {
			param2["$lte"] = query.EndKey
		}
		if len(param) != 0 {
			filter[index] = param
		}
		if len(param2) != 0 {
			// if oderByKey and startKey both exist, apply startKey
			filter["_id"] = param2
		}
	}

	var option *options.FindOptions
	if query.Limit > 0 {
		sortOrder := 1
		if query.LimitOrder == "r" {
			sortOrder = -1
		}
		// if index equals, sort by key.
		sort := bson.D{}
		if index != "" && index != "_id" {
			sort = append(sort, bson.E{Key: index, Value: sortOrder})
		}
		sort = append(sort, bson.E{Key: "_id", Value: sortOrder})
		option = options.Find().SetSort(sort).SetLimit(int64(query.Limit))
	}

	// find the documents with query.
	cursor, err := c.Database().Collection(hash(ref)).Find(ctx, filter, option)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", ref, err)
	}
	defer cursor.Close(ctx)

	// iterate the matched documents.
	var documents interface{}
	for cursor.Next(ctx) {
		var document map[string]interface{}
		if err := cursor.Decode(&document); err != nil {
			return nil, fmt.Errorf("failed to decode: %v", err)
		}
		if documents == nil {
			documents = map[string]interface{}{}
		}
		var value interface{} = document
		if query.Shallow {
			// handle shallow query.
			value = true
		}
		documents.(map[string]interface{})[document["_id"].(string)] = value
		// delete the _id in document for response.
		delete(document, "_id")
	}

	// TODO: check no index by security rule.

	return &dataSnap{
		val: documents,
	}, nil
}

func (c *client) createIndex(ctx context.Context, ref string, indexes []string) {
	// create indexes for collection.
	indexModels := make([]mongo.IndexModel, len(indexes))
	for i, index := range indexes {
		indexModels[i] = mongo.IndexModel{
			Keys: bson.M{index: 1},
		}
	}
	result, err := c.Database().Collection(hash(ref)).Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Printf("failed to create index for %s: %v", ref, err)
		return
	}
	log.Printf("index %+v created for %s", result, ref)
}
