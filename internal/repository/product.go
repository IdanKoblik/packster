package repository

import (
	"packster/internal/utils"
	"packster/pkg/config"
	"packster/pkg/types"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type IProductRepo interface {
	CreateProduct(product *types.Product) error
	DeleteProduct(name, token string, admin bool) error
	FetchProduct(name string) (*types.Product, error)
	DeleteToken(productName, sourceToken, targetToken string, admin bool) error
	AddToken(productName, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error
	DeleteVersion(productName, version, token string, admin bool) error
	AddVersion(productName, version, token string, admin bool, v types.Version) error
	ListProducts() ([]string, error)
	ListProductsByToken(hashedToken string) ([]string, error)
}

type ProductRepository struct {
	MongoClient   *mongo.Client
	MongoDatabase *mongo.Database
	Cfg           *config.Config
}

func NewProductRepository(mongoClient *mongo.Client, cfg *config.Config) *ProductRepository {
	return &ProductRepository{
		MongoClient:   mongoClient,
		MongoDatabase: mongoClient.Database(cfg.Mongo.Database),
		Cfg:           cfg,
	}
}

func (r *ProductRepository) ListProducts() ([]string, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var products []types.Product
	if err := cursor.All(context.Background(), &products); err != nil {
		return nil, err
	}

	names := make([]string, len(products))
	for i, p := range products {
		names[i] = p.Name
	}

	return names, nil
}

func (r *ProductRepository) ListProductsByToken(hashedToken string) ([]string, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"tokens." + hashedToken: bson.M{"$exists": true}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var products []types.Product
	if err := cursor.All(context.Background(), &products); err != nil {
		return nil, err
	}

	names := make([]string, len(products))
	for i, p := range products {
		names[i] = p.Name
	}

	return names, nil
}

func (r *ProductRepository) CreateProduct(product *types.Product) error {
	existing, err := r.FetchProduct(product.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		return errors.New("product already exists")
	}

	product.HashTokens()

	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, product)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) DeleteProduct(name, token string, admin bool) error {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product, err := r.FetchProduct(name)
	if err != nil {
		return err
	}

	permissions := product.Tokens[utils.Hash(token)]
	if !admin && !permissions.Maintainer && !permissions.Delete {
		return errors.New("missing delete permission")
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": name})
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) FetchProduct(name string) (*types.Product, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product types.Product
	err := collection.FindOne(ctx, bson.M{"_id": name}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) DeleteToken(productName, sourceToken, targetToken string, admin bool) error {
	product, err := r.FetchProduct(productName)
	if err != nil {
		return err
	}

	if product == nil {
		return errors.New("product not found")
	}

	permissions := product.Tokens[utils.Hash(sourceToken)]
	if !admin && !permissions.Maintainer {
		return errors.New("missing maintainer permission")
	}

	hashedToken := utils.Hash(targetToken)
	delete(product.Tokens, hashedToken)

	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.UpdateOne(ctx, bson.M{"_id": productName}, bson.M{"$set": product})
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) AddToken(productName, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error {
	product, err := r.FetchProduct(productName)
	if err != nil {
		return err
	}

	if product == nil {
		return errors.New("product not found")
	}

	tokenPermissions := product.Tokens[utils.Hash(sourceToken)]
	if !admin && !tokenPermissions.Maintainer {
		return errors.New("missing maintainer permission")
	}

	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product.Tokens[utils.Hash(targetToken)] = permissions

	_, err = collection.UpdateOne(ctx, bson.M{"_id": productName}, bson.M{"$set": product})
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) DeleteVersion(productName, version, token string, admin bool) error {
	product, err := r.FetchProduct(productName)
	if err != nil {
		return err
	}

	if product == nil {
		return errors.New("product not found")
	}

	permissions := product.Tokens[utils.Hash(token)]
	if !admin && !permissions.Maintainer && !permissions.Delete {
		return errors.New("missing maintainer / delete permission")
	}

	delete(product.Versions, version)
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.UpdateOne(ctx, bson.M{"_id": productName}, bson.M{"$set": product})
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) AddVersion(productName, version, token string, admin bool, v types.Version) error {
	product, err := r.FetchProduct(productName)
	if err != nil {
		return err
	}

	if product == nil {
		return errors.New("product not found")
	}

	permissions := product.Tokens[utils.Hash(token)]
	if !admin && !permissions.Upload {
		return errors.New("missing upload permission")
	}

	if _, ok := product.Versions[version]; ok {
		return errors.New("version already exists")
	}

	product.Versions[version] = v
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.ProductCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.UpdateOne(ctx, bson.M{"_id": productName}, bson.M{"$set": product})
	if err != nil {
		return err
	}

	return nil
}
