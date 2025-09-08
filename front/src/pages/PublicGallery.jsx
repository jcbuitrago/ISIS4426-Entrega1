import React from "react";
import { useNavigate } from "react-router-dom";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

/**
 * PublicGallery
 * TODO:
 * - GET /api/videos/public?search=&page=
 * - Acción votar: POST /api/videos/:id/vote
 * - Modal player/overlay para reproducir
 */
export default function PublicGallery({
  onVote = (id) => {},     // TODO: conectar voto al backend
  onOpenVideo = (id) => {},// TODO: abrir modal reproductor
}) {
  const navigate = useNavigate();
  const primary = "#38e07b";

  const items = [
    { id: "1", name: "Carlos Mendoza", city: "Ciudad de México", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuDJWKfX_OM6cQKzC9PRwBYkapbiMSvbmkTEnngc1_HrYIw_qyAYusB1QA6Jbx8J3yjKco62tYtQEpH63bK977sbpnucWIAD6RpZ-YpJ2WwzUF8QbL_rnYYU4YsKV6JjdTJoOajGphY7ksUfKomlqMIEm6uyDwxfb-HiZFMAgQpnQCno6FeGSQZqJhOjGzlAMKnSkK57Yo4cFekFdJGzgrFLzAIDvO74MDekTgJTG-quExVjzw2UVj0Uviy_kG5XkohNH-LWh2-P4FLM" },
    { id: "2", name: "Sofía Ramirez", city: "Guadalajara", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuC24ks7L9jlVCD85qURQelUSrmEO8RDlCoXEKJeOgLylSqQevRZZp4q3HdKdqBcyuOnBvaDcXhfW6nKc4L6rLb5DowhS8C-nt-ZJZdcnO4B-UL4nFWDGOWY7L8b3Yg7nNxS43Kbv_DK7Dj7n0JGi8Dq6McLR8UqKtVnH-Sr9lI2vnZcLM_dETWkQKFgsSnTsxW3KOPgMooo-Lzj78D0k8Bbi0Lsgl7mQVdb_98f7Ewkv67GxqMLYLFk8V8CFzH1zxpOMJpT_zzSDy7m" },
    { id: "3", name: "Diego Torres", city: "Monterrey", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuA7PTwKae4eX4caggYhIL7thfkgxq_3aezWyxQk184xPZigJzg5V1j0bJcbuv8lzRTX5xoMkoNAQM6kJcq-gr44zxbyijSHsTtn8dxGjmolJwHGyF3eaqxf_bIexsDiL68GlRvCAmDMPbdIjLdN04bpwG-2dVDJa9AHOjpOX1ZuF5wZVVIpPgrMmE_857ZUsR95XaDNxsFK-4gTaAMRsCRmHoVysBnahLsHzi-Xqmsn1n7GNxfV-XWjdfRApM6diJdKwLflTXDNaeAn" },
    { id: "4", name: "Isabella Vargas", city: "Puebla", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuAdxfHOB5TDn1oeU9t-yZrxXH0ar02NjIQ30oi6bR3ICgsPwZlC7jL6klgbGBmbJLuCbsVdkHJyS2LwxQJogfnoMsqPk9TP-4WO-kho6dOqbcevHTHbpTD8OZ2sMDsFeDB5zNPM--fH4D2Bc1hkfIdMPx-4Vm-X2_zrnszPehP_3rZaSFAak_XlwTXcMJtcLngvfex7C-KrlDvXkHfvgybkAko-LcwnmsHEid70V-LMzWb4k9hQQj8mfvODNyYp480RBM7Q4jx5E8sP" },
    { id: "5", name: "Ricardo García", city: "Ciudad de México", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuC4UgWY8K5Mi4x5IBfqOBu_HAJlo2ZvWMPcKEejVWLPrWG1R0z67J3glHR3V5YdKcc8QZFKXNnFspTfnUAYYTWVlG9NllmLOE0VlEtbbjMSbGxRFFZB2fgOm9iA-ZvaEjWrBMt2kZWlvehVS8ziSgTpPbaIJ3VZpL1Lun0acbje5sp3MIHggOIMT6-gMAMtrffdSAcxSDowD8k6GKYaAJrcM71Fcas2rlLlWCEdjbnxIxPyV42TIiMAFdC-JL8k43Ppjxr7fxddrWr2" },
    { id: "6", name: "Valeria Castro", city: "Guadalajara", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuAP__ED_EfpbWBmsVOErizZ2HLVhyIY_42-JtoLkFEa8nAmTUFqK5vo4y3KIBldkPdcIh6dY2OeZlQtGLjjyV3weC50hAtUE8m2DAJyW1JQXZdHHIcyOxLDxX15zrZx1qCoZcb0RFLR0OaXeXY5heTOYPOJBr9tU8vIHs7cRYJGwJwSX8QWuE-MwRnTQfXYvjmHRE7B3mREzC48w-FkwxmsaL-D7QWYUO8GqDjHZeqyBHxoqe5SYUTwUgXch1_LlKVwpuMVj_0fYqRG" },
    { id: "7", name: "Alejandro Ruiz", city: "Monterrey", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuDhGLzmFmUXaLqJb42RxpdRIaYsjlvvupEOKP5sC0nrDN1tJ9YtdqyJ9m8bvuBpwJimNIM7YjVrHr9uXdYpImqu9LDVkiClBItADW5wbmkA8CuI7Y12jIY70UIqUGbc7xTlpHnz5G5Ux8TlgpNTRmxdNzkRwbEnsP3YMi3mglKy_uLa27B1O5r3f4ecLvtZnwTAyJLLHEMDCHQV_XMNQh5o3AX6XSsXFlGWx6YbBEnmo8Ev2uJtRW8BIAgMhgSWGAwy0-kqHvs73qyc" },
    { id: "8", name: "Camila Soto", city: "Puebla", thumb: "https://lh3.googleusercontent.com/aida-public/AB6AXuCtjOS4vvpgLaXom7lgwB7OCEoDDjkJ5EYtRyD65IjVPVxVccyhMlbNA_w_dWmgoiE4hg-9ca_0uFTG8CoqxV_HT_IUu1-19w9xBHE2tk5jwSmnv3hlQdq84edlFjKri0jUyMUJGJHv8bIKxIwS4OhiYdwDKzVUE7gyasTbxujEqdjOgnxQXGSYbCkML6E82JL-VPlTA0ecfJrN3_JnZvjQBW-XQtwifJ0CXqx_QzI3848gV_hvmwOgdOtKo2mL24uf773oBa9ypcT7" },
  ];

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />
      <main className="container py-5">
        <div className="text-center mb-5">
          <h2 className="display-5 fw-bold">Galería Pública de Videos</h2>
          <p className="text-secondary fs-5">
            Descubre a los futuros talentos del baloncesto. Vota por tus jugadores favoritos y ayúdalos a brillar.
          </p>
        </div>

        <div className="row g-4">
          {items.map((it) => (
            <div key={it.id} className="col-12 col-sm-6 col-lg-4 col-xl-3">
              <div className="card bg-black border-0 rounded-4 overflow-hidden h-100">
                <div
                  className="position-relative"
                  style={{ aspectRatio: "16/9", cursor: "pointer" }}
                  onClick={() => {
                    onOpenVideo(it.id);
                    // Ejemplo: navigate(`/galeria/${it.id}`);
                  }}
                >
                  <img src={it.thumb} alt={it.name} className="w-100 h-100 object-fit-cover" />
                  <div className="position-absolute top-0 start-0 w-100 h-100 d-flex align-items-center justify-content-center"
                       style={{ background: "rgba(0,0,0,0.4)", opacity: 0, transition: ".3s" }}
                       onMouseEnter={(e) => (e.currentTarget.style.opacity = 1)}
                       onMouseLeave={(e) => (e.currentTarget.style.opacity = 0)}>
                    <i className="bi bi-play-circle-fill display-4 text-white" />
                  </div>
                </div>

                <div className="card-body d-flex flex-column">
                  <div className="flex-grow-1">
                    <p className="h6 fw-semibold mb-0">{it.name}</p>
                    <p className="text-secondary small mb-0">{it.city}</p>
                  </div>
                  <button
                    className="btn fw-bold mt-3"
                    style={{ background: primary, color: "#000" }}
                    onClick={() => onVote(it.id)}
                  >
                    <i className="bi bi-hand-thumbs-up me-2" />
                    Votar
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
