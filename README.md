# 🌀 RuneEngine v1.0

Conversor de imagens e GIFs para ASCII art (ANSI 256 colors) utilizando microserviços, Go e SvelteKit.

## 📁 Estrutura do Projeto

```text
rune-engine/
├── backend/           # API em Go (Processamento e Engine)
│   ├── main.go        # Entrypoint, Router e Middlewares
│   ├── internal/      # Core: ASCII converter, Redis cache e Worker Pool
│   └── Dockerfile
├── frontend/          # SPA em SvelteKit
│   ├── src/           # Componentes (XTerm.js) e lógica de Streaming
│   ├── svelte.config.js # Configurado com adapter-node
│   └── Dockerfile
├── nginx/
│   └── default.conf   # Configuração do Proxy Reverso
├── docker-compose.yml
└── Makefile           # Atalhos para comandos Docker

```

## 🛠️ Tecnologias e Recursos

* **Backend**: Go (Gin Gonic), Redis (Cache/Rate Limit).
* **Frontend**: SvelteKit, XTerm.js (Renderização Terminal).
* **Infra**: Nginx (Proxy Reverso), Docker, Docker Compose.

## 🚀 Como Rodar

1. **Requisitos**: Docker e Docker Compose.
2. **Configuração**: Crie um arquivo `.env` em `backend/` com a variável `REDIS_PASSWORD`.
3. **Execução**:
```bash
make up

```

4. **Acesso**: [http://localhost](https://www.google.com/search?q=http://localhost)

## ⌨️ Atalhos do Makefile

* `make up`: Builda e sobe todos os containers em background.
* `make down`: Para e remove os containers e redes.
* `make logs`: Exibe logs em tempo real de todos os serviços.
* `make stats`: Monitora consumo de CPU/RAM dos containers.
* `make redis-cli`: Abre o terminal interativo do Redis.

