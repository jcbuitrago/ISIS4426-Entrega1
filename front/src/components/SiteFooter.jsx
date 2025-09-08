import React from "react";

export default function SiteFooter() {
  return (
    <footer className="bg-black border-top border-dark py-4 text-white-50">
      <div className="container d-flex flex-column flex-md-row justify-content-between align-items-center">
        <div className="d-flex align-items-center gap-2 mb-3 mb-md-0">
          <i className="bi bi-basket2-fill fs-5 text-warning" aria-hidden />
          <h2 className="h6 m-0 text-white">ANB Rising Stars</h2>
        </div>
        <p className="m-0 small">© 2024 Asociación Nacional de Baloncesto. Todos los derechos reservados.</p>
        <div className="d-flex gap-3 mt-3 mt-md-0">
          <button className="btn btn-link link-light link-opacity-75-hover p-0">Términos</button>
          <button className="btn btn-link link-light link-opacity-75-hover p-0">Privacidad</button>
        </div>
      </div>
    </footer>
  );
}


