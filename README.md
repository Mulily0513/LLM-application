# LLM-application
rag multi-agent
Invoke-WebRequest -Uri https://raw.githubusercontent.com/anthropics/anthropic-cookbook/refs/heads/main/skills/contextual-embeddings/data/evaluation_set.jsonl -OutFile evaluation_set.jsonl
Invoke-WebRequest -Uri https://raw.githubusercontent.com/anthropics/anthropic-cookbook/refs/heads/main/skills/contextual-embeddings/data/codebase_chunks.json -OutFile codebase_chunks.json
wget http://qim.fs.quoracdn.net/quora_duplicate_questions.tsv