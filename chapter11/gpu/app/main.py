import os
import time
import logging
import threading
import requests
from io import BytesIO

from flask import Flask, request, jsonify
from prometheus_client import Counter, Histogram, Gauge, start_http_server
import torch
from torchvision.models import resnet18, ResNet18_Weights
import torchvision.transforms as transforms
from PIL import Image

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Prometheus metrics (existing ones)
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

    if torch.cuda.is_available():
        device = torch.device("cuda")
        logger.info(f"Using GPU: {torch.cuda.get_device_name(0)}")
    else:
        device = torch.device("cpu")
        logger.info("CUDA not available, using CPU")

    start_time = time.time()
    weights = ResNet18_Weights.DEFAULT
    model = resnet18(weights=weights)
    model = model.to(device)
    model.eval()
    load_time = time.time() - start_time
    gpu_model_load_time.observe(load_time)
    model_loaded.set(1)

    transform = weights.transforms()
    logger.info(f"Model loaded on {device} in {load_time:.2f} seconds")

def simulate_gpu_inference(image_tensor):
    """Simulate GPU-intensive inference"""
    with torch.no_grad():
        for _ in range(10):
            _ = torch.matmul(image_tensor, image_tensor.transpose(-2, -1))
        output = model(image_tensor)
        probabilities = torch.nn.functional.softmax(output[0], dim=0)
        top_prob, top_class = torch.topk(probabilities, 1)
    return {
        "predicted_class": int(top_class[0]),
        "confidence": float(top_prob[0]),
        "device_used": str(image_tensor.device)
    }

def download_image_from_url(image_url, timeout=10):
    """Download image from URL and return PIL Image object"""
    try:
        response = requests.get(image_url, timeout=timeout, stream=True)
        response.raise_for_status()
        
        # Check content type
        content_type = response.headers.get('content-type', '')
        if not content_type.startswith('image/'):
            raise ValueError(f"URL does not point to an image. Content-Type: {content_type}")
        
        # Load image from response content
        image_stream = BytesIO(response.content)
        image = Image.open(image_stream).convert('RGB')
        
        logger.info(f"Successfully downloaded image from URL: {image_url}, size: {len(response.content)} bytes")
        return image
        
    except requests.exceptions.RequestException as e:
        logger.error(f"Failed to download image from URL {image_url}: {e}")
        raise ValueError(f"Failed to download image: {str(e)}")
    except Exception as e:
        logger.error(f"Failed to process image from URL {image_url}: {e}")
        raise ValueError(f"Invalid image data: {str(e)}")

@app.route('/health')
def health():
    """Health check endpoint"""
    request_count.labels(endpoint='/health', method=request.method, status='200').inc()
    return jsonify({"status": "healthy", "model_loaded": bool(model)})

@app.route('/predict', methods=['POST'])
def predict():
    """Main prediction endpoint that accepts image URL"""
    start_time = time.time()
    active_requests.inc()
    
    try:
        if not model:
            logger.error("Model not loaded")
            request_count.labels(endpoint='/predict', method=request.method, status='500').inc()
            return jsonify({"error": "Model not loaded"}), 500

        # Get image URL from JSON payload
        data = request.get_json()
        if not data or 'image_url' not in data:
            logger.error("No 'image_url' field in request JSON")
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": "Missing 'image_url' field in JSON payload"}), 400

        image_url = data['image_url']
        if not image_url or not image_url.strip():
            logger.error("Empty image_url provided")
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": "Empty image_url provided"}), 400

        # Download and process image
        try:
            image = download_image_from_url(image_url)
        except ValueError as e:
            request_count.labels(endpoint='/predict', method=request.method, status='400').inc()
            return jsonify({"error": str(e)}), 400

        image_tensor = transform(image).unsqueeze(0).to(device)
        result = simulate_gpu_inference(image_tensor)
        
        # Add source URL to result
        result['source_url'] = image_url
        
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
