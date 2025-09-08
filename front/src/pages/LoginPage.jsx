import React from "react";
import { useNavigate } from "react-router-dom";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

export default function LoginPage({
  onSubmitLogin = async ({ username, password, remember }) => {
    // TODO: conectar al backend (ejemplo):
    // const res = await fetch("/api/auth/login", { method: "POST", body: JSON.stringify({ username, password, remember }) });
    // if (!res.ok) throw new Error("Credenciales inválidas");
    console.log("login", { username, remember });
  },
  onForgotPassword = () => {}, // TODO: navegación/flujo recuperar
}) {
  const navigate = useNavigate();
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [remember, setRemember] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await onSubmitLogin({ username, password, remember });
      navigate("/"); // TODO: redirigir a dashboard o donde quieras
    } catch (err) {
      setError(err?.message || "Ocurrió un error al iniciar sesión.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-black text-white min-vh-100 d-flex flex-column">
      <SiteHeader />

      <section className="position-relative flex-grow-1 d-flex align-items-center justify-content-center py-5">
        {/* Fondo */}
        <img
          alt="Basketball player"
          className="position-absolute top-0 start-0 w-100 h-100 object-fit-cover opacity-25"
          src="https://lh3.googleusercontent.com/aida-public/AB6AXuD0p1DJlAX_9Ezb0TGO2IklZWfB12-z9NDVysXk_66Arj_Fp5Eg1wCqdRUJDFnlv9v3a9e2jZ9C9fS6sz89IHW9pVhKHTKNbt2EPsHjFqzIVsXY1XEoxvPFFf9Tl5J8K0oV38SGdKI8WWO9fFarHNdKTaLmxRqY7uc_R6Hg7prjBRRIkjjT-SIQHBnlPiWGRI6Xrd-t46Tk14BP4UcJRg8WNz9LMxQJERV_8PZhB3CQPdmos6PfXe_hhBEwZTwgyiA4rdMpg-Uox_Vp"
        />
        <div className="position-absolute top-0 start-0 w-100 h-100" style={{ backdropFilter: "blur(2px)" }} />

        <div className="position-relative container" style={{ zIndex: 1 }}>
          <div className="mx-auto" style={{ maxWidth: 448 }}>
            <div className="text-center mb-4">
              <h2 className="fw-bold display-6">Accede a tu cuenta</h2>
              <p className="text-secondary m-0">Descubre a las próximas estrellas del baloncesto.</p>
            </div>

            <div className="card bg-dark bg-opacity-75 border-secondary-subtle rounded-4 shadow">
              <div className="card-body p-4 p-md-5">
                {error && <div className="alert alert-danger" role="alert">{error}</div>}

                <form className="d-flex flex-column gap-3" onSubmit={handleSubmit}>
                  <div>
                    <label className="form-label text-secondary">Nombre de usuario o correo</label>
                    <div className="input-group input-group-lg">
                      <span className="input-group-text bg-dark text-secondary border-secondary-subtle">
                        <i className="bi bi-person" />
                      </span>
                      <input
                        type="text"
                        className="form-control bg-dark text-white border-secondary-subtle"
                        placeholder="usuario@correo.com"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="form-label text-secondary">Contraseña</label>
                    <div className="input-group input-group-lg">
                      <span className="input-group-text bg-dark text-secondary border-secondary-subtle">
                        <i className="bi bi-lock" />
                      </span>
                      <input
                        type="password"
                        className="form-control bg-dark text-white border-secondary-subtle"
                        placeholder="••••••••"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                      />
                    </div>
                  </div>

                  <button className="btn btn-warning btn-lg fw-bold" type="submit" disabled={loading}>
                    {loading ? "Ingresando…" : "Ingresar"}
                  </button>

                  <div className="d-flex align-items-center justify-content-between">
                    <div className="form-check">
                      <input
                        className="form-check-input"
                        type="checkbox"
                        id="remember-me"
                        checked={remember}
                        onChange={(e) => setRemember(e.target.checked)}
                      />
                      <label className="form-check-label text-secondary" htmlFor="remember-me">Recuérdame</label>
                    </div>
                    <button type="button" className="btn btn-link link-warning text-decoration-none p-0" onClick={onForgotPassword}>
                      ¿Olvidaste tu contraseña?
                    </button>
                  </div>

                  <p className="text-center text-secondary m-0">
                    ¿Eres nuevo aquí?{" "}
                    <button
                      type="button"
                      className="btn btn-link link-warning text-decoration-none p-0 fw-semibold"
                      onClick={() => navigate("/register")}
                    >
                      Crea una cuenta
                    </button>
                  </p>
                </form>
              </div>
            </div>
          </div>
        </div>
      </section>

      <SiteFooter />
    </div>
  );
}
