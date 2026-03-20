# VTC Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Cloud-native VTC (French for "Private Hire") management platform built with Go microservices. This project implements a scalable, event-driven architecture designed to handle the full lifecycle of private transport—from booking to automated legal compliance.

## 🧪 Project Purpose
This is a **distributed systems laboratory**. While a monolith would be simpler, this project is intentionally built as a microservices ecosystem to study:
* **Go & Hexagonal Architecture** (Clean, testable code)
* **Event-Driven Design** (NATS JetStream)
* **Cloud-Native Infrastructure** (Kubernetes/k3s, Traefik, Docker)

The goal is to master the complexity of scaling and orchestrating a production-grade backend.

## Architecture Overview

The platform is designed using **Hexagonal Architecture** (Ports and Adapters) to ensure business logic remains decoupled from external dependencies like databases and message brokers.

- **Monorepo Approach:** Single source of truth for all services, enabling atomic changes to shared contracts.
- **Event-Driven:** Asynchronous communication between services powered by **NATS JetStream**.
- **Orchestration:** Containerized via Docker and orchestrated with **k3s (Kubernetes)**.
- **Authentication:** Identity management provided by **Clerk**.

## Tech Stack

* **Language:** Go (Golang)
* **Database:** PostgreSQL (One instance per service for data isolation)
* **Messaging:** NATS JetStream
* **Auth:** Clerk (OIDC / JWT)
* **Infrastructure:** K3s, Traefik (Ingress), Docker

## Repository Structure

```text
├── cmd/                # Entry points for each service binary
├── internal/           # Private business logic (Hexagonal: Handler -> Service -> Repo)
│   ├── shared/         # Shared internal utilities (Auth, DB, Redis)
│   └── [service]/      # Service-specific logic
├── pkg/                # Publicly importable packages (Events, Shared Types)
├── deployments/        # Kubernetes (k3s) manifests and Helm charts
└── scripts/            # Database migrations and automation tools