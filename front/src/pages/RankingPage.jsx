import React from "react";
import SiteHeader from "../components/SiteHeader.jsx";
import SiteFooter from "../components/SiteFooter.jsx";

/**
 * RankingPage
 * TODO:
 * - GET /api/ranking?limit=&page=
 * - Ordenar/filtrar por categoría/edad/posición
 */
export default function RankingPage() {
  const rows = [
    { rank: 1,  name: "Ethan Carter",      city: "Springfield", score: 1250 },
    { rank: 2,  name: "Olivia Bennett",    city: "Riverside",   score: 1180 },
    { rank: 3,  name: "Noah Thompson",     city: "Oakwood",     score: 1120 },
    { rank: 4,  name: "Sophia Davis",      city: "Greenfield",  score: 1050 },
    { rank: 5,  name: "Liam Wilson",       city: "Maplewood",   score: 980  },
    { rank: 6,  name: "Ava Martinez",      city: "Hillcrest",   score: 920  },
    { rank: 7,  name: "Jackson Anderson",  city: "Westwood",    score: 850  },
    { rank: 8,  name: "Isabella Taylor",   city: "Northwood",   score: 780  },
    { rank: 9,  name: "Lucas Thomas",      city: "Southwood",   score: 720  },
    { rank: 10, name: "Mia Jackson",       city: "Eastwood",    score: 650  },
  ];

  const highlightRank = (rank) => (rank === 1 ? "text-success fw-bold" : "text-white");

  return (
    <div className="bg-dark text-light min-vh-100 d-flex flex-column">
      <SiteHeader />

      <main className="container py-5">
        <div className="mb-4">
          <h1 className="display-6 fw-bold">Ranking de Jugadores</h1>
          <p className="text-secondary">Clasificación actualizada de los mejores talentos emergentes.</p>
        </div>

        <div className="border rounded-4 border-secondary-subtle bg-black">
          <div className="table-responsive">
            <table className="table table-dark table-hover align-middle mb-0">
              <thead>
                <tr className="text-secondary">
                  <th className="text-center" style={{ width: 64 }}>#</th>
                  <th>Jugador</th>
                  <th>Ciudad</th>
                  <th className="text-end">Puntuación</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => (
                  <tr key={r.rank} className="table-row">
                    <td className={`text-center fs-5 ${r.rank === 1 ? "text-success" : ""}`}>{r.rank}</td>
                    <td className="fw-semibold">{r.name}</td>
                    <td className="text-secondary">{r.city}</td>
                    <td className="text-end fw-semibold">{r.score.toLocaleString()}</td>
                  </tr>
                ))}
                {rows.length === 0 && (
                  <tr>
                    <td colSpan={4} className="text-center text-secondary py-4">No hay datos de ranking.</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </main>

      <SiteFooter />
    </div>
  );
}
