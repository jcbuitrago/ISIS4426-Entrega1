import React from "react";
import { useNavigate } from "react-router-dom";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";
import { api, getAccessToken } from "../api.js";
import VideoPlayerModal from "../components/VideoPlayerModal.jsx";

export default function VotingGallery() {
  const navigate = useNavigate();
  const authed = !!getAccessToken();
  const [items, setItems] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [q, setQ] = React.useState("");
  const [player, setPlayer] = React.useState({ open: false, url: "", title: "" });
  const [votedIds, setVotedIds] = React.useState(new Set());
  const [remaining, setRemaining] = React.useState(2);

  React.useEffect(() => {
    if (!authed) return; // require login
    let mounted = true;
    (async () => {
      try {
        const [videos, my] = await Promise.all([
          api.listPublicVideos({ limit: 48, offset: 0 }),
          // Fix: Use proper API function instead of hardcoded fetch
          api.getMyVotes() // We need to create this function
        ]);
        if (!mounted) return;
        const mapped = (videos || []).map((d) => ({
          id: String(d.video_id),
          name: d.author,
          city: d.city,
          url: d.processed_url,
          votes: d.votes,
          title: d.title,
        }));
        setItems(mapped);
        if (my?.video_ids) setVotedIds(new Set(my.video_ids.map(String)));
        if (typeof my?.remaining === 'number') setRemaining(my.remaining);
      } catch (e) {
        setError(e?.message || "No se pudieron cargar los videos.");
      } finally {
        setLoading(false);
      }
    })();
    return () => { mounted = false };
  }, [authed]);

  const vote = async (id) => {
    try {
      await api.voteVideo(id);
      setItems((prev) => prev.map((it) => it.id === String(id) ? { ...it, votes: (it.votes || 0) + 1 } : it));
      setVotedIds(new Set([ ...Array.from(votedIds), String(id) ]));
      setRemaining((r) => Math.max(r - 1, 0));
    } catch (e) {
      alert(e?.message || "No se pudo registrar el voto.");
    }
  };
  
  const unvote = async (id) => {
    try {
      await api.unvoteVideo(id);
      setItems((prev) => prev.map((it) => it.id === String(id) ? { ...it, votes: Math.max((it.votes || 0) - 1, 0) } : it));
      const next = new Set(Array.from(votedIds)); 
      next.delete(String(id)); 
      setVotedIds(next);
      setRemaining((r) => Math.min(r + 1, 2));
    } catch (e) {
      alert(e?.message || "No se pudo retirar el voto.");
    }
  };

  const filtered = items.filter((it) => `${it.title} ${it.name} ${it.city}`.toLowerCase().includes(q.toLowerCase()));

  if (!authed) {
    return (
      <div className="bg-dark text-light min-vh-100 d-flex flex-column">
        <SiteHeader />
        <main className="container py-5 text-center">
          <h2 className="fw-bold">Inicia sesión para votar</h2>
          <p className="text-secondary">La votación requiere autenticación. Esto asegura 1 voto por persona.</p>
          <button className="btn btn-warning fw-bold" onClick={() => navigate("/login")}>Ir a iniciar sesión</button>
        </main>
        <SiteFooter />
      </div>
    );
  }

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />
      <main className="container py-5">
        <div className="d-flex flex-column flex-md-row align-items-md-center justify-content-between gap-3 mb-4">
          <div>
            <h2 className="display-6 fw-bold m-0">Galería para votar</h2>
            <p className="text-secondary m-0">Te quedan {remaining} voto(s).</p>
          </div>
          <div style={{ minWidth: 280 }}>
            <input className="form-control bg-black text-white border-secondary-subtle" placeholder="Buscar por jugador o título"
                   value={q} onChange={(e) => setQ(e.target.value)} />
          </div>
        </div>

        {loading && <p className="text-center text-secondary">Cargando…</p>}
        {error && <div className="alert alert-danger">{error}</div>}

        <div className="row g-4">
          {filtered.map((it) => {
            const voted = votedIds.has(String(it.id));
            return (
              <div key={it.id} className="col-12 col-sm-6 col-lg-4 col-xl-3">
                <div className="card bg-black border-0 rounded-4 overflow-hidden h-100">
                  <div className="position-relative" style={{ aspectRatio: "16/9", cursor: 'pointer' }} onClick={() => setPlayer({ open: true, url: it.url, title: it.title })}>
                    <video src={it.url} className="w-100 h-100" style={{ objectFit: "cover" }} muted playsInline preload="metadata" />
                  </div>
                  <div className="card-body d-flex flex-column">
                    <div className="flex-grow-1">
                      <p className="h6 fw-semibold mb-0">{it.title}</p>
                      <p className="text-secondary small mb-0">{it.name} · {it.city}</p>
                      <p className="text-secondary small mb-0">Votos: {it.votes || 0}</p>
                    </div>
                    {voted ? (
                      <button className="btn btn-outline-danger fw-bold mt-3" onClick={() => unvote(it.id)}>
                        <i className="bi bi-hand-thumbs-down me-2" /> Retirar voto
                      </button>
                    ) : (
                      <button className="btn btn-success fw-bold mt-3" onClick={() => vote(it.id)} disabled={remaining <= 0}>
                        <i className="bi bi-hand-thumbs-up me-2" /> Votar
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
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
