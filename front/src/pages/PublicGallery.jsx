import React from "react";
import { useNavigate } from "react-router-dom";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";
import { api } from "../api.js";
import VideoPlayerModal from "../components/VideoPlayerModal.jsx";

/**
 * PublicGallery (Showcase sin votación)
 */
export default function PublicGallery() {
  const navigate = useNavigate();
  const [items, setItems] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [q, setQ] = React.useState("");
  const [player, setPlayer] = React.useState({ open: false, url: "", title: "" });

  React.useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const data = await api.listPublicVideos({ limit: 48, offset: 0 });
        if (!mounted) return;
        const mapped = (data || []).map((d) => ({
          id: String(d.video_id),
          name: d.author,
          city: d.city,
          url: d.processed_url,
          thumb: d.thumb_url || d.processed_url,
          votes: d.votes,
          title: d.title,
        }));
        setItems(mapped);
      } catch (e) {
        setError(e?.message || "No se pudieron cargar los videos públicos.");
      } finally {
        setLoading(false);
      }
    })();
    return () => { mounted = false };
  }, []);

  const filtered = items.filter((it) => {
    const s = `${it.title} ${it.name} ${it.city}`.toLowerCase();
    return s.includes(q.toLowerCase());
  });

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />
      <main className="container py-5">
        <div className="d-flex flex-column flex-md-row align-items-md-center justify-content-between gap-3 mb-4">
          <div>
            <h2 className="display-5 fw-bold m-0">Galería Pública de Videos</h2>
            <p className="text-secondary m-0">Showcase de videos procesados. Usa el buscador para filtrar.</p>
          </div>
          <div style={{ minWidth: 280 }}>
            <input className="form-control bg-black text-white border-secondary-subtle" placeholder="Buscar por jugador o título"
                   value={q} onChange={(e) => setQ(e.target.value)} />
          </div>
        </div>

        {loading && <p className="text-center text-secondary">Cargando…</p>}
        {error && <div className="alert alert-danger">{error}</div>}

        <div className="row g-4">
          {filtered.map((it) => (
            <div key={it.id} className="col-12 col-sm-6 col-lg-4 col-xl-3">
              <div className="card bg-black border-0 rounded-4 overflow-hidden h-100">
                <div className="position-relative" style={{ aspectRatio: "16/9", cursor: "pointer" }} onClick={() => setPlayer({ open: true, url: it.url, title: it.title })}>
                  <img src={it.thumb} className="w-100 h-100" style={{ objectFit: "cover" }} alt={it.title} />
                  <div className="position-absolute bottom-0 start-0 p-2 bg-dark bg-opacity-50 w-100">
                    <span className="small">Votos: {it.votes || 0}</span>
                  </div>
                </div>
                <div className="card-body d-flex flex-column">
                  <div className="flex-grow-1">
                    <p className="h6 fw-semibold mb-0">{it.title}</p>
                    <p className="text-secondary small mb-0">{it.name} · {it.city}</p>
                  </div>
                  <button className="btn btn-outline-light mt-3" onClick={() => navigate("/galeria-votar")}>Ir a galería para votar</button>
                </div>
              </div>
            </div>
          ))}
          {!loading && filtered.length === 0 && (
            <p className="text-center text-secondary">No hay resultados para "{q}".</p>
          )}
        </div>
      </main>
      <VideoPlayerModal open={player.open} url={player.url} title={player.title} onClose={() => setPlayer({ open: false, url: "", title: "" })} />
      <SiteFooter />
    </div>
  );
}
