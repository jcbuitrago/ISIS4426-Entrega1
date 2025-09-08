import React from "react";
import { useNavigate } from "react-router-dom";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

export default function RegisterPage({
  onSubmitRegister = async ({ fullName, email, password, accept }) => {
    // TODO: conectar al backend (ejemplo):
    // const res = await fetch("/api/auth/register", { method: "POST", body: JSON.stringify({ fullName, email, password, accept }) });
    // if (!res.ok) throw new Error("No se pudo crear la cuenta");
    console.log("register", { fullName, email, accept });
  },
}) {
  const navigate = useNavigate();
  const [fullName, setFullName] = React.useState("");
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [accept, setAccept] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    if (!accept) return setError("Debes aceptar los Términos y la Política de Privacidad.");
    if (password !== confirm) return setError("Las contraseñas no coinciden.");

    setLoading(true);
    try {
      await onSubmitRegister({ fullName, email, password, accept });
      navigate("/login"); // TODO: o navegar directo al dashboard si haces login automático
    } catch (err) {
      setError(err?.message || "Ocurrió un error al crear la cuenta.");
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
          <div className="mx-auto" style={{ maxWidth: 520 }}>
            <div className="text-center mb-4">
              <h2 className="fw-bold display-6">Crea tu cuenta</h2>
              <p className="text-secondary m-0">Únete al ANB Rising Stars Showcase.</p>
            </div>

            <div className="card bg-dark bg-opacity-75 border-secondary-subtle rounded-4 shadow">
              <div className="card-body p-4 p-md-5">
                {error && <div className="alert alert-danger" role="alert">{error}</div>}

                <form className="d-flex flex-column gap-3" onSubmit={handleSubmit}>
                  <div>
                    <label className="form-label text-secondary">Nombre completo</label>
                    <div className="input-group input-group-lg">
                      <span className="input-group-text bg-dark text-secondary border-secondary-subtle">
                        <i className="bi bi-person" />
                      </span>
                      <input
                        type="text"
                        className="form-control bg-dark text-white border-secondary-subtle"
                        placeholder="Tu nombre y apellido"
                        value={fullName}
                        onChange={(e) => setFullName(e.target.value)}
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="form-label text-secondary">Correo electrónico</label>
                    <div className="input-group input-group-lg">
                      <span className="input-group-text bg-dark text-secondary border-secondary-subtle">
                        <i className="bi bi-envelope" />
                      </span>
                      <input
                        type="email"
                        className="form-control bg-dark text-white border-secondary-subtle"
                        placeholder="nombre@dominio.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        required
                      />
                    </div>
                  </div>

                  <div className="row g-3">
                    <div className="col-12 col-md-6">
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
                    <div className="col-12 col-md-6">
                      <label className="form-label text-secondary">Confirmar contraseña</label>
                      <div className="input-group input-group-lg">
                        <span className="input-group-text bg-dark text-secondary border-secondary-subtle">
                          <i className="bi bi-shield-lock" />
                        </span>
                        <input
                          type="password"
                          className="form-control bg-dark text-white border-secondary-subtle"
                          placeholder="••••••••"
                          value={confirm}
                          onChange={(e) => setConfirm(e.target.value)}
                          required
                        />
                      </div>
                    </div>
                  </div>

                  <div className="form-check">
                    <input
                      className="form-check-input"
                      type="checkbox"
                      id="terms"
                      checked={accept}
                      onChange={(e) => setAccept(e.target.checked)}
                    />
                    <label className="form-check-label text-secondary" htmlFor="terms">
                      Acepto los <a href="#" className="link-warning text-decoration-none">Términos</a> y la{" "}
                      <a href="#" className="link-warning text-decoration-none">Política de Privacidad</a>.
                    </label>
                  </div>

                  <button className="btn btn-warning btn-lg fw-bold" type="submit" disabled={loading}>
                    {loading ? "Creando cuenta…" : "Crear cuenta"}
                  </button>

                  <p className="text-center text-secondary m-0">
                    ¿Ya tienes cuenta?{" "}
                    <button
                      type="button"
                      className="btn btn-link link-warning text-decoration-none p-0 fw-semibold"
                      onClick={() => navigate("/login")}
                    >
                      Inicia sesión
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
