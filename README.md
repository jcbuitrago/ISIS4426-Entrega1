## Integrantes

|Nombre|Correo|Codigo|
|------|------|------|
|Francisco Santamaría|F.santamaria@uniandes.edu.co|202022134|
|David Octavio Ibarra Muñoz| d.ibarra@uniandes.edu.co| 202014446|
|Luis Fernando Ruiz| lf.ruizo1@uniandes.edu.co| 202211513|
|Juan Camilo Buitrago Ariza|Jc.buitragoa1@uniandes.edu.co | 201729194|

# ANB Showcase — Plataforma de videos con procesamiento asíncrono

Plataforma donde jugadores suben videos cortos demostrando habilidades de baloncesto. Los videos se **procesan en segundo plano** para asegurar escalabilidad y buena UX.

## Objetivos
* Subida de videos (20–60s, 1080p o superior).
* **Procesamiento asíncrono**:

  * Recorte a **máx. 30s**.
  * Normalización a **16:9, 720p**.
  * Inserción de **cortinilla de apertura y cierre (ANB)** con un máximo de **+5s** extra.
* Actualización automática de estado en la BD: `uploaded → processing → processed` (o `failed`).
* Exposición de endpoints públicos para listar y consultar videos procesados, y endpoints autenticados para gestión y votos.

---

## Arquitectura (resumen)

* **API (Go)**: subida de archivos, encolado de tareas, consultas (lista/detalle), votos.
* **Worker (Go + ffmpeg)**: consume cola, procesa video y actualiza BD.
* **Redis + Asynq**: cola de trabajos. **Asynqmon** para monitoreo.
* **PostgreSQL**: persistencia.
* **Frontend (Vite + NGINX)**: UI.
* **Volúmenes**: `/data` compartido entre API y worker para los archivos subidos y procesados.

---

## Requisitos

* Docker & Docker Compose
* (Opcional) `jq` para ver salida de tests en modo JSON
* (Opcional) `newman` para correr colección de Postman

---

## Variables de entorno (ejemplo)

Crea un archivo `.env` en la raíz:

```
POSTGRES_USER=anb
POSTGRES_PASSWORD=anbpass
POSTGRES_DB=anbdb
JWT_SECRET=devsecret123
```

---

## Ejecutar la app

```bash
docker compose up -d --build
```

Servicios:

