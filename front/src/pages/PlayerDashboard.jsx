import React from "react";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";
import { api, decodeAccessToken, getAccessToken } from "../api.js";

/**
 * PlayerDashboard
 */
export default function PlayerDashboard({
  onEditProfile = () => {},
}) {
  const [videos, setVideos] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [uploading, setUploading] = React.useState(false);

  const [me, setMe] = React.useState(null);
  const [saving, setSaving] = React.useState(false);
  const [avatarUploading, setAvatarUploading] = React.useState(false);

  const tokenPayload = decodeAccessToken();
  const userId = tokenPayload?.user_id;

  const load = React.useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      if (getAccessToken()) {
        const my = await api.getMe();
        setMe(my);
      }
      const list = await api.listMyVideos({ user_id: userId });
      const mapped = (list || []).map((v) => ({
        id: v.video_id,
        title: v.title,
        status: v.status === "processed" || v.status === "Processed" ? "Processed" : "Uploaded",
      }));
      setVideos(mapped);
    } catch (e) {
      setError(e?.message || "No se pudieron cargar tus datos.");
    } finally {
      setLoading(false);
    }
  }, [userId]);

  React.useEffect(() => { load(); }, [load]);

  const handleDelete = async (id) => {
    if (!confirm("¿Eliminar este video?")) return;
    try {
      await api.deleteVideo(id);
      setVideos((prev) => prev.filter((v) => v.id !== id));
    } catch (e) {
      alert(e?.message || "No se pudo eliminar el video.");
    }
  };

  const fileInputRef = React.useRef(null);
  const onUploadClick = () => {
    if (!getAccessToken()) {
      alert("Primero inicia sesión.");
      return;
    }
    fileInputRef.current?.click();
  };
  const onFileChange = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const title = file.name;
      const res = await api.uploadVideo({ title, file });
      await load();
      alert(res?.message || "Video subido. Procesamiento en curso.");
    } catch (err) {
      alert(err?.message || "Error al subir el video.");
    } finally {
      setUploading(false);
      e.target.value = "";
    }
  };

  const onSaveProfile = async (e) => {
    e.preventDefault();
    if (!me) return;
    setSaving(true);
    try {
      const updated = await api.updateMe({
        first_name: me.first_name,
        last_name: me.last_name,
        city: me.city,
        country: me.country,
      });
      setMe(updated);
    } catch (err) {
      alert(err?.message || "No se pudo actualizar el perfil.");
    } finally {
      setSaving(false);
    }
  };

  const onAvatarChange = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setAvatarUploading(true);
    try {
      const { avatar_url } = await api.uploadAvatar(file);
      setMe((prev) => ({ ...prev, avatar_url }));
    } catch (err) {
      alert(err?.message || "No se pudo subir el avatar.");
    } finally {
      setAvatarUploading(false);
      e.target.value = "";
    }
  };

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
        <div className="mb-4 d-flex flex-column flex-md-row align-items-md-center justify-content-between gap-3">
          <div>
            <h2 className="display-6 fw-bold">Player Dashboard</h2>
            <p className="text-secondary">Manage your profile and video submissions.</p>
          </div>
          <button className="btn btn-outline-light" onClick={() => window.location.assign("/galeria-votar")}>Ir a galería para votar</button>
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
                    backgroundSize: "cover",
                    backgroundPosition: "center",
                    backgroundImage: `url(${me?.avatar_url || "https://i.pravatar.cc/256"})`
                  }}
                />
                <p className="h5 fw-bold mb-0">{me ? `${me.first_name} ${me.last_name}` : (tokenPayload?.email || "Mi perfil")}</p>
                <p className="text-secondary small">ID: {userId || "—"}</p>
                <div className="mt-2">
                  <label className="btn btn-sm btn-outline-light mb-0">
                    {avatarUploading ? "Subiendo…" : "Cambiar foto"}
                    <input type="file" accept="image/*" className="d-none" onChange={onAvatarChange} disabled={avatarUploading} />
                  </label>
                </div>
              </div>

              <form className="mt-3" onSubmit={onSaveProfile}>
                <div className="mb-2">
                  <label className="form-label text-secondary">Nombre</label>
                  <input className="form-control bg-black text-white border-secondary-subtle"
                         value={me?.first_name || ""}
                         onChange={(e) => setMe((p) => ({ ...p, first_name: e.target.value }))}
                         required />
                </div>
                <div className="mb-2">
                  <label className="form-label text-secondary">Apellido</label>
                  <input className="form-control bg-black text-white border-secondary-subtle"
                         value={me?.last_name || ""}
                         onChange={(e) => setMe((p) => ({ ...p, last_name: e.target.value }))}
                         required />
                </div>
                <div className="row g-2">
                  <div className="col-6">
                    <label className="form-label text-secondary">Ciudad</label>
                    <input className="form-control bg-black text-white border-secondary-subtle"
                           value={me?.city || ""}
                           onChange={(e) => setMe((p) => ({ ...p, city: e.target.value }))}
                           required />
                  </div>
                  <div className="col-6">
                    <label className="form-label text-secondary">País</label>
                    <input className="form-control bg-black text-white border-secondary-subtle"
                           value={me?.country || ""}
                           onChange={(e) => setMe((p) => ({ ...p, country: e.target.value }))}
                           required />
                  </div>
                </div>
                <button className="btn btn-secondary w-100 mt-3" type="submit" disabled={saving}>
                  {saving ? "Guardando…" : "Guardar cambios"}
                </button>
              </form>
            </div>
          </div>

          {/* Col derecha - Videos */}
          <div className="col-12 col-lg-8">
            <div className="p-4 rounded-4 border border-secondary-subtle bg-black">
              <div className="d-flex flex-column flex-sm-row align-items-start align-items-sm-center justify-content-between gap-3">
                <h3 className="h4 fw-bold mb-0">My Videos</h3>
                <div className="d-flex align-items-center gap-2">
                  <button className="btn btn-success fw-bold rounded-pill d-flex align-items-center gap-2"
                          onClick={onUploadClick} disabled={uploading}>
                    <i className="bi bi-upload" />
                    {uploading ? "Uploading…" : "Upload Video"}
                  </button>
                  <input ref={fileInputRef} type="file" accept="video/*" className="d-none" onChange={onFileChange} />
                </div>
              </div>

              {loading && <p className="text-secondary mt-3">Cargando…</p>}
              {error && <div className="alert alert-danger mt-3">{error}</div>}

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
                              onClick={() => handleDelete(v.id)}
                              aria-label={`Eliminar ${v.title}`}
                            >
                              <i className="bi bi-trash fs-5" />
                            </button>
                          </td>
                        </tr>
                      ))}
                      {!loading && videos.length === 0 && (
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
