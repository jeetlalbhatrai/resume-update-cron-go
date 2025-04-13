# -------- Stage 1: Build the Go binary --------
    FROM golang:1.24 AS builder

    # Set working directory inside container
    WORKDIR /app
    
    # Copy dependency files and download modules
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the entire project
    COPY . .
    
    # Build the binary
    RUN go build -o resume-updater main.go
    
    
    # -------- Stage 2: Create lightweight runtime image --------
    FROM debian:bullseye-slim
    
    # Install system dependencies and Chrome
    RUN apt-get update && apt-get install -y \
        curl \
        wget \
        unzip \
        gnupg \
        ca-certificates \
        fonts-liberation \
        libappindicator3-1 \
        libasound2 \
        libatk-bridge2.0-0 \
        libatk1.0-0 \
        libcups2 \
        libdbus-1-3 \
        libgdk-pixbuf2.0-0 \
        libnspr4 \
        libnss3 \
        libxcomposite1 \
        libxdamage1 \
        libxrandr2 \
        xdg-utils \
        --no-install-recommends && rm -rf /var/lib/apt/lists/*
    
    # Add Chrome repo and install Chrome
    RUN curl -fsSL https://dl.google.com/linux/linux_signing_key.pub | gpg --dearmor -o /usr/share/keyrings/google-linux-signing-key.gpg && \
        echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-linux-signing-key.gpg] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list && \
        apt-get update && apt-get install -y google-chrome-stable && rm -rf /var/lib/apt/lists/*
    
    # Install ChromeDriver
    RUN CHROME_DRIVER_VERSION=$(curl -sS https://chromedriver.storage.googleapis.com/LATEST_RELEASE) && \
        wget -q -O /tmp/chromedriver.zip https://chromedriver.storage.googleapis.com/$CHROME_DRIVER_VERSION/chromedriver_linux64.zip && \
        unzip /tmp/chromedriver.zip -d /usr/local/bin && \
        chmod +x /usr/local/bin/chromedriver && \
        rm /tmp/chromedriver.zip
    
    # Set workdir in final container
    WORKDIR /app
    
    # Copy built binary from builder
    COPY --from=builder /app/resume-updater .
    
    # Expose chromedriver port (optional)
    EXPOSE 9515
    
    # Use env file during run, not in Dockerfile for secrets
    # Run the Go application
    CMD ["./resume-updater"]
    