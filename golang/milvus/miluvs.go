package milvus

import (
	"context"
	"errors"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"strings"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// 默认设置
const (
	MaxLimitSize = 10000

	CollectionNameSQL = "c2sql"
	CollectionNameDDL = "c2ddl"
	CollectionNameDoc = "c2doc"
)

type MilvusVectorStore struct {
	indexer       *milvus.Indexer
	milvusClient  client.Client
	embedding     embedding.Embedder
	embeddingDim  int
	nResults      int
	collectionSQL string
	collectionDDL string
	collectionDoc string
}

type Config struct {
	MilvusClient client.Client
	Embedding    embedding.Embedder
	NResults     int
}

type DefaultEmbedding struct{}

func (d *DefaultEmbedding) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embeddings[i] = make([]float64, 128)
		for j := range embeddings[i] {
			embeddings[i][j] = 0.1 * float64(j%10)
		}
	}
	return embeddings, nil
}

// NewMilvusVectorStore 创建新的MilvusVectorStore实例
func NewMilvusVectorStore(ctx context.Context, config *Config) (*MilvusVectorStore, error) {
	if config == nil {
		config = &Config{}
	}

	store := &MilvusVectorStore{
		nResults:      10,
		collectionSQL: CollectionNameSQL,
		collectionDDL: CollectionNameDDL,
		collectionDoc: CollectionNameDoc,
	}

	// 设置结果数量
	if config.NResults > 0 {
		store.nResults = config.NResults
	}

	// 设置Milvus客户端
	if config.MilvusClient != nil {
		store.milvusClient = config.MilvusClient
	} else {
		// 创建默认客户端
		cli, err := client.NewClient(ctx, client.Config{
			Address: DefaultMilvusURI,
		})
		if err != nil {
			return nil, fmt.Errorf("创建Milvus客户端失败: %v", err)
		}
		store.milvusClient = cli
	}

	// 设置嵌入模型
	if config.Embedding != nil {
		store.embedding = config.Embedding
	} else {
		store.embedding = &DefaultEmbedding{}
	}

	// 创建Eino Milvus索引器
	indexerConfig := &milvus.IndexerConfig{
		Client:    store.milvusClient,
		Embedding: store.embedding,
	}

	indexer, err := milvus.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Milvus索引器失败: %v", err)
	}
	store.indexer = indexer.(*milvus.Indexer)

	// 创建所需集合
	if err := store.CreateCollections(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *MilvusVectorStore) CreateCollections(ctx context.Context) error {
	has, err := s.milvusClient.HasCollection(ctx, s.collectionSQL)
	if err != nil {
		return err
	}
	if !has {
		if err := s.createSQLCollection(ctx); err != nil {
			return err
		}
	}

	has, err = s.milvusClient.HasCollection(ctx, s.collectionDDL)
	if err != nil {
		return err
	}
	if !has {
		if err := s.createDDLCollection(ctx); err != nil {
			return err
		}
	}

	has, err = s.milvusClient.HasCollection(ctx, s.collectionDoc)
	if err != nil {
		return err
	}
	if !has {
		if err := s.createDocCollection(ctx); err != nil {
			return err
		}
	}

	return nil
}

// createSQLCollection 创建SQL集合
func (s *MilvusVectorStore) createSQLCollection(ctx context.Context) error {
	// 使用Eino的Milvus索引器创建集合
	return s.indexer.CreateCollection(ctx, s.collectionSQL)
}

// createDDLCollection 创建DDL集合
func (s *MilvusVectorStore) createDDLCollection(ctx context.Context) error {
	// 使用Eino的Milvus索引器创建集合
	return s.indexer.CreateCollection(ctx, s.collectionDDL)
}

// createDocCollection 创建文档集合
func (s *MilvusVectorStore) createDocCollection(ctx context.Context) error {
	return s.indexer.CreateCollection(ctx, s.collectionDoc)
}

// AddQuestionSQL 添加问题和对应的SQL语句
func (s *MilvusVectorStore) AddQuestionSQL(ctx context.Context, question, sql string) (string, error) {
	if question == "" || sql == "" {
		return "", errors.New("问题和SQL不能为空")
	}

	id := uuid.New().String() + "-sql"

	// 创建文档对象
	doc := &schema.Document{
		ID:      id,
		Content: question,
		MetaData: map[string]any{
			"sql": sql,
		},
	}

	// 存储文档
	_, err := s.indexer.Store(ctx, []*schema.Document{doc})
	if err != nil {
		return "", err
	}

	return id, nil
}

