package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	client *mongo.Client
}

func newServer(c *mongo.Client) *Server {
	return &Server{
		client: c,
	}
}

func (s *Server) handleGetAllFacts(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("catfacts").Collection("facts")
	query := bson.M{}
	cursor, err := coll.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	results := []bson.M{}
	if err := cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type CatFactwWorker struct {
	client *mongo.Client
}

func NewCatFactWorker(c *mongo.Client) *CatFactwWorker {
	return &CatFactwWorker{
		client: c,
	}
} 

func (cfw *CatFactwWorker) start() error {
	coll := cfw.client.Database("catfacts").Collection("facts")
	ticket := time.NewTicker(2 * time.Second)

	for {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		var catFact bson.M
		if err := json.NewDecoder(resp.Body).Decode(&catFact); err != nil {
			return err
		}
		_, err = coll.InsertOne(context.TODO(), catFact)

		if err != nil {
			return err
		}

		<-ticket.C
	}
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	worker := NewCatFactWorker(client)
	go worker.start()
	server := newServer(client)
	http.HandleFunc("/facts", server.handleGetAllFacts)
	http.ListenAndServe(":3000", nil)
}
