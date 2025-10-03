# 🚀 Migration from Local to AWS: Project Deployment Guide

This document explains how the application was migrated from a local Docker Compose setup to a production-like architecture using AWS services. It includes service layout, networking, storage, containerization, secrets management, and system resilience strategies.

---

## 📦 What We Changed

### 🧱 Split Services Across AWS Components

| Component     | AWS Service / Role                                |
|---------------|---------------------------------------------------|
| Backend       | EC2 (Private) – Hosts Go API container + Redis    |
| Worker        | EC2 (Private) – Dedicated to processing jobs      |
| NFS Storage   | EC2 (Private) – `nfs-kernel-server` exports `/srv/anbdata` |
| Frontend      | EC2 (Public) – NGINX serves SPA and proxies requests + SonarQube + Asynqmon |
| Database      | Amazon RDS PostgreSQL (Managed)                   |

---

## 🗃️ Storage & Mounts

- **NFS server exports**: `/srv/anbdata`
- **Backend & Worker mount**: NFS at `/data`
- **SystemD config**: Automount + correct start-order to survive stop/start events

---

## 🐳 Containers per Node

### 🔹 Backend (Private EC2)
- **API Container**: Mounts `/data`, connects to RDS and Redis at `BACKEND_IP:6379`
- **Redis Container**: Based on `redis:7`, AOF enabled, exposed only within security group

### 🔹 Worker (Private EC2)
- Mounts:
  - `/data` from NFS
  - `/home/ubuntu/ISIS4426-Entrega1/assets:/assets:ro`
- Connects to Redis & RDS

### 🔹 Frontend (Public EC2)
- **NGINX**:
  - Serves SPA
  - Proxies `/api` and `/static` to Backend EC2:8080
  - Asynqmon (port 8081)
  - SonarQube (port 9000)

---

## 🔐 Secrets & Config

- **AWS SSM Parameter Store** used for:
  - `DB_PASSWORD`
  - `JWT_SECRET`
  - `SONAR_DB_PASSWORD`
- **IAM Roles** attached to EC2 instances for secure access

---

## 🔐 Security Groups

| Source       | Destination         | Port    | Purpose                      |
|--------------|---------------------|---------|------------------------------|
| Frontend     | Backend             | 8080    | Proxy API and static files   |
| Frontend     | Backend             | 6379    | Asynqmon access              |
| Worker       | Backend             | 6379    | Access Redis                 |
| Worker       | Backend             | 8080    | Proxy API and static files   |
| Backend/Worker | NFS Server        | 2049    | NFS mounts                   |
| Backend/Worker | RDS               | 5432    | Database access              |
| Public       | Frontend            | 80/443  | SPA access                   |
| My IP      | Frontend            | 8081    | Asynqmon                     |
| My IP      | Frontend            | 9000    | SonarQube                    |

---

## 🧪 Technologies Used

- **AWS**: EC2, RDS (Postgres), SSM Parameter Store, IAM Roles
- **Linux/NFS**: `nfs-kernel-server`, `nfs-common`, `systemd` automounts
- **Docker**: API, Worker, Redis, NGINX, Asynqmon, SonarQube containers
- **Backend**: Go API + [Asynq](https://github.com/hibiken/asynq)
- **Frontend**: SPA + NGINX reverse proxy
- **Observability**: Asynqmon, SonarQube
- **Tooling**: AWS CLI, `psql`, `redis-cli`

---

## 🔄 End-to-End Flow

### 🌐 Browser → Frontend EC2 (NGINX)
- `GET /` → Serves SPA
- `POST /api/*` → Proxied to Backend:8080
- `GET /static/*` → Proxied to Backend:8080 (files from `/data`)

### ⚙️ Backend EC2
- Auth / Queries → RDS (5432, SSL required)
- File uploads → `/data/uploads` (NFS)
- Jobs → Enqueued in Redis (`BACKEND_IP:6379`)
- Processed media → Served from `/static` (maps to `/data/processed`)

### 🧠 Redis (Backend EC2)
- Receives jobs from API
- Serves jobs to Worker
- Visualized in Asynqmon

### 🛠️ Worker EC2
- Listens to Redis
- Reads: `/data/uploads`
- Writes: `/data/processed`
- Updates RDS with status

---

## 🛠 Observability and Monitoring Tools

| Tool       | Description                                      |
|------------|--------------------------------------------------|
| **Asynqmon** | Web UI (port 8081) for Redis job queues        |
| **SonarQube** | Code analysis dashboard (port 9000)           |

Both run on Frontend EC2 and are secured using security groups.