// AddDDL 添加DDL语句
func (s *MilvusVectorStore) AddDDL(ctx context.Context, ddl string) (string, error) {
	if ddl == "" {
		return "", errors.New("DDL不能为空")
	}

	id := uuid.New().String() + "-ddl"

	// 创建文档对象
	doc := &schema.Document{
		ID:      id,
		Content: ddl,
	}

	// 存储文档
	_, err := s.indexer.Store(ctx, []*schema.Document{doc})
	if err != nil {
		return "", err
	}

	return id, nil
}

// AddDocumentation 添加文档
func (s *MilvusVectorStore) AddDocumentation(ctx context.Context, documentation string) (string, error) {
	if documentation == "" {
		return "", errors.New("文档不能为空")
	}

	id := uuid.New().String() + "-doc"

	// 创建文档对象
	doc := &schema.Document{
		ID:      id,
		Content: documentation,
	}

	// 存储文档
	_, err := s.indexer.Store(ctx, []*schema.Document{doc})
	if err != nil {
		return "", err
	}

	return id, nil
}

// QuestionSQL
type QuestionSQL struct {
	Question string `json:"question"`
	SQL      string `json:"sql"`
}

// GetSimilarQuestionSQL
func (s *MilvusVectorStore) GetSimilarQuestionSQL(ctx context.Context, question string) ([]QuestionSQL, error) {
	docs, err := s.indexer.Search(ctx, question, s.nResults)
	if err != nil {
		return nil, err
	}

	var results []QuestionSQL
	for _, doc := range docs {
		if strings.HasSuffix(doc.ID, "-sql") {
			sql, ok := doc.MetaData["sql"].(string)
			if !ok {
				continue
			}

			results = append(results, QuestionSQL{
				Question: doc.Content,
				SQL:      sql,
			})
		}
	}

	return results, nil
}

// GetRelatedDDL 获取相关的DDL语句
func (s *MilvusVectorStore) GetRelatedDDL(ctx context.Context, question string) ([]string, error) {
	// 使用Eino的Milvus索引器进行相似性搜索
	docs, err := s.indexer.Search(ctx, question, s.nResults)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, doc := range docs {
		if strings.HasSuffix(doc.ID, "-ddl") {
			results = append(results, doc.Content)
		}
	}

	return results, nil
}

// GetRelatedDocumentation
func (s *MilvusVectorStore) GetRelatedDocumentation(ctx context.Context, question string) ([]string, error) {
	docs, err := s.indexer.Search(ctx, question, s.nResults)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, doc := range docs {
		if strings.HasSuffix(doc.ID, "-doc") {
			results = append(results, doc.Content)
		}
	}

	return results, nil
}

// TrainingData 训练数据结构
type TrainingData struct {
	ID       string
	Question string
	Content  string
}

// GetTrainingData 获取所有训练数据
func (s *MilvusVectorStore) GetTrainingData(ctx context.Context) ([]TrainingData, error) {
	// 这个功能在Eino框架中可能需要自定义实现
	// 暂时简化实现
	var allDocs []*schema.Document

	// 获取SQL数据
	sqlDocs, err := s.indexer.Search(ctx, "", MaxLimitSize)
	if err != nil {
		return nil, err
	}
	allDocs = append(allDocs, sqlDocs...)

	var trainingData []TrainingData
	for _, doc := range allDocs {
		data := TrainingData{
			ID: doc.ID,
		}

		if strings.HasSuffix(doc.ID, "-sql") {
			data.Question = doc.Content
			if sql, ok := doc.MetaData["sql"].(string); ok {
				data.Content = sql
			}
		} else if strings.HasSuffix(doc.ID, "-ddl") || strings.HasSuffix(doc.ID, "-doc") {
			data.Content = doc.Content
		}

		trainingData = append(trainingData, data)
	}

	return trainingData, nil
}

// RemoveTrainingData
func (s *MilvusVectorStore) RemoveTrainingData(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, errors.New("ID不能为空")
	}

	err := s.indexer.Delete(ctx, []string{id})
	if err != nil {
		return false, err
	}

	return true, nil
}

// Close 关闭连接
func (s *MilvusVectorStore) Close() error {
	if s.milvusClient != nil {
		return s.milvusClient.Close()
	}
	return nil
}
