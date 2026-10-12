package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client *mongo.Client
	coll   *mongo.Collection
}

type SignalDoc struct {
	TS      time.Time      `bson:"ts" json:"ts"`
	Ticker  string         `bson:"ticker" json:"ticker"`
	Signal  string         `bson:"signal" json:"signal"`
	Analyst string         `bson:"analyst" json:"analyst"`
	Email   string         `bson:"email" json:"email"`
	PDFFile string         `bson:"pdf_file" json:"pdf_file"`
	PDFData []byte         `bson:"pdf_data,omitempty" json:"-"`
	Source  string         `bson:"source" json:"source"`
	States  map[string]any `bson:"states" json:"states"`
}

func NewMongoClient(uri string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Verify connection immediately — catches Atlas IP whitelist issues
	if err := c.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed (check Atlas Network Access whitelist): %w", err)
	}

	// DB: smif_quant  Collection: signals
	coll := c.Database("smif_quant").Collection("signals")
	return &MongoClient{client: c, coll: coll}, nil
}

func (m *MongoClient) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.client.Disconnect(ctx)
}

func (m *MongoClient) InsertSignal(doc SignalDoc) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_, err := m.coll.InsertOne(ctx, bson.M{
		"ts":       doc.TS,
		"ticker":   doc.Ticker,
		"signal":   doc.Signal,
		"analyst":  doc.Analyst,
		"email":    doc.Email,
		"pdf_file": doc.PDFFile,
		"pdf_data": doc.PDFData,
		"source":   doc.Source,
		"states":   doc.States,
	})
	return err
}

func (m *MongoClient) GetPDFData(pdfFile string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result struct {
		PDFData []byte `bson:"pdf_data"`
	}
	err := m.coll.FindOne(ctx, bson.M{"pdf_file": pdfFile}).Decode(&result)
	return result.PDFData, err
}

// GetSignals retrieves signals with optional filters. Pass limit=0 for all records.
func (m *MongoClient) GetSignals(email, ticker, signal, dateFrom, dateTo string, limit int64) ([]SignalDoc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "ts", Value: -1}})
	if limit > 0 {
		findOptions.SetLimit(limit)
	}

	filter := bson.M{}
	if email != "" {
		filter["$or"] = []bson.M{
			{"email": email},
			{"analyst": email},
		}
	}
	if ticker != "" {
		filter["ticker"] = strings.ToUpper(ticker)
	}
	if signal != "" {
		filter["signal"] = bson.M{"$regex": signal, "$options": "i"}
	}
	if dateFrom != "" || dateTo != "" {
		dateFilter := bson.M{}
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			dateFilter["$gte"] = t
		}
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			dateFilter["$lte"] = t.Add(24 * time.Hour)
		}
		if len(dateFilter) > 0 {
			filter["ts"] = dateFilter
		}
	}

	cursor, err := m.coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	signals := []SignalDoc{}
	for cursor.Next(ctx) {
		var doc SignalDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		signals = append(signals, doc)
	}
	return signals, nil
}
