# Text Subject Analyses

Here I'll try to make some notes on how to generate subject and sentiment analysis learning.

### Recommended models

| Model              | Size   | Key | Why for Pi 5?                                                                                                    |
|--------------------|--------|---------------------------|------------------------------------------------------------------------------------------------------------------|
| phi3:mini (3.8B)   | ~2.3GB | Best speed. Extremely fast on CPUs while maintaining high reasoning quality for its size.                        |
| llama3:8b (4-bit)  | ~4.7GB | llama3:8b-instruct-q4_K_S | Best all-rounder. Excellent at following complex instructions for subject categorization and keyword extraction. |
| mistral-nemo (12B) | ~7.5GB | Deep analysis. If you need more nuance and can tolerate a slower response (approx. 1-2 tokens/sec).              |
| nomic-embed-text   | ~274MB | Embeddings. Already used in your project, perfect for RAG and similarity-based grouping.                         |

#### Some notes on model options

- models:
  - phi-3-mini
  - llama3:8b
  - phi3:mini
  - gemma:7b (4-bit quant)
  - qwen2 (it may push to hardware limits)
  - mistral-nemo (it may push to hardware limits)
