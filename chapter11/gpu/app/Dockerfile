FROM nvidia/cuda:12.9.1-runtime-ubuntu24.04

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && \
    apt-get install -y python3-pip python3-venv

# Create working directory
WORKDIR /app

# Create a virtual environment
RUN python3 -m venv /opt/venv

# Activate it, and install dependencies inside
ENV PATH="/opt/venv/bin:$PATH"

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY main.py .

# Create a non-root user
RUN useradd --create-home --shell /bin/bash app
USER app

# Expose ports
EXPOSE 5000 8080

# Set default command
CMD ["python3", "main.py"]
