# Informe de Resultados de Pruebas de Carga - ANB Showcase

**Fecha:** 1 de Diciembre de 2025
**Archivo analizado:** `load-tests/summary_20_180.json`
**Basado en:** `Archivos_plan_de_pruebas/plan_de_pruebas.md`

---

## 1. Resumen Ejecutivo

Las pruebas de carga se realizaron escalando desde usuarios concurrentes hasta un pico de **683 VUs** (Virtual Users). Se observó un comportamiento mixto:

*   **Capacidad de Carga (Uploads):** El sistema manejó un volumen considerable de subidas (~3,575 exitosas) con un tiempo promedio de **1.67s**, lo cual es positivo, aunque el percentil 95 (5.06s) supera el umbral ideal.
*   **Reproducción y Activos (Assets):** Se detectó un **cuello de botella crítico** en la descarga de videos procesados (`processed asset 200`). Hubo **11,130 fallos** al intentar acceder a los assets, con tiempos de respuesta muy altos (P95 ~7.3s). Esto sugiere problemas con el almacenamiento (S3), la CDN o que los videos no se procesaron a tiempo.
*   **Estabilidad:** Se registraron **11,385 iteraciones descartadas** (`dropped_iterations`), indicando que los inyectores de carga o el sistema bajo prueba no pudieron mantener el ritmo solicitado. La tasa de error global de peticiones HTTP fue del **36%**, principalmente debido a los fallos en la recuperación de assets.

---

## 2. Análisis Detallado por Escenario

### Escenario 1: Carga de Videos (Uploads)
*   **Métrica clave:** `http_req_duration{scenario:upload_videos}`
*   **Resultados:**
    *   Promedio: **1.67 s** (Cumple objetivo < 3s en promedio)
    *   P95: **5.06 s** (Excede objetivo < 3s)
    *   Máximo: ~59.8 s (Casos extremos de latencia)
    *   Éxitos: 3,575 uploads aceptados (HTTP 201/202).

### Escenario 2: Reproducción (Browse Public & Assets)
*   **Métrica clave:** `s3_asset_duration` y `http_req_duration{scenario:browse_public}`
*   **Resultados:**
    *   **Tiempo de asset (S3):** P95 de **7.27 s**. (No cumple objetivo < 1.5s).
    *   **Navegación pública:** P95 de **7.09 s**.
    *   **Tasa de éxito de assets:** Crítica. Solo 823 descargas exitosas vs 11,130 fallidas. Esto invalida la prueba de fluidez de reproducción para la mayoría de usuarios.

### Escenario 3: Navegación Frontend (Estático)
*   **Métrica clave:** `http_req_duration{scenario:browse_frontend}`
*   **Resultados:**
    *   Promedio: **0.51 s**.
    *   P95: **1.23 s**.
    *   Comportamiento estable.

### Escenario 4: Votación (Vote/Unvote)
*   **Métrica clave:** `http_req_duration{scenario:vote_cycle}`
*   **Resultados:**
    *   Promedio: **0.87 s**.
    *   P95: **1.76 s**.
    *   Se procesaron ~6,543 ciclos de voto/desvoto exitosamente.

---

## 3. Comparación con Criterios de Aceptación

| Escenario | Usuarios (VUs Max) | Métrica Clave (Target) | Resultado Obtenido | Estado |
|-----------|--------------------|------------------------|--------------------|--------|
| **1. Carga de Videos** | 683 | Latencia < 3 s | Avg: 1.67s / P95: 5.06s | **Parcial** (Avg OK, P95 Alto) |
| **2. Reproducción** | 683 | Inicio < 1.5 s | P95: 7.27s | **Fallido** |
| | | Fluidez (Errores) | 11k fallos (93% error en assets) | **Fallido** |
| **General** | 683 | HTTP Req Failed < 2% | 36% | **Fallido** |
| **General** | 683 | Dropped Iterations = 0 | 11,385 | **Fallido** |

---

## 4. Hallazgos y Recomendaciones

1.  **Investigar Errores de Assets (S3):** La mayoría de los fallos provienen de `processed asset 200`. Verificar:
    *   Si el *Worker* está procesando los videos a la velocidad adecuada.
    *   Si los archivos realmente existen en S3 cuando se solicitan.
    *   Si hay limitaciones de ancho de banda o *throttling* en el bucket S3.
2.  **Optimizar Tiempos de Respuesta Públicos:** El endpoint `/api/public/videos` tiene una latencia P95 de 7s, lo cual es inaceptable para la experiencia de usuario. Revisar índices de base de datos o caché.
3.  **Ajustar Capacidad del Sistema:** La gran cantidad de `dropped_iterations` sugiere que el sistema (o el agente de pruebas) se saturó. Si el servidor de aplicaciones tiene CPU/RAM al límite, se requiere escalar horizontalmente (más réplicas) o verticalmente.
4.  **Latencia de Upload:** Aunque el promedio es bueno, los picos de 60s indican que algunos uploads se quedan colgados. Revisar configuración de timeouts en NGINX y Go.

---
*Generado automáticamente basado en los resultados de `summary_20_180.json`.*

## 5. Gráficas de Evolución

### 5.1. Evolución de Carga (Usuarios Virtuales)
![Evolución de VUs](load-test/report_assets/graph_vus.png)
*La gráfica muestra cómo la carga de usuarios (VUs) aumentó progresivamente durante la prueba hasta alcanzar el pico.*

### 5.2. Latencia por Escenario (Escala Logarítmica)
![Latencias](load-test/report_assets/graph_latency_log.png)
*Distribución de tiempos de respuesta en el tiempo. Se utiliza escala logarítmica para visualizar tanto las peticiones rápidas como los picos de latencia extremos (outliers).*


