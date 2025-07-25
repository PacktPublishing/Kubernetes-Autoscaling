FROM nvidia/cuda:11.8.0-runtime-ubuntu20.04

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Create working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt

# Copy application code
COPY main.py .

# Create a non-root user
RUN useradd --create-home --shell /bin/bash app
USER app

# Expose ports
EXPOSE 5000 8080

# Set default command
CMD ["python3", "main.py"]
