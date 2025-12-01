import matplotlib.pyplot as plt

# === Datos simulados ===
# Escenario 1: Carga/Descarga
usuarios_carga = [10, 50, 100, 200, 300]
tiempos_carga = [800, 1200, 1800, 2500, 3200]  # ms
throughput_carga = [20, 80, 150, 230, 280]     # videos/min
cpu_carga = [15, 30, 45, 65, 85]               # % CPU
mem_carga = [20, 35, 50, 70, 85]               # % RAM

# Escenario 2: Reproducción
usuarios_playback = [20, 100, 250, 500]
tiempos_playback = [500, 800, 1200, 1600]      # ms
concurrentes_ok = [20, 100, 230, 450]          # usuarios sin buffering
cpu_playback = [10, 25, 50, 75]                # % CPU
mem_playback = [15, 30, 55, 80]                # % RAM

# === Escenario 1: Carga/Descarga ===

# Tiempo de respuesta vs usuarios
plt.figure(figsize=(7,5))
plt.plot(usuarios_carga, tiempos_carga, marker="o")
plt.axhline(3000, color="r", linestyle="--", label="Límite aceptación 3000 ms")
plt.title("Escenario 1: Tiempo de Respuesta vs Usuarios (Carga/Descarga)")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("Tiempo de respuesta (ms)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_tiempo_carga.png")

# Throughput vs usuarios
plt.figure(figsize=(7,5))
plt.plot(usuarios_carga, throughput_carga, marker="s", color="green")
plt.axhline(50, color="r", linestyle="--", label="Criterio mínimo 50/min")
plt.title("Escenario 1: Throughput vs Usuarios (Carga/Descarga)")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("Throughput (videos/min)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_throughput_carga.png")

# CPU vs usuarios
plt.figure(figsize=(7,5))
plt.plot(usuarios_carga, cpu_carga, marker="^", color="purple")
plt.axhline(70, color="r", linestyle="--", label="Criterio CPU < 70%")
plt.title("Escenario 1: Uso de CPU vs Usuarios (Carga/Descarga)")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("CPU (%)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_cpu_carga.png")

# RAM vs usuarios
plt.figure(figsize=(7,5))
plt.plot(usuarios_carga, mem_carga, marker="d", color="orange")
plt.axhline(75, color="r", linestyle="--", label="Criterio RAM < 75%")
plt.title("Escenario 1: Uso de RAM vs Usuarios (Carga/Descarga)")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("RAM (%)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_ram_carga.png")

# === Escenario 2: Reproducción ===

# Tiempo de inicio reproducción vs usuarios
plt.figure(figsize=(7,5))
plt.plot(usuarios_playback, tiempos_playback, marker="^", color="purple")
plt.axhline(1500, color="r", linestyle="--", label="Límite aceptación 1500 ms")
plt.title("Escenario 2: Tiempo de Inicio Reproducción vs Usuarios")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("Tiempo inicio (ms)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_tiempo_playback.png")

# Sesiones concurrentes sin buffering
plt.figure(figsize=(7,5))
plt.plot(usuarios_playback, concurrentes_ok, marker="d", color="orange")
plt.axhline(200, color="r", linestyle="--", label="Criterio mínimo 200 usuarios sin buffering")
plt.title("Escenario 2: Sesiones Concurrentes sin Buffering")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("Usuarios reproducidos sin buffering")
plt.legend()
plt.grid(True)
plt.savefig("grafico_buffering_playback.png")

# CPU vs usuarios (playback)
plt.figure(figsize=(7,5))
plt.plot(usuarios_playback, cpu_playback, marker="s", color="green")
plt.axhline(70, color="r", linestyle="--", label="Criterio CPU < 70%")
plt.title("Escenario 2: Uso de CPU vs Usuarios (Reproducción)")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("CPU (%)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_cpu_playback.png")

# RAM vs usuarios (playback)
plt.figure(figsize=(7,5))
plt.plot(usuarios_playback, mem_playback, marker="o", color="blue")
plt.axhline(75, color="r", linestyle="--", label="Criterio RAM < 75%")
plt.title("Escenario 2: Uso de RAM vs Usuarios (Reproducción)")
plt.xlabel("Usuarios concurrentes")
plt.ylabel("RAM (%)")
plt.legend()
plt.grid(True)
plt.savefig("grafico_ram_playback.png")

print("✅ Gráficas generadas para todos los escenarios y métricas esperadas.")