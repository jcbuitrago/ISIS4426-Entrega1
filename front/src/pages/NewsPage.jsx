import React from "react";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

export default function NewsPage() {
  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />
      <main className="container py-5">
        <h2 className="display-6 fw-bold">Noticias</h2>
        <p className="text-secondary">Próximamente: novedades y anuncios del ANB Rising Stars.</p>
      </main>
      <SiteFooter />
    </div>
  );
}
