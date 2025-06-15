/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package milvus

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/joho/godotenv"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

func TestMilvusVectorStore(t *testing.T) {
	// 加载环境变量
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Logf("无法加载.env文件: %v, 将使用默认URI", err)
	}

	// 获取Milvus地址，如果未设置则使用默认值
	addr := os.Getenv("MILVUS_ADDR")
	if addr == "" {
		addr = DefaultMilvusURI
	}

	apikey := os.Getenv("MILVUS_APIKEY")

	// 创建Milvus客户端
	ctx := context.Background()
	clientConfig := client.Config{
		Address: addr,
	}

	// 如果有API key，则使用TLS认证
	if apikey != "" {
		clientConfig.APIKey = apikey
		clientConfig.EnableTLSAuth = true
	}

	cli, err := client.NewClient(ctx, clientConfig)
	if err != nil {
		t.Fatalf("创建Milvus客户端失败: %v", err)
	}
	defer cli.Close()

	// 创建MilvusVectorStore
	store, err := NewMilvusVectorStore(ctx, &Config{
		MilvusClient: cli,
		Embedding:    &mockEmbedding{},
	})
	if err != nil {
		t.Fatalf("创建MilvusVectorStore失败: %v", err)
	}
	defer store.Close()

	// 测试添加SQL问题
	sqlID, err := store.AddQuestionSQL(ctx, "如何查询所有用户？", "SELECT * FROM users")
	if err != nil {
		t.Fatalf("添加SQL问题失败: %v", err)
	}
	t.Logf("添加SQL问题成功，ID: %s", sqlID)

	// 测试添加DDL
	ddlID, err := store.AddDDL(ctx, "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255))")
	if err != nil {
		t.Fatalf("添加DDL失败: %v", err)
	}
	t.Logf("添加DDL成功，ID: %s", ddlID)

	// 测试添加文档
	docID, err := store.AddDocumentation(ctx, "用户表包含用户的基本信息，包括ID和姓名")
	if err != nil {
		t.Fatalf("添加文档失败: %v", err)
	}
	t.Logf("添加文档成功，ID: %s", docID)

	// 测试搜索相似问题
	similarQuestions, err := store.GetSimilarQuestionSQL(ctx, "怎么获取用户列表？")
	if err != nil {
		t.Fatalf("搜索相似问题失败: %v", err)
	}
	t.Logf("找到 %d 个相似问题", len(similarQuestions))
	for i, q := range similarQuestions {
		t.Logf("相似问题 %d: %s -> %s", i+1, q.Question, q.SQL)
	}

	// 测试获取相关DDL
	relatedDDL, err := store.GetRelatedDDL(ctx, "用户表结构")
	if err != nil {
		t.Fatalf("获取相关DDL失败: %v", err)
	}
	t.Logf("找到 %d 个相关DDL", len(relatedDDL))
	for i, ddl := range relatedDDL {
		t.Logf("相关DDL %d: %s", i+1, ddl)
	}

	// 测试获取相关文档
	relatedDocs, err := store.GetRelatedDocumentation(ctx, "用户信息")
	if err != nil {
		t.Fatalf("获取相关文档失败: %v", err)
	}
	t.Logf("找到 %d 个相关文档", len(relatedDocs))
	for i, doc := range relatedDocs {
		t.Logf("相关文档 %d: %s", i+1, doc)
	}

	// 测试获取训练数据
	trainingData, err := store.GetTrainingData(ctx)
	if err != nil {
		t.Fatalf("获取训练数据失败: %v", err)
	}
	t.Logf("共有 %d 条训练数据", len(trainingData))

	// 测试删除训练数据
	success, err := store.RemoveTrainingData(ctx, sqlID)
	if err != nil {
		t.Fatalf("删除训练数据失败: %v", err)
	}
	if success {
		t.Logf("成功删除训练数据: %s", sqlID)
	} else {
		t.Errorf("删除训练数据失败")
	}
}

// 定义模拟的嵌入模型，用于测试
type mockEmbedding struct{}

// EmbedStrings 实现embedding.Embedder接口
func (m *mockEmbedding) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	// 从文件加载嵌入向量，如果文件不存在则生成随机向量
	bytes, err := os.ReadFile("./embeddings.json")
	if err == nil {
		var v vector
		if err := sonic.Unmarshal(bytes, &v); err == nil && len(v.Data) > 0 {
			res := make([][]float64, 0, len(texts))
			for range texts {
				// 复用第一个向量，实际应用中应该根据文本生成不同的向量
				res = append(res, v.Data[0].Embedding)
			}
			return res, nil
		}
	}

	// 如果无法加载向量，则生成随机向量
	log.Printf("无法加载嵌入向量文件，使用随机向量: %v", err)
	res := make([][]float64, len(texts))
	for i := range texts {
		res[i] = make([]float64, 128)
		for j := range res[i] {
			res[i][j] = 0.1 * float64((i*j)%10)
		}
	}
	return res, nil
}

type vector struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
