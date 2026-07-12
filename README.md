# 🚀 API Client Request

API REST desenvolvida em **Go** para gerenciamento de clientes, utilizando autenticação segura baseada em **JWT Access Token** e **Refresh Token com rotação**, seguindo boas práticas de segurança para APIs.

---

## 📋 Funcionalidades

- Cadastro de usuários
- Login com autenticação JWT
- Refresh Token com rotação
- Logout com revogação de sessão
- CRUD completo de clientes
- Proteção de rotas com middleware JWT
- Senhas armazenadas com BCrypt
- Migrations para gerenciamento do banco de dados
- PostgreSQL executando via Docker

---

## 🛠 Tecnologias

- Go 1.26+
- PostgreSQL
- Docker
- Docker Compose
- golang-migrate
- JWT
- BCrypt

---

## 📁 Estrutura do Projeto

```
.
├── internal/
│   ├── auth/              # Autenticação e JWT
│   ├── clientes/          # Regras de negócio dos clientes
│   ├── config/            # Configuração da aplicação
│   ├── database/          # Conexão com PostgreSQL
│   └── httpapi/           # Rotas, handlers e middlewares
│
├── migrations/            # Versionamento do banco
├── docker-compose.yml
├── main.go
└── README.md
```

---

## 🔐 Fluxo de Autenticação

```text
Usuário
    │
    ▼
POST /auth/login
    │
    ▼
JWT Access Token (15 min)
+
Refresh Token
    │
    ▼
Requisições autenticadas
    │
    ▼
Quando expira
    │
    ▼
POST /auth/refresh
    │
    ▼
Novo Access Token
+
Novo Refresh Token
```

### Segurança

- Senhas protegidas com BCrypt
- Access Token assinado com JWT
- Refresh Token armazenado apenas como SHA-256
- Rotação automática de Refresh Tokens
- Revogação da família de tokens em caso de reutilização
- PostgreSQL acessível apenas localmente

---

# ⚙️ Pré-requisitos

- Go 1.26+
- Docker Desktop
- Docker Compose
- golang-migrate

---

## 📦 Instalação

Clone o projeto:

```bash
git clone https://github.com/seu-usuario/api-client-request.git

cd api-client-request
```

---

## 🔧 Configuração

Copie o arquivo de exemplo:

```bash
cp .env.example .env
```

Configure as variáveis conforme necessário.

---

## 🐳 Subindo o PostgreSQL

```bash
docker compose up -d postgres
```

Verifique se o container está em execução:

```bash
docker ps
```

---

## 🗄 Executando as Migrations

Instale o migrate:

```bash
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Execute:

```bash
migrate \
-path migrations \
-database "$DATABASE_URL" \
up
```

---

## ▶️ Executando a API

```bash
go run .
```

A aplicação ficará disponível em:

```
http://localhost:8080
```

---

# 📬 Endpoints

## Autenticação

| Método | Endpoint | Descrição |
|---------|----------|-----------|
| POST | `/auth/register` | Cadastro |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Renovar Token |
| POST | `/auth/logout` | Logout |

---

## Clientes

| Método | Endpoint |
|---------|----------|
| POST | `/cliente` |
| GET | `/cliente` |
| GET | `/cliente/{id}` |
| PUT | `/cliente/{id}` |
| DELETE | `/cliente/{id}` |

---

# 🧪 Testando

Após iniciar a API:

1. Registrar usuário

```
POST /auth/register
```

↓

2. Fazer Login

```
POST /auth/login
```

↓

3. Copiar o Access Token

↓

4. Adicionar no Header

```
Authorization: Bearer <token>
```

↓

5. Consumir os endpoints de Cliente

---

# 🔒 Banco de Dados

O PostgreSQL é executado em um container Docker e exposto apenas para:

```
127.0.0.1
```

Dessa forma o banco não fica acessível pela rede externa.

Toda comunicação deve ocorrer através da API.

---

# 📈 Próximas melhorias

- [ ] Testes unitários
- [ ] Testes de integração
- [ ] Swagger/OpenAPI
- [ ] Logs estruturados
- [ ] Rate Limiting
- [ ] Observabilidade
- [ ] CI/CD
- [ ] Deploy em Cloud

---

# 👨‍💻 Autor

Desenvolvido por **Deivid Ferreira**.