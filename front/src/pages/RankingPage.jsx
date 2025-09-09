import React from "react";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";
import { api } from "../api.js";

/**
 * RankingPage
 */
export default function RankingPage() {
  const [rows, setRows] = React.useState([]);
  const [city, setCity] = React.useState("");
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");

  const load = React.useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await api.rankings({ city: city || undefined });
      const mapped = (data || []).map((r) => ({
        rank: r.position,
        name: r.username,
        city: r.city,
        score: r.votes,
      }));
      setRows(mapped);
    } catch (e) {
      setError(e?.message || "No se pudo cargar el ranking");
    } finally {
      setLoading(false);
    }
  }, [city]);

  React.useEffect(() => { load(); }, [load]);

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />

      <main className="container py-5">
        <div className="mb-4 d-flex flex-column flex-md-row align-items-md-end gap-3 justify-content-between">
          <div>
            <h1 className="display-6 fw-bold">Ranking de Jugadores</h1>
            <p className="text-secondary">Clasificación actualizada de los mejores talentos emergentes.</p>
          </div>
          <div className="d-flex align-items-end gap-2">
            <div>
              <label className="form-label text-secondary">Filtrar por ciudad</label>
              <input
                type="text"
                className="form-control bg-black text-white border-secondary-subtle"
                placeholder="Ciudad (opcional)"
                value={city}
                onChange={(e) => setCity(e.target.value)}
              />
            </div>
            <button className="btn btn-outline-light mt-4" onClick={load}>Aplicar</button>
          </div>
        </div>

        {loading && <p className="text-center text-secondary">Cargando…</p>}
        {error && <div className="alert alert-danger">{error}</div>}

        <div className="border rounded-4 border-secondary-subtle bg-black">
          <div className="table-responsive">
            <table className="table table-dark table-hover align-middle mb-0">
              <thead>
                <tr className="text-secondary">
                  <th className="text-center" style={{ width: 64 }}>#</th>
                  <th>Jugador</th>
                  <th>Ciudad</th>
                  <th className="text-end">Puntuación</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => (
                  <tr key={r.rank} className="table-row">
                    <td className={`text-center fs-5 ${r.rank === 1 ? "text-success" : ""}`}>{r.rank}</td>
                    <td className="fw-semibold">{r.name}</td>
                    <td className="text-secondary">{r.city}</td>
                    <td className="text-end fw-semibold">{Number(r.score || 0).toLocaleString()}</td>
                  </tr>
                ))}
                {!loading && rows.length === 0 && (
                  <tr>
                    <td colSpan={4} className="text-center text-secondary py-4">No hay datos de ranking.</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </main>

      <SiteFooter />
    </div>
  );
}
