---
layout: home
hero:
  name: Packster
  text: Package Version Management
  tagline: Store, retrieve, and manage versioned build artifacts with a simple REST API.
  image:
    src: /packster-logo.png
    alt: Packster
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started
    - theme: alt
      text: API Reference
      link: /api
    - theme: alt
      text: View on GitHub
      link: https://github.com/IdanKoblik/packster
features:
  - title: Token-based Auth
    details: Every request is authenticated via an X-Api-Token header. Admin tokens can manage users; per-product permissions control who can upload, download, or delete.
  - title: Product Lifecycle
    details: Create products with optional group names (e.g. staging/production), grant token access, and manage multiple named versions — all through a clean REST API.
  - title: Artifact Storage
    details: Upload arbitrary files as versioned artifacts. SHA-256 checksums are computed and stored automatically. Metadata is persisted in MySQL with token lookups cached in Redis.
---
