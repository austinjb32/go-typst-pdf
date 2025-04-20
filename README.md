# Go-Typst PDF Generator

A Go-based service that compiles Typst templates into PDFs, supporting image embedding and AWS S3 storage.

> 🚧 **Under Development**  
>This project is currently in active development. Features may be unstable, and the API is subject to change

---

## Goal
To create a PDF generation server which can generate 10000s PDF under a minute with 2gb of RAM.

## 🛠 Features

 Compile Typst templates into PDFs
 Embed images from local paths or remote URLs
 Upload generated PDFs to AWS S3 with pre-signed URLs
 Dockerized for consistent deployment
 Supports both HTTP and gRPC interfaces

---

## 🚀 Getting Started

### Prerequisites
- Go
- Typst
- Redis
- Docker Compose
- AWS S3

### Build and Run

```bash
# Build the Docker image
docker build -t go-typst .

# Run the container
docker run -p 8080:8080 -p 50051:50051 go-typst
```

---

## ⚙️ Configuratio

The application uses a `.env` file for configuration. Ensure this file is present in the root directory with the following variables:

```env
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=your_region
AWS_BUCKET=your_bucket_name
```

> ❗ **Note:* Ensure the `.env` file is correctly formatted and placd.

---

## 🖼 Image Embedding in Typt (In Development)

To embed images in your Typst templates, use the following syntax:

```typst
#image("https://example.com/image.jpg", width: 80%)
``


Ensure that remote images are accessible and local images are correctly referenced within the Docker container.

---

## 🐳 Dockerfile Highlight

- Multi-stage build with Go
- Installs Typst CLI
- Exposes ports 8080 (HTTP) and 50051 (RPC)

---

## 🧪 Troubleshooting

- **Typst Compilation Erros:** Ensure all image paths in your Typst templates are correct and accesible.
- **Docker Build Issus:** If you encounter issues with `xz-utils`, ensure it's installed before extracting ypst:

  ```dockerfile
  RUN apt-get update && apt-get install -y xz-utils
  ```
- **AWS S3 Upload Failures:** Check your AWS credentials and bucket permissions. Ensure the bucket exists and is correctly configured for public access if needed.
---

## 📂 Project Structure
```

.
├── Dockerfile
├── main.go
├── go.mod
├── static/
├── pdf/
│   └── templates/
├── .env
└── REDME.md
```

---

## 📄 Lcense

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
