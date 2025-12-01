# Plan de Pruebas de Carga – Plataforma de Videos de Básquetbol

## 1. Objetivos
- Evaluar la capacidad máxima de la plataforma para soportar operaciones de carga, descarga y reproducción de videos.  
- Identificar cuellos de botella en el backend en Go y la infraestructura cloud subyacente.  
- Establecer métricas de desempeño aceptables en términos de throughput, tiempos de respuesta y utilización de recursos.  
- Preparar la infraestructura para pruebas futuras de estrés y escalabilidad.  
- Validar la robustez del sistema ante archivos inválidos o escenarios de error.  

---

## 2. Entorno de Pruebas
- **Infraestructura Cloud:** Docker + docker-compose (Postgres, Redis, App, K6).  
- **Lenguaje backend:** Go.  
- **Almacenamiento de videos:** `/data/processed`.  
- **Reproducción:** vía API REST + servicio de streaming.  
- **Herramienta de pruebas:** k6 (Docker image `grafana/k6`).  

---

## 3. Criterios de Aceptación
- **Tiempo de respuesta:**  
  - Carga/descarga de video < 3 s para archivos ≤ 100 MB.  
  - Inicio de reproducción < 1.5 s.  
  - Rechazo de archivos inválidos < 2 s.  
- **Throughput esperado:**  
  - Al menos 50 videos cargados/descargados por minuto en concurrencia moderada.  
  - Al menos 200 sesiones concurrentes de reproducción con fluidez.  
- **Utilización de recursos:**  
  - CPU < 70% en promedio.  
  - Memoria < 75% en promedio.  

---

## 4. Escenarios de Prueba

### Escenario 1 – Carga y Descarga de Videos
- Validar upload y download de archivos válidos hasta 100 MB.  

### Escenario 2 – Reproducción de Videos
- Validar inicio de streaming (<1.5 s) y fluidez sin buffering.  

### Escenario 3 – Upload con resolución incorrecta
- Envío de archivo `.mp4` válido pero con resolución no soportada (ej. 4K).  
- **Esperado:** rechazo inmediato con código `400` o `422`.  

### Escenario 4 – Upload con formato incorrecto
- Envío de archivo no multimedia (ej. `.txt`).  
- **Esperado:** rechazo inmediato con código `415 Unsupported Media Type`.  

### Escenario 5 – Upload con nombres duplicados
- Envío concurrente de archivos con mismo nombre.  
- **Esperado:** el sistema debe evitar sobrescritura (respuesta `409 Conflict` o renombrado automático).  

---

## 5. Estrategia de Pruebas
- **Fases:** smoke test, carga progresiva, estrés.  
- **Monitoreo:** k6 outputs + logs del contenedor + métricas de Docker stats.  

---

## 6. Resultados Esperados
- Identificación del punto máximo de concurrencia.  
- Validación de que se cumplen los criterios de aceptación.  
- Confirmación de que errores comunes (resolución, formato, duplicados) son manejados correctamente sin afectar la estabilidad del sistema.  

---

## 7. Tabla Resumen de Escenarios

| Escenario | Usuarios Concurrentes | Métricas Clave | Criterios de Aceptación |
|-----------|------------------------|----------------|--------------------------|
| 1. Carga/Descarga de Videos | 10 → 300 | Latencia, throughput, CPU/RAM | < 3 s por carga, throughput ≥ 50/min |
| 2. Reproducción de Videos   | 20 → 500 | Tiempo de inicio, fluidez | < 1.5 s inicio, ≥ 200 concurrentes sin buffering |
| 3. Upload Resolución Incorrecta | 10 → 100 | Tiempo rechazo, HTTP code | Respuesta < 2 s, error 400/422 |
| 4. Upload Formato Incorrecto    | 10 → 100 | Tiempo rechazo, HTTP code | Respuesta < 1 s, error 415 |
| 5. Upload Nombre Duplicado      | 10 → 200 | Latencia, tasa de errores | No sobrescribir, error 409 o renombrado |
