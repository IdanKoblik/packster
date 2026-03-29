# Getting Started

## What is Packster?

Packster is a self-hosted REST API service for storing, retrieving, and managing versioned build artifacts. It lets you upload files tied to a product and version, download them later, and control access through fine-grained token permissions.

Artifacts are stored on disk, with metadata (products, versions, tokens) persisted in MongoDB and token lookups cached in Redis.

## Prerequisites

- Go 1.25+
- MongoDB 7.0+
- Redis 7+
