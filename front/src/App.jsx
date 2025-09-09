import React from "react";
import { Routes, Route } from "react-router-dom";
import LandingPage from "./pages/LandingPage.jsx";
import LoginPage from "./pages/LoginPage.jsx";
import RegisterPage from "./pages/RegisterPage.jsx";

import PlayerDashboard from "./pages/PlayerDashboard.jsx";
import PublicGallery from "./pages/PublicGallery.jsx";
import RankingPage from "./pages/RankingPage.jsx";
import NewsPage from "./pages/NewsPage.jsx";
import VotingGallery from "./pages/VotingGallery.jsx";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      <Route path="/dashboard" element={<PlayerDashboard />} />
      <Route path="/galeria" element={<PublicGallery />} />
      <Route path="/galeria-votar" element={<VotingGallery />} />
      <Route path="/ranking" element={<RankingPage />} />
      <Route path="/noticias" element={<NewsPage />} />
    </Routes>
  );
}
