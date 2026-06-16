# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copia o módulo e baixa dependências (camada cacheável)
COPY go.mod ./
RUN go mod download

# Copia todo o código fonte
COPY . .

# TARGET define qual binário compilar: broker | cliente | drone | sensor | testes
ARG TARGET=broker
RUN go build -o /bin/app ./cmd/${TARGET}

# ============================================================
# Stage 2: Runtime mínimo
# ============================================================
FROM alpine:3.19

WORKDIR /root/

COPY --from=builder /bin/app /bin/app

ENTRYPOINT ["/bin/app"]