from dotenv import load_dotenv
import os
from openai import OpenAI
load_dotenv("../../.env")
from vanna.ollama import Ollama
from vanna.milvus import Milvus_VectorStore
from vanna.flask import VannaFlaskApp
from vanna.openai import OpenAI_Chat

class MyVanna(Milvus_VectorStore, Ollama, OpenAI_Chat):
    def __init__(self, config=None):
        Milvus_VectorStore.__init__(self, config=config)
        OpenAI_Chat.__init__(self, config=config)
        #Ollama.__init__(self, config=config)

client = OpenAI(
    base_url="https://ark.cn-beijing.volces.com/api/v3",
    api_key=os.environ.get("ARK_API_KEY"),
)

from milvus_model.hybrid import BGEM3EmbeddingFunction
ef = BGEM3EmbeddingFunction(use_fp16=False, device="cpu")

vn = MyVanna(
    config={
        'embedding_function': ef,
        'client': client,
    }
)

#connect to postgres
vn.connect_to_postgres(
    host='localhost',
    port=5432,
    dbname='postgres',
    user='root',
    password='123456'
)

# train
# DDL statements are powerful because they specify table names, colume names, types, and potentially relationships
vn.train(ddl="""
    CREATE TABLE IF NOT EXISTS my-table (
        id INT PRIMARY KEY,
        name VARCHAR(100),
        age INT
    )
""")
# Sometimes you may want to add documentation about your business terminology or definitions.
vn.train(documentation="Our business defines OTIF score as the percentage of orders that are delivered on time and in full")
# You can also add SQL queries to your training data. This is useful if you have some queries already laying around. You can just copy and paste those from your editor to begin generating new SQL.
vn.train(sql="SELECT * FROM my-table WHERE name = 'John Doe'")

if __name__ == '__main__':
    app = VannaFlaskApp(vn)
    app.run(host='0.0.0.0', port=8000)