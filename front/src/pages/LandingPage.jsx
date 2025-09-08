import React from "react";
import { useNavigate } from "react-router-dom";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

export default function LandingPage({
  onJoin = () => {},          // TODO: conectar a registro/back-end
  onViewTalents = () => {},   // TODO: navegar a /talentos o cargar lista
}) {
  const navigate = useNavigate();

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />

      {/* Hero */}
      <section
        className="position-relative d-flex align-items-center"
        style={{
          minHeight: "calc(100vh - 81px)",
          backgroundImage:
            'linear-gradient(90deg, rgba(17,24,39,0.95) 0%, rgba(17,24,39,0.6) 50%, rgba(17,24,39,0.2) 100%), url(https://lh3.googleusercontent.com/aida-public/AB6AXuDRnwkLk-XKSTcsWt5djEMpKJl0tRmvmdO8bAvgZ8bu79cRFx2PLLRt7aWV3sKlfOA0sZmZXeOcN0qtrUVUY2Q7QlE61fQFIFofRCVAK5umB8CT3Y_R49CAg70gGPieJuZFnVy3MJm1BZ4kVjcf0Qx5bgZU5AB0FuM2DJ6G1D7KiDmYuPY65K78shG4V3bTLPDv-mHJiIOnZgHrOu5pQk2mbNYGUitm44f2QNR03g7oP3ta5go4CPQhhCs6pxCiRIoLwPy9oJGuj3yo)',
          backgroundSize: "cover",
          backgroundPosition: "center",
        }}
      >
        <div className="container py-5">
          <div className="col-12 col-lg-7">
            <h1 className="display-4 fw-bold lh-1 mb-3">
              El Futuro del Baloncesto Empieza Aquí.
            </h1>
            <p className="fs-5 text-light opacity-75 mb-4">
              Descubre, compite y brilla en el ANB Rising Stars Showcase. La plataforma para los talentos del mañana.
            </p>
            <div className="d-flex gap-3">
              <button
                className="btn btn-warning btn-lg fw-bold"
                onClick={() => {
                  onJoin();          // placeholder lógica
                  navigate("/register");
                }}
              >
                Únete Ahora
              </button>
              <button
                className="btn btn-primary btn-lg fw-bold"
                onClick={() => {
                  onViewTalents();   // placeholder lógica
                  navigate("/talentos"); // si aún no existe, deja el navigate para futuro
                }}
              >
                Ver Talentos
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* Why Section */}
      <section className="bg-secondary-subtle text-white py-5">
        <div className="container">
          <h2 className="display-6 fw-bold text-center mb-5">¿Por qué ANB Rising Stars?</h2>
          <div className="row g-4 text-center">
            <div className="col-12 col-md-4">
              <div className="d-flex flex-column align-items-center h-100 p-3">
                <div
                  className="bg-warning rounded-circle d-inline-flex align-items-center justify-content-center mb-3"
                  style={{ width: 72, height: 72 }}
                >
                  <i className="bi bi-eye fs-3 text-white" />
                </div>
                <h3 className="h4 fw-bold mb-2">Visibilidad Nacional</h3>
                <p className="text-light opacity-75 m-0">Destaca ante reclutadores y equipos de todo el país.</p>
              </div>
            </div>
            <div className="col-12 col-md-4">
              <div className="d-flex flex-column align-items-center h-100 p-3">
                <div
                  className="bg-warning rounded-circle d-inline-flex align-items-center justify-content-center mb-3"
                  style={{ width: 72, height: 72 }}
                >
                  <i className="bi bi-trophy fs-3 text-white" />
                </div>
                <h3 className="h4 fw-bold mb-2">Competencia de Élite</h3>
                <p className="text-light opacity-75 m-0">Mide tu nivel contra los mejores prospectos de tu categoría.</p>
              </div>
            </div>
            <div className="col-12 col-md-4">
              <div className="d-flex flex-column align-items-center h-100 p-3">
                <div
                  className="bg-warning rounded-circle d-inline-flex align-items-center justify-content-center mb-3"
                  style={{ width: 72, height: 72 }}
                >
                  <i className="bi bi-mortarboard fs-3 text-white" />
                </div>
                <h3 className="h4 fw-bold mb-2">Desarrollo Profesional</h3>
                <p className="text-light opacity-75 m-0">Accede a recursos y entrenamientos para potenciar tu carrera.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <SiteFooter />
    </div>
  );
}
