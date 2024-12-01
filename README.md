# 🚀 ZTE SMS HTTP Server

This project provides an HTTP server to interact with a ZTE router to retrieve SMS messages 📩. It supports Docker-based development 🐳 and deployment and includes a `pprof` performance profiling server 📈.

---

## 🌟 Features

- **📬 SMS Retrieval Endpoint** (`/getSMS`): Fetches SMS messages from the ZTE router.
- **⚙️ Configurable Query Parameters**: Supports pagination and filtering via query parameters.
- **📊 Performance Profiling**: Includes a `pprof` server for performance analysis.
- **🐳 Docker Support**:
  - Development with hot-reloading using [Air](https://github.com/cosmtrek/air).
  - Lightweight, secure production builds.
- **📝 Environment Configuration**: Fully configurable using `.env` files.

---

## 🛠️ Getting Started

### 📋 Prerequisites

- **🐳 Docker** and **Docker Compose**: Ensure Docker and Docker Compose are installed on your system.
- **Go (Optional)**: If running locally without Docker, install Go 1.22 or later.

---

## ⚙️ Configuration

### 🔧 Environment Variables

The application reads configuration values from a `.env` file. Create a `.env` file in the root directory with the following content:

```dotenv
PASSWORD=supersecret                 # Router password
ENDPOINT=http://192.168.1.1          # Router base URL
PPROF_LISTEN_ADDR=127.0.0.1:6060     # Address and port for the pprof server
SERVER_LISTEN_ADDR=127.0.0.1:8080    # Address and port for the main server
```

---

## 🧑‍💻 Development

### 🐳 Run Locally with Docker Compose

1. Build and start the server using Docker Compose:
   ```bash
   docker-compose up --build
   ```

2. The server will be available at:
   - 🌐 Main server: `http://localhost:38080`
   - 🔍 `pprof` server: `http://localhost:36060`

3. Edit the source files in the `src` directory, and the server will reload automatically (using [Air](https://github.com/cosmtrek/air)).

---

## 🚀 Production Deployment

### 🏗️ Build the Production Image

1. Build the production-ready Docker image:
   ```bash
   docker build -t zte-sms-read .
   ```

2. Run the container:
   ```bash
   docker run --env-file .env -p 38080:8080 -p 36060:6060 zte-sms-read
   ```

3. The server will now be running:
   - 🌐 Main server: `http://localhost:38080`
   - 🔍 `pprof` server: `http://localhost:36060`

---

## 🔌 Endpoints

### **1. `/getSMS`**

Retrieves SMS messages from the ZTE router.

- **Query Parameters**:
  - `page`: Page number (default: `0`, range: `[0, 100]`).
  - `perPage`: Number of messages per page (default: `500`, range: `[1, 1000]`).
  - `memStore`: Memory storage option (default: `1`, range: `[0, 2]`).
  - `tag`: Filter by tag (default: `10`, range: `[0, 10]`).
    - Tag `1`: Unread messages.
    - Other tags can filter other types of messages as per router configuration.

- **Example**:
  ```bash
  curl "http://localhost:38080/getSMS?page=1&perPage=100&memStore=1&tag=1"
  ```

- **Response**:
  A JSON array of SMS messages with fields like `id`, `number`, `content`, `read`, and more.

### **2. 🔍 `pprof`**

Performance profiling interface for debugging and optimization.

- **Example URLs**:
  - 🧠 CPU Profile: `http://localhost:36060/debug/pprof/profile`
  - 🛠️ Heap Profile: `http://localhost:36060/debug/pprof/heap`
  - 📜 Goroutines: `http://localhost:36060/debug/pprof/goroutine`

---

## 📂 Directory Structure

```plaintext
.
├── .env                      # Environment configuration
├── Dockerfile                # Multi-stage Docker build file
├── docker-compose.yml        # Docker Compose configuration
├── LICENSE                   # Project license
├── src/                      # Go source code
│   ├── main.go               # Main application entry point
│   ├── handler.go            # HTTP handlers
│   ├── zte_connector.go      # Router connector logic
│   └── zte_sms.go            # SMS parsing logic
└── README.md                 # Project documentation
```

---

## 🛠️ Development Tools

### **Air** 🌀

[Air](https://github.com/cosmtrek/air) is used for hot-reloading during development.

- Install Air:
  ```bash
  go install github.com/cosmtrek/air@latest
  ```

---

## 📊 Performance Profiling with `pprof`

### 🧠 Collect a CPU Profile

Run the following command to collect a 30-second CPU profile:

```bash
curl -s http://localhost:36060/debug/pprof/profile?seconds=30 > cpu.prof
```

### 🛠️ Analyze the Profile

Use the `pprof` tool to analyze the collected profile:

```bash
go tool pprof cpu.prof
```

Inside the interactive shell, you can use commands like `top`, `list`, or `web` for analysis.

---

## 📝 Notes

1. **🔒 Security**:
   - Ensure the `pprof` server is accessible only locally or via trusted IPs in production.
   - Avoid hardcoding sensitive information like passwords; use `.env` files or secrets management tools.
   - Use HTTPS for secure communication in production.

2. **⚙️ Environment Variables**:
   - Modify the `.env` file to customize the configuration.

3. **📈 Scaling**:
   - The service is designed to be lightweight and easily containerized for scaling.

---

## 🤝 Contributing

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Commit your changes (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Open a pull request.

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgments

- [Go](https://golang.org) for its simplicity and performance.
- 🌀 [Air](https://github.com/cosmtrek/air) for hot-reloading during development.
- 🐳 [Docker](https://www.docker.com) for containerization.
- 🔍 [net/http/pprof](https://pkg.go.dev/net/http/pprof) for profiling and debugging tools.
