import subprocess

def main():
    # Dependências necessárias:
    # pip install transformers sentencepiece onnx onnxruntime
    # Ollama instalado

    MODEL = "neuralmind/bert-base-portuguese-cased"
    OUTPUT_DIR = "onnx_model"

    # 1. Exportar modelo Hugging Face para ONNX
    subprocess.run([
        "python3", "-m", "transformers.onnx",
        "--model", MODEL,
        OUTPUT_DIR
    ], check=True)

    # 2. Converter para GGUF
    subprocess.run([
        "python3", "convert.py",
        "--onnx", f"{OUTPUT_DIR}/model.onnx",
        "--out", "bertimbau.gguf"
    ], check=True)

    # 3. Quantizar
    subprocess.run([
        "./quantize", "bertimbau.gguf",
        "bertimbau.Q4_K_M.gguf", "Q4_K_M"
    ], check=True)

    # 4. Criar Modelfile
    modelfile_content = """FROM ./bertimbau.Q4_K_M.gguf
    PARAMETER temperature 0.0
    PARAMETER top_p 0.9
    PARAMETER top_k 40
    TEMPLATE \"\"\"{{ .Prompt }}\"\"\""""

    with open("Modelfile", "w") as f:
        f.write(modelfile_content)

    # 5. Registrar no Ollama
    subprocess.run([
        "ollama", "create", "bertimbau", "-f", "Modelfile"
    ], check=True)

    print("Modelo pronto! Execute: ollama run bertimbau")

if __name__ == "__main__":
    main()