* **Frontend**: [http://localhost:3000](http://localhost:3000)
* **API**: [http://localhost:8080](http://localhost:8080)

---

## Flujo de uso (alto nivel)

1. **Subir video** (multipart/form-data):

   * `POST /api/videos`
   * Campos: `title` (texto), `file` (mp4).
   * Respuesta `201/202`:

     ```json
     { "message":"Video subido correctamente. Procesamiento en curso.", "task_id":"<uuid>" }
     ```
2. **Monitorear la tarea** (opcional):

   * `GET /api/jobs/{id}` → estado `queued|processing|done|failed`.
   * Asynqmon en `:8081` para ver la cola y workers.
3. **Consultar listados**:

   * **Público** procesados (para la UI): `GET /api/public/videos?limit=&offset=`
   * **General** (admin/dev): `GET /api/videos`
   * **De un usuario** (JWT): `GET /api/users/{id}/videos`
4. **Detalle de un video**:

   * `GET /api/videos/{id}` → incluye `original_url`, `processed_url`, `status`, timestamps y `votes`.
5. **Votar / retirar voto** (JWT):

   * `POST /api/public/videos/{id}/vote`
   * `DELETE /api/public/videos/{id}/vote`

Estados posibles: `uploaded`, `processing`, `processed`, `failed`.

---

## Tests

### Services

```bash
cd back
go test ./app/services -json | jq .
# o verbose:
go test ./app/services -v
```

### Routers

```bash
cd back
go test ./app/routers -json | jq .
# o verbose:
go test ./app/routers -v
```

> Los tests de routers usan `sqlmock` y `httptest` para endpoints públicos (listados, rankings, 401 en endpoints que requieren JWT).

---

## Postman

Ejecuta la colección:

```bash
newman run postman/API.postman_collection.json
```

> La colección incluye: **Upload**, **Job status**, **List/Detail**, **Votar/Unvotar** y ejemplos de éxito/error.

---

## Pruebas de carga (k6)

- **Script**: `load-tests/k6-load-scenarios.js`
- **Config base**: duplica `load-tests/env.loadtest.example` → `load-tests/env.loadtest.aws` y ajusta las variables según tu infraestructura.
- **Tooling**: usamos [k6](https://github.com/grafana/k6) porque ofrece métricas detalladas sin requerir un stack complejo. El mismo script se puede ejecutar localmente o desde un contenedor liviano.

### Paso a paso

1. **Instala k6 o usa Docker**
   - Local (Windows/macOS/Linux): sigue la guía oficial https://k6.io/docs/get-started/installation/
   - Docker (recomendado para mantener el entorno limpio):
     ```bash
     docker pull grafana/k6
     ```
2. **Configura las variables de entorno**
   - Copia el archivo de ejemplo y cámbialo con tus datos reales:
     ```bash
     cp load-tests/env.loadtest.example load-tests/env.loadtest.aws
     ```
   - Ajusta los campos clave (detalles abajo) con las IP/hostnames de tus instancias EC2, el bucket S3 y el token JWT que vayas a utilizar.
3. **Verifica conectividad desde el generador de carga**
   - Abre los puertos 80/3000 (frontend) y 8080 (API) en los Security Groups para la IP pública desde donde correrás k6.
   - Si usas un Application/Network Load Balancer, apunta `API_BASE_URL` y `FRONT_BASE_URL` al DNS del ALB/NLB en vez de la IP directa de la instancia.
4. **Ejecuta el escenario**
   - Con Docker (acepta archivos `--env-file`, ideal cuando tienes muchas variables):
     ```bash
     docker run --rm -it ^
       --env-file load-tests/env.loadtest.aws ^
       -v "%cd%":/scripts grafana/k6 run /scripts/load-tests/k6-load-scenarios.js
     ```
     *(en Bash reemplaza los `^` por `\` y usa `${PWD}`)*.
   - Ejecución local (PowerShell) sin Docker:
     ```powershell
     Get-Content load-tests\env.loadtest.aws | ForEach-Object {
       if ($_ -match '^(.*)=(.*)$') { $name=$matches[1]; $value=$matches[2]; if ($name -and $value) { set-item -path env:$name -value $value } }
     }
     k6 run load-tests/k6-load-scenarios.js
     ```
5. **Analiza métricas**
   - k6 mostrará throughput, latencias (p(95/p(99)), tasa de errores, y nuestras métricas personalizadas: `upload_duration`, `s3_asset_duration`, `vote_failures`.
   - Exporta a CloudWatch/Influx/Grafana si necesitas históricos (`k6 run --out cloudwatch`).

### Variables que debes personalizar

| Variable | Para qué sirve | Ejemplo en AWS |
| --- | --- | --- |
| `API_BASE_URL` | URL base del backend (`/api`) detrás del ALB o IP pública de la instancia | `http://api.anb-showcase.internal:8080/api` o `http://54.180.x.x:8080/api` |
| `FRONT_BASE_URL` | URL del frontend en NGINX/EC2 o CloudFront | `https://showcase.uniandes.edu` |
| `VIDEO_FILE` | Ruta local del MP4 que se subirá en el escenario de uploads (puedes reutilizar `assets/intro.mp4`) | `assets/intro.mp4` |
| `JWT_TOKEN` | Token válido para un usuario con permisos de upload/vote | `eyJhbGciOiJIUzI1NiIs...` |
| `ENABLE_UPLOADS` | Activa el escenario `POST /api/videos`. Requiere `JWT_TOKEN`, `VIDEO_FILE` y abrir el puerto 8080 hacia S3 (si la instancia procesa y sube al bucket). | `true`/`false` |
| `ENABLE_VOTES` | Activa el flujo `POST/DELETE /api/public/videos/{id}/vote`. También necesita `JWT_TOKEN`. | `true`/`false` |
| `CHECK_ASSETS` | Hace peticiones directas al `processed_url` (S3 o CloudFront) de videos aleatorios para validar que los archivos generados son accesibles. Útil para detectar permisos en el bucket. | `true` si tu bucket permite lectura pública o via CloudFront |
| `PUBLIC_START_RATE`, `PUBLIC_PEAK_RATE`, `PUBLIC_END_RATE`, `PUBLIC_MIN_VUS`, `PUBLIC_MAX_VUS` | Controlan la rampa de usuarios que consultan el carrusel público. Ajusta según el tráfico objetivo de tu entrega (ej. 50 req/s sostenidos). | `5 / 60 / 10 / 20 / 200` |
| `UPLOAD_RATE`, `UPLOAD_DURATION` | Cantidad de uploads concurrentes. Considera los límites de ancho de banda y CPU de la instancia que ejecuta ffmpeg. | `3` y `10m` |
| `ASSET_TIMEOUT` | Tiempo máximo para esperar respuesta del archivo en S3/CloudFront. Si tu bucket está en otra región con más latencia, aumenta el valor (ej. `8s`). | `5s` |

> Si vas a apuntar el escenario contra direcciones privadas (VPC), levanta k6 dentro de la misma red (por ejemplo, una tercera instancia EC2 "bastion" que solo se usa para pruebas de carga).

### ¿Qué valida cada escenario?

- `browse_public`: stress de `GET /api/public/videos` y verificación opcional de los objetos en S3 (`processed_url`).
- `browse_frontend`: tráfico constante al sitio estático (ideal si lo sirves con NGINX en EC2).
- `upload_videos` (opcional): sube videos simulados para forzar la cola Asynq, el worker y la escritura en S3.
- `vote_cycle` (opcional): combina `POST` y `DELETE` de votos para validar bloqueos de la BD y límites de rate.

Cada función expone checks y métricas custom (trends/counters) que te ayudan a correlacionar cuellos de botella con CloudWatch (CPU, disco, throughput del ALB).

### Recursos recomendados

- Repositorio oficial de k6: https://github.com/grafana/k6
- Ejemplos listos para producción (templates): https://github.com/grafana/k6-template-esbuild

---

## Plan de pruebas (entregable)

En la carpeta **`Archivos_plan_de_pruebas.zip`** se encuentra:

* Un archivo **`.md`** con la **descripción del plan de pruebas**.
* Archivos preliminares/datasets para ejecutar las pruebas manuales y semi-automatizadas.
* Un **script en Python** para **generar gráficas** de resultados esperados.

---

