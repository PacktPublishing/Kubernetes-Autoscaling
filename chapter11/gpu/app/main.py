import os
import time
import logging
import threading
import io

from flask import Flask, request, jsonify
from prometheus_client import Counter, Histogram, Gauge, generate_latest, start_http_server
import torch
from torchvision.models import resnet18, ResNet18_Weights
import torchvision.transforms as transforms
from PIL import Image

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Prometheus metrics
request_count = Counter(
    'gpu_inference_requests_total', 
    'Total number of inference requests', 
    ['endpoint', 'method', 'status'])

request_latency = Histogram(
    'gpu_inference_request_duration_seconds',
    'Request latency in seconds',
    ['endpoint'])

gpu_model_load_time = Histogram(
    'gpu_model_load_duration_seconds',
    'Time taken to load model on GPU')

active_requests = Gauge(
    'gpu_inference_active_requests',
    'Number of currently active inference requests')

model_loaded = Gauge(
    'gpu_model_loaded',
    'Whether the model is loaded (1) or not (0)')

# GPU and model globals
device = None
model = None
transform = None

def setup_gpu():
    """Initialize and load model"""
    global device, model, transform

    # Check if CUDA is available
    if torch.cuda.is_available():
        device = torch.device("cuda")
        logger.info(f"Using GPU: {torch.cuda.get_device_name(0)}")
    else:
        device = torch.device("cpu")
        logger.info("CUDA not available, using CPU")

    # Load ResNet18 model using modern weights API (pretrained argument deprecated)
    start_time = time.time()
    weights = ResNet18_Weights.DEFAULT  # You can pick IMAGENET1K_V1 or DEFAULT
    model = resnet18(weights=weights)
    model = model.to(device)
    model.eval()
    load_time = time.time() - start_time
    gpu_model_load_time.observe(load_time)
    model_loaded.set(1)

    # Define image transforms (from weights' preprocessing)
    transform = weights.transforms()
    logger.info(f"Model loaded on {device} in {load_time:.2f} seconds")

def simulate_gpu_inference(image_tensor):
    """Simulate GPU-intensive inference"""
    with torch.no_grad():
        # Artificial compute load to simulate GPU usage
        for _ in range(10):
            _ = torch.matmul(image_tensor, image_tensor.transpose(-2, -1))
        # Actual model inference
        output = model(image_tensor)
        probabilities = torch.nn.functional.softmax(output[0], dim=0)
        top_prob, top_class = torch.topk(probabilities, 1)
    return {
        "predicted_class": int(top_class[0]),
        "confidence": float(top_prob[0]),
        "device_used": str(image_tensor.device)
    }

@app.route('/health')
def health():
    """Health check endpoint"""
    request_count.labels(endpoint='/health', method=request.method, status='200').inc()
    return jsonify({"status": "healthy", "model_loaded": bool(model)})

@app.route('/predict', methods=['POST'])
def predict():
    """Main prediction endpoint"""
    start_time = time.time()
    active_requests.inc()
    try:
        if not model:
            logger.error("Model not loaded")
            request_count.labels(endpoint='/predict', method=request.method, status='500').inc()
            return jsonify({"error": "Model not loaded"}), 500

        if 'image' not in request.files:
            logger.error("No 'image' field in request.files")
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": "No image provided"}), 400

        file = request.files['image']
        if file.filename == '':
            logger.error("Empty filename for uploaded image")
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": "No image selected"}), 400

        file_content = file.read()
        if not file_content:
            logger.error("Uploaded image file is empty")
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": "Empty file content"}), 400

        logger.info(f"Received image file size: {len(file_content)} bytes")

        # Use BytesIO to create an in-memory stream for safe PIL decoding
        image_stream = BytesIO(file_content)
        try:
            image = Image.open(image_stream).convert('RGB')
        except Exception as e:
            logger.error(f"Invalid image file: {e}")
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": "Invalid image file"}), 400

        image_tensor = transform(image).unsqueeze(0).to(device)
        result = simulate_gpu_inference(image_tensor)

        request_count.labels(endpoint='/predict', method=request.method, status='200').inc()
        return jsonify(result), 200

    except Exception as e:
        logger.exception(f"Prediction error: {str(e)}")
        request_count.labels(endpoint='/predict', method=request.method, status='500').inc()
        return jsonify({"error": "Internal server error"}), 500

    finally:
        active_requests.dec()
        request_latency.labels(endpoint='/predict').observe(time.time() - start_time)

@app.route('/metrics')
def metrics():
    """Prometheus metrics endpoint"""
    return generate_latest()

if __name__ == '__main__':
    setup_gpu()

    def start_metrics_server():
        start_http_server(8080)
    metrics_thread = threading.Thread(target=start_metrics_server, daemon=True)
    metrics_thread.start()

    logger.info("Starting Flask app on port 5000")
    logger.info("Prometheus metrics available on port 8080/metrics")
    app.run(host='0.0.0.0', port=5000, threaded=True)