#!/usr/bin/env python3
"""
向量嵌入服务
提供文本向量化 API
"""
from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer
import numpy as np
import logging
import os

app = Flask(__name__)
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 配置
MODEL_NAME = os.getenv('MODEL_NAME', 'all-MiniLM-L6-v2')
HOST = os.getenv('HOST', '0.0.0.0')
PORT = int(os.getenv('PORT', 8765))

# 全局模型实例
model = None

def load_model():
    """加载模型（只加载一次）"""
    global model
    if model is None:
        logger.info(f"Loading model: {MODEL_NAME}")
        model = SentenceTransformer(MODEL_NAME)
        logger.info(f"Model loaded! Dimension: {model.get_sentence_embedding_dimension()}")
    return model

@app.route('/health', methods=['GET'])
def health():
    """健康检查"""
    try:
        m = load_model()
        return jsonify({
            "status": "healthy",
            "model": MODEL_NAME,
            "dimension": m.get_sentence_embedding_dimension()
        })
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return jsonify({"status": "unhealthy", "error": str(e)}), 500

@app.route('/embed', methods=['POST'])
def embed():
    """单个文本向量化"""
    try:
        data = request.get_json()
        if not data:
            return jsonify({"error": "Request body must be JSON"}), 400
        
        text = data.get('text', '')
        if not text:
            return jsonify({"error": "Field 'text' is required"}), 400
        
        m = load_model()
        embedding = m.encode(text, convert_to_numpy=True)
        
        return jsonify({
            "embedding": embedding.tolist(),
            "dimension": len(embedding),
            "model": MODEL_NAME
        })
    
    except Exception as e:
        logger.error(f"Embedding error: {e}")
        return jsonify({"error": str(e)}), 500

@app.route('/embed/batch', methods=['POST'])
def embed_batch():
    """批量文本向量化"""
    try:
        data = request.get_json()
        if not data:
            return jsonify({"error": "Request body must be JSON"}), 400
        
        texts = data.get('texts', [])
        if not texts or not isinstance(texts, list):
            return jsonify({"error": "Field 'texts' must be a non-empty array"}), 400
        
        batch_size = data.get('batch_size', 32)
        
        m = load_model()
        embeddings = m.encode(texts, convert_to_numpy=True, batch_size=batch_size)
        
        return jsonify({
            "embeddings": [emb.tolist() for emb in embeddings],
            "count": len(embeddings),
            "dimension": len(embeddings[0]) if len(embeddings) > 0 else 0,
            "model": MODEL_NAME
        })
    
    except Exception as e:
        logger.error(f"Batch embedding error: {e}")
        return jsonify({"error": str(e)}), 500

@app.route('/similarity', methods=['POST'])
def similarity():
    """计算两个文本的相似度"""
    try:
        data = request.get_json()
        if not data:
            return jsonify({"error": "Request body must be JSON"}), 400
        
        text1 = data.get('text1', '')
        text2 = data.get('text2', '')
        
        if not text1 or not text2:
            return jsonify({"error": "Fields 'text1' and 'text2' are required"}), 400
        
        m = load_model()
        embeddings = m.encode([text1, text2], convert_to_numpy=True)
        
        # 计算余弦相似度
        similarity_score = np.dot(embeddings[0], embeddings[1]) / (
            np.linalg.norm(embeddings[0]) * np.linalg.norm(embeddings[1])
        )
        
        return jsonify({
            "similarity": float(similarity_score),
            "text1": text1,
            "text2": text2,
            "model": MODEL_NAME
        })
    
    except Exception as e:
        logger.error(f"Similarity error: {e}")
        return jsonify({"error": str(e)}), 500

@app.errorhandler(404)
def not_found(error):
    return jsonify({"error": "Endpoint not found"}), 404

@app.errorhandler(500)
def internal_error(error):
    return jsonify({"error": "Internal server error"}), 500

if __name__ == '__main__':
    # 预加载模型
    try:
        load_model()
        logger.info(f"Starting server on {HOST}:{PORT}")
        app.run(host=HOST, port=PORT, debug=False, threaded=True)
    except Exception as e:
        logger.error(f"Failed to start server: {e}")
        exit(1)
