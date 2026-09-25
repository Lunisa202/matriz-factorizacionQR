# Testing manual — guía paso a paso (despliegue Render)

> Cómo probar el proyecto a mano contra el **despliegue en Render** (evaluadores, QA, demo).
> Valores verificados en vivo el 2026-09-25: Q·R≈A ✔, stats coherentes ✔.
> Para local sustituye las bases por `http://localhost:3001` (Go) y `http://localhost:3002` (Node).
> Referencia de endpoints y códigos: `README.md` §5.

---

## 0. URLs del despliegue

| Servicio | Base Render | Swagger |
|---|---|---|
| Go (flujo completo) | https://reto-api-go-e11w.onrender.com | https://reto-api-go-e11w.onrender.com/docs/index.html |
| Node (directo) | https://reto-api-node-jusy.onrender.com | https://reto-api-node-jusy.onrender.com/docs |

En adelante `{{GO}}` = `https://reto-api-go-e11w.onrender.com` y
`{{NODE}}` = `https://reto-api-node-jusy.onrender.com`.
Si el primer request tarda o da 502, **reintenta una vez** (cold-start del free tier:
despierta primero `{{NODE}}/health` y luego prueba).
| Node Swagger | http://localhost:3002/docs | https://reto-api-node-jusy.onrender.com/docs |

En adelante `{{GO}}` = base de Go y `{{NODE}}` = base de Node según la columna que uses.
En Render, si el primer request tarda o da 502, **reintenta una vez** (cold-start del free tier).

---

## 1. Obtener el token (vale para ambas APIs)

Swagger: `POST {{GO}}/docs → /api/auth/token` → *Try it out* → body:

```json
{ "user": "demo" }
```

*Execute* → copiar `token`. Luego **Authorize** 🔓 → `Bearer <token>` → *Authorize*.

cURL:

```bash
TOKEN=$(curl -s -X POST {{GO}}/api/auth/token \
  -H "Content-Type: application/json" -d '{"user":"demo"}' \
  | python -c "import sys,json;print(json.load(sys.stdin)['token'])")
echo $TOKEN   # debe imprimir un JWT de 3 partes separadas por puntos
```

✅ Esperado: `200 {"token": "eyJ..."}` · ❌ `404` = `ENABLE_DEMO_TOKEN=false` en ese entorno.

---

## 2. Caso feliz: matriz → QR → estadísticas (TM-01)

Swagger: `POST /api/matrix/qr` (Go, con Authorize activo) → body:

```json
{ "matrix": [[1, 2], [3, 4], [5, 6]] }
```

cURL:

```bash
curl -s -X POST {{GO}}/api/matrix/qr \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"matrix": [[1,2],[3,4],[5,6]]}' | python -m json.tool
```

✅ Esperado `200` con estos valores exactos (confirmados contra Render el 2026-09-25):

- `q` (3×2): `[[0.1690, 0.8970], [0.5070, 0.2760], [0.8451, -0.3450]]`
- `r` (2×2): `[[5.9160, 7.4373], [0, 0.8280]]` (el `0` confirma triangular superior)
- `stats`: `max=7.4373`, `min=-0.3450`, `sum=16.5308`, `count=10`, `average=1.6530`,
  `isDiagonal=false`, `diagonalMatrix=null`, `matricesAnalyzed=["q","r"]`

Comprobaciones mentales: `Q·R` reconstruye la matriz · `sum/count = average` (16.5308/10).

---

## 3. Node directo con matrices diagonales (TM-02)

Swagger/cURL contra `{{NODE}}/api/stats` (mismo Bearer):

```json
{ "q": [[1, 0], [0, 1]], "r": [[5, 0], [0, 3]] }
```

```bash
curl -s -X POST {{NODE}}/api/stats \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"q": [[1,0],[0,1]], "r": [[5,0],[0,3]]}' | python -m json.tool
```

✅ Esperado `200`: `max=5`, `min=0`, `sum=10`, `count=8`, `average=1.25`,
`isDiagonal=true`, `diagonalMatrix="q"`.

---

## 4. Casos de error (TM-03 a TM-07)

| ID | Prueba | Cómo | Esperado |
|---|---|---|---|
| TM-03 | Sin token | `POST {{GO}}/api/matrix/qr` sin header `Authorization` | `401 {"error":"missing Authorization header","code":401}` |
| TM-04 | Token inválido | `Authorization: Bearer abc` | `401 {"error":"invalid or expired token","code":401}` |
| TM-05 | Matriz no rectangular | `{"matrix": [[1,2],[3]]}` con token | `400` con mensaje `must be rectangular` |
| TM-06 | Matriz vacía | `{"matrix": []}` con token | `400` |
| TM-07 | Matriz gigante | 101×101 unos (script abajo) con token | `413` con mensaje `exceed the limit` |

Script TM-07 (genera el body con Python):

```bash
python -c "import json; print(json.dumps({'matrix': [[1]*101 for _ in range(101)]}))" > big.json
curl -s -X POST {{GO}}/api/matrix/qr \
  -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  --data @big.json | python -m json.tool; rm -f big.json
```

---

## 5. Node inalcanzable (fail-soft 502) (TM-08)

Caso observado en este despliegue: con `NODE_API_URL` mal configurado, TM-01 devolvió `502`
con `q` y `r` presentes + `error: "qr computed but forwarding to node failed: ..."` (el cálculo
no se pierde aunque Node falle). Tras corregir la URL a
`https://reto-api-node-jusy.onrender.com/api/stats`, TM-01 volvió a `200`.

En local se reproduce a voluntad: `docker stop api-node` → TM-01 da `502` con datos →
`docker start api-node` → `200` normal.

---

## 6. Checklist de aceptación (para dar el visto bueno)

- [ ] TM-01 devuelve `q`, `r` y `stats` con los valores exactos de §2.
- [ ] TM-02 devuelve `isDiagonal=true` y `diagonalMatrix="q"`.
- [ ] TM-03/TM-04 → 401 · TM-05/TM-06 → 400 · TM-07 → 413.
- [ ] Swagger `/docs` carga en ambas APIs y *Try it out* funciona con el Bearer.
- [ ] En Render: ambos `/health` responden `ok` (tras el cold-start).
