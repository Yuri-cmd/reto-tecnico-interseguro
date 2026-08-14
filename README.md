# Coding Challenge — Interseguro (División TI)

Dos APIs REST comunicadas por HTTP:

- **`go-api`** (Go + Fiber): recibe una matriz rectangular y calcula su **factorización QR**.
- **`node-api`** (Node.js + Express): recibe las matrices `Q` y `R`, calcula **estadísticas** (máximo, mínimo, promedio, suma total, verificación de matriz diagonal) y las devuelve.

Ambas están protegidas con **JWT** y containerizadas con Docker.

## Decisión de diseño: la contradicción del enunciado

El PDF del reto tiene una inconsistencia entre dos secciones:

| Sección | Dice que Go... | Dice que Node... |
|---|---|---|
| "Arquitectura de la solución" | recibe la matriz y **rota** la matriz | calcula estadísticas sobre la matriz **rotada** |
| "Funcionalidad requerida" | recibe la matriz y devuelve su **factorización QR** | recibe el resultado y calcula la operación adicional (estadísticas) |

**Decisión tomada: implementar la factorización QR** (no la rotación), por dos razones:

1. La sección "Funcionalidad requerida" es la que define contractualmente qué debe *devolver* cada API RESTful, y remite explícitamente a la sección "Operación adicional" para el detalle de las estadísticas — es la especificación operativa, no un resumen.
2. **Argumento más fuerte:** la sección de "Operación adicional" pide "estadísticas sobre los datos de **las matrices** devueltas" y "verificar si **alguna matriz** es diagonal" — todo en plural. Una rotación produce **una sola matriz** de salida. La factorización QR produce **dos matrices** (`Q` y `R`), lo cual encaja de forma natural y literal con el enunciado detallado. La sección de arquitectura es, con toda probabilidad, una descripción genérica/reciclada de un reto previo (rotación de matriz) que no se actualizó al cambiar el requerimiento a QR.

Como el enunciado mismo habilita esto ("en caso de dudas, se espera que el candidato tome decisiones informadas y las sustente"), esta decisión se documenta aquí y se defiende con el argumento anterior en la entrevista.

## Arquitectura y flujo

```
Cliente                    go-api (Fiber, :3000)              node-api (Express, :4000)
   │                              │                                    │
   │  POST /api/auth/login        │                                    │
   ├─────────────────────────────>│  (login también disponible         │
   │  <── { token } ───────────────┤   directo en go-api)               │
   │                              │                                    │
   │  POST /api/qr                │                                    │
   │  Authorization: Bearer <tk>  │                                    │
   ├─────────────────────────────>│                                    │
   │                              │  1. Valida JWT                     │
   │                              │  2. Calcula QR (Householder)       │
   │                              │  3. POST /api/estadisticas         │
   │                              │     { q, r } + mismo Bearer token  │
   │                              ├───────────────────────────────────>│
   │                              │                                    │  Valida JWT
   │                              │                                    │  Calcula stats
   │                              │  <── { max, min, avg, sum, diag } ──┤
   │  <── { matrizOriginal, factorizacionQR, estadisticas } ───────────┤
```

`go-api` actúa como orquestador: es el único punto de entrada para el cliente, calcula la QR, reenvía `Q`/`R` a `node-api` (reutilizando el mismo Bearer token — un único secreto compartido `JWT_SECRET` valida en ambos servicios), y devuelve al cliente la respuesta combinada.

## Cómo correrlo

```bash
docker compose up --build
```

- `go-api` queda expuesta en `http://localhost:3000`
- `node-api` queda expuesta en `http://localhost:4000` (normalmente no se llama directo; es consumida internamente por `go-api`, pero queda accesible para pruebas/depuración)

Variables de entorno (con defaults de desarrollo, ver `docker-compose.yml`): `JWT_SECRET`, `DEMO_USER`, `DEMO_PASS`.

### Flujo de prueba manual

```bash
# 1. Login (usuario/clave de demo — no hay base de usuarios en el enunciado)
TOKEN=$(curl -s -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"usuario":"admin","clave":"admin123"}' | jq -r .token)

# 2. Pedir la factorización QR + estadísticas
curl -s -X POST http://localhost:3000/api/qr \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"matriz":[[12,-51,4],[6,167,-68],[-4,24,-41]]}'
```

## Contrato de las APIs

### `go-api` — POST /api/auth/login
```json
// request
{ "usuario": "admin", "clave": "admin123" }
// response 200
{ "token": "eyJ...", "expiraEn": "1h" }
```

### `go-api` — POST /api/qr  (requiere `Authorization: Bearer <token>`)
```json
// request
{ "matriz": [[12,-51,4],[6,167,-68],[-4,24,-41]] }

// response 200
{
  "matrizOriginal": [[12,-51,4],[6,167,-68],[-4,24,-41]],
  "factorizacionQR": { "q": [[...]], "r": [[...]] },
  "estadisticas": {
    "maximo": 70, "minimo": -175, "promedio": -8.968, "sumaTotal": -161.44,
    "matrizDiagonal": {
      "algunaEsDiagonal": false,
      "detalle": { "q": { "esDiagonal": false, "razon": "..." }, "r": { "esDiagonal": false, "razon": "..." } }
    }
  }
}
```

**Restricción:** la matriz debe ser rectangular y tener `filas >= columnas` (m ≥ n). Es la condición estándar para que exista una QR reducida con `Q` de columnas ortonormales y `R` cuadrada invertible; si `m < n` la API responde `422` con un mensaje explícito en vez de forzar un resultado ambiguo.

### `node-api` — POST /api/estadisticas  (requiere `Authorization: Bearer <token>`, normalmente invocado internamente por `go-api`)
```json
// request
{ "q": [[...]], "r": [[...]] }
// response 200
{ "maximo": 0, "minimo": 0, "promedio": 0, "sumaTotal": 0, "matrizDiagonal": { "algunaEsDiagonal": false, "detalle": {...} } }
```

## Decisiones técnicas a resaltar en la entrevista

- **Factorización QR vía reflexiones de Householder** (no Gram-Schmidt clásico): numéricamente más estable ante columnas casi colineales. Verificado con test que reconstruye `A = Q·R` ([go-api/matrix/qr_test.go](go-api/matrix/qr_test.go)).
- **"Alguna matriz es diagonal"** se evalúa por separado sobre `Q` y `R` (no sobre el conjunto combinado), porque "diagonal" es una propiedad de una matriz cuadrada individual, no de un conjunto de valores.
- **JWT compartido**: un único `JWT_SECRET` entre ambas APIs evita mantener dos sistemas de autenticación; `go-api` reenvía el mismo Bearer token del cliente en su llamada interna a `node-api`, así el servicio interno también queda protegido sin necesitar un token de servicio separado.
- **go-api como orquestador**: el cliente solo conoce un endpoint; la comunicación Go → Node queda encapsulada en el backend, no expuesta al frontend.
- **Manejo de errores de red**: si `node-api` no responde, `go-api` devuelve `502` con el detalle, en vez de fallar silenciosamente o colgarse (timeout de 5s en el cliente HTTP).

## Opcionales del enunciado no incluidos en esta iteración

Frontend y pruebas automatizadas de integración quedaron fuera de este alcance por decisión explícita de priorización; la lógica más sensible (la factorización QR) sí cuenta con test de corrección ([go-api/matrix/qr_test.go](go-api/matrix/qr_test.go)) verificado en Docker.
