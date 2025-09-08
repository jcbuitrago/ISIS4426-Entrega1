import React from "react";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

/**
 * PlayerDashboard
 * TODO: Conectar datos reales con tu backend:
 * - fetch de perfil del jugador
 * - fetch de videos del jugador
 * - acciones: editar perfil, subir video, borrar video
 */
export default function PlayerDashboard({
  onEditProfile = () => {},           // TODO: abrir modal o navegar a edición
  onUploadVideo = () => {},            // TODO: abrir selector/drag&drop y POST
  onDeleteVideo = (id) => {},          // TODO: DELETE /api/videos/:id
}) {
  // placeholders (reemplázalos por datos del backend)
  const player = {
    name: "Ethan Carter",
    city: "Los Angeles, USA",
    id: "12345",
    avatar: "https://lh3.googleusercontent.com/aida-public/AB6AXuD_bMnsJOQpHI-KERPOwPbW8o4K3imF6FmJRblAelnUceHUBWY_RlIa375NBY_FVpAXTFR_P2MlrdUyGtfzETw4fWVvhHreFnemzbubOs7yq5i2jokgDXeBg8oVDBfGIIs1eFqzz9j4tKws6c-z8pSpBqLdIGmlljhHi35kNwitwAAqMxqg64I51JGnd9OGFF36e9MWV2DwroJGWzTPFZ0RI7JsgajburWJC-JdyO38VZ_a2yyS5om8-wYcXrfYsqGH4ZB0YvTkd5g5",
  };

  const videos = [
    { id: "v1", title: "Highlight Reel 2023", status: "Uploaded" },
    { id: "v2", title: "Game Highlights vs. Titans", status: "Processed" },
    { id: "v3", title: "Skills Showcase", status: "Uploaded" },
  ];

  const statusPill = (status) => {
    const map = {
      Uploaded: { bg: "bg-success-subtle", text: "text-success-emphasis", dot: "bg-success" },
      Processed: { bg: "bg-primary-subtle", text: "text-primary-emphasis", dot: "bg-primary" },
    };
    const s = map[status] || map.Uploaded;
    return (
      <span className={`d-inline-flex align-items-center rounded-pill px-3 py-1 text-uppercase ${s.bg} ${s.text}`} style={{ fontSize: 12, fontWeight: 600 }}>
        <span className={`me-2 rounded-circle ${s.dot}`} style={{ width: 8, height: 8 }} />
        {status}
      </span>
    );
  };

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />
      <main className="flex-grow-1 container py-5">
        <div className="mb-4">
          <h2 className="display-6 fw-bold">Player Dashboard</h2>
          <p className="text-secondary">Manage your profile and video submissions.</p>
        </div>

        <div className="row g-4">
          {/* Col izquierda - Perfil */}
          <div className="col-12 col-lg-4">
            <div className="p-4 rounded-4 border border-secondary-subtle bg-black">
              <div className="d-flex flex-column align-items-center text-center">
                <div
                  className="rounded-circle border border-secondary-subtle mb-3"
                  style={{
                    width: 128,
                    height: 128,
                    backgroundImage: `url(${player.avatar})`,
                    backgroundSize: "cover",
                    backgroundPosition: "center",
                  }}
                />
                <p className="h5 fw-bold mb-0">{player.name}</p>
                <p className="text-secondary mb-0">{player.city}</p>
                <p className="text-secondary small">ID: {player.id}</p>
              </div>

              <button
                className="btn btn-secondary w-100 mt-3 d-flex align-items-center justify-content-center gap-2"
                onClick={onEditProfile}
              >
                <i className="bi bi-pencil" />
                Edit Profile
              </button>
            </div>
          </div>

          {/* Col derecha - Videos */}
          <div className="col-12 col-lg-8">
            <div className="p-4 rounded-4 border border-secondary-subtle bg-black">
              <div className="d-flex flex-column flex-sm-row align-items-start align-items-sm-center justify-content-between gap-3">
                <h3 className="h4 fw-bold mb-0">My Videos</h3>
                <button className="btn btn-success fw-bold rounded-pill d-flex align-items-center gap-2"
                        onClick={onUploadVideo}>
                  <i className="bi bi-upload" />
                  Upload Video
                </button>
              </div>

              <div className="mt-3 border rounded-3 border-secondary-subtle">
                <div className="table-responsive">
                  <table className="table table-dark table-hover align-middle mb-0">
                    <tbody>
                      {videos.map((v) => (
                        <tr key={v.id}>
                          <td className="px-3 py-3 fw-medium">{v.title}</td>
                          <td className="px-3 py-3">{statusPill(v.status)}</td>
                          <td className="px-3 py-3 text-end">
                            <button
                              className="btn btn-sm btn-link text-danger"
                              onClick={() => onDeleteVideo(v.id)}
                              aria-label={`Eliminar ${v.title}`}
                            >
                              <i className="bi bi-trash fs-5" />
                            </button>
                          </td>
                        </tr>
                      ))}
                      {videos.length === 0 && (
                        <tr>
                          <td className="text-center text-secondary py-4" colSpan={3}>
                            No hay videos todavía. ¡Sube tu primer video!
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

            </div>
          </div>
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
