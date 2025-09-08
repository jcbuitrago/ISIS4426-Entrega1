import React from "react";
import { NavLink, useNavigate } from "react-router-dom";

export default function SiteHeader({ onRegister = () => {}, onLogin = () => {} }) {
  const navigate = useNavigate();

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
            <NavLink className="nav-link text-secondary" to="/showcase">Showcase</NavLink>
          </li>
          <li className="nav-item">
            <NavLink className="nav-link text-secondary" to="/talentos">Talentos</NavLink>
          </li>
          <li className="nav-item">
            <NavLink className="nav-link text-secondary" to="/noticias">Noticias</NavLink>
          </li>
        </ul>

        <div className="d-flex align-items-center gap-2">
          <button
            className="btn btn-warning fw-bold"
            onClick={() => {
              onRegister();          // TODO: lógica/telemetría opcional
              navigate("/register"); // navegación
            }}
          >
            Regístrate
          </button>
          <button
            className="btn btn-outline-light fw-bold"
            onClick={() => {
              onLogin();             // TODO: lógica/telemetría opcional
              navigate("/login");    // navegación
            }}
          >
            Acceder
          </button>
        </div>
      </nav>
    </header>
  );
}
