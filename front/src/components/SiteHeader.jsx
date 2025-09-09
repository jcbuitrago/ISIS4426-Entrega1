import React from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { api, getAccessToken, decodeAccessToken } from "../api.js";

export default function SiteHeader({ onRegister = () => {}, onLogin = () => {} }) {
  const navigate = useNavigate();
  const [user, setUser] = React.useState(null);

  React.useEffect(() => {
    const tok = getAccessToken();
    if (!tok) { setUser(null); return; }
    (async () => {
      try { const me = await api.getMe(); setUser(me); } catch { setUser(null); }
    })();
  }, []);

  const logout = () => {
    localStorage.removeItem("access_token");
    localStorage.removeItem("access_token_expires_at");
    localStorage.removeItem("remember_me");
    setUser(null);
    navigate("/");
  };

  return (
    <header className="border-bottom border-dark bg-black py-3">
      <nav className="container d-flex align-items-center justify-content-between">
        <button
          className="btn btn-link p-0 d-flex align-items-center text-decoration-none"
          onClick={() => navigate("/")}
        >
          <i className="bi bi-basket2-fill fs-3 text-warning" aria-hidden />
          <span className="ms-2 h5 m-0 fw-bold text-white">ANB Rising Stars</span>
        </button>

        <ul className="d-none d-md-flex nav gap-2">
          <li className="nav-item">
            <NavLink className="nav-link text-secondary" to="/">Inicio</NavLink>
          </li>
          <li className="nav-item">
            <NavLink className="nav-link text-secondary" to="/galeria">Showcase</NavLink>
          </li>
          <li className="nav-item">
            <NavLink className="nav-link text-secondary" to="/ranking">Ranking</NavLink>
          </li>
          <li className="nav-item">
            <NavLink className="nav-link text-secondary" to="/noticias">Noticias</NavLink>
          </li>
        </ul>

        <div className="d-flex align-items-center gap-2">
          {user ? (
            <div className="dropdown">
              <button className="btn btn-outline-light d-flex align-items-center gap-2 dropdown-toggle" data-bs-toggle="dropdown">
                <span
                  className="rounded-circle"
                  style={{ width: 32, height: 32, backgroundImage: `url(${user.avatar_url || "https://i.pravatar.cc/100"})`, backgroundSize: "cover", backgroundPosition: "center" }}
                />
                <span className="d-none d-sm-inline">{user.first_name}</span>
              </button>
              <ul className="dropdown-menu dropdown-menu-end">
                <li><button className="dropdown-item" onClick={() => navigate("/dashboard")}>Mi perfil</button></li>
                <li><button className="dropdown-item" onClick={() => navigate("/galeria-votar")}>Galería para votar</button></li>
                <li><hr className="dropdown-divider"/></li>
                <li><button className="dropdown-item" onClick={logout}>Cerrar sesión</button></li>
              </ul>
            </div>
          ) : (
            <>
              <button
                className="btn btn-warning fw-bold"
                onClick={() => { onRegister(); navigate("/register"); }}
              >
                Regístrate
              </button>
              <button
                className="btn btn-outline-light fw-bold"
                onClick={() => { onLogin(); navigate("/login"); }}
              >
                Acceder
              </button>
            </>
          )}
        </div>
      </nav>
    </header>
  );
}
