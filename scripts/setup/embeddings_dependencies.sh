#!/bin/bash
set -e

echo "🚀 Iniciando instalação das dependências para conversão de modelos Hugging Face → GGUF → Ollama"

# Atualizar pacotes
sudo apt update && sudo apt upgrade -y

# Instalar ferramentas básicas
sudo apt install -y git make build-essential python3 python3-pip python3-venv wget curl git-lfs

# Inicializar Git LFS
git lfs install

# Clonar llama.cpp
if [ ! -d "llama.cpp" ]; then
  git clone https://github.com/ggerganov/llama.cpp
fi
cd llama.cpp
make
cd ..

# Criar ambiente virtual Python (opcional, recomendado)
python3 -m venv venv
source venv/bin/activate

# Instalar dependências Python
pip install --upgrade pip
pip install transformers sentencepiece onnx onnxruntime accelerate

# Instalar Ollama (se não estiver instalado)
if ! command -v ollama &> /dev/null
then
  curl -fsSL https://ollama.com/download.sh | sh
fi

echo "✅ Ambiente pronto! Dependências instaladas:"
echo "- git, make, build-essential, python3, pip"
echo "- git-lfs"
echo "- llama.cpp compilado"
echo "- libs Hugging Face (transformers, sentencepiece)"
echo "- ONNX e onnxruntime"
echo "- Ollama instalado"
