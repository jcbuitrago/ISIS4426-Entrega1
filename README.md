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

## Plan de pruebas (entregable)

En la carpeta **`Archivos_plan_de_pruebas.zip`** se encuentra:

* Un archivo **`.md`** con la **descripción del plan de pruebas**.
* Archivos preliminares/datasets para ejecutar las pruebas manuales y semi-automatizadas.
* Un **script en Python** para **generar gráficas** de resultados esperados.

---

