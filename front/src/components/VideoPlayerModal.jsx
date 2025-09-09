import React from "react";

export default function VideoPlayerModal({ url, title, open, onClose }) {
  if (!open) return null;
  return (
    <div className="position-fixed top-0 start-0 w-100 h-100" style={{ zIndex: 1050, background: "rgba(0,0,0,0.8)" }}>
      <div className="container h-100 d-flex align-items-center justify-content-center">
        <div className="bg-black rounded-4 border border-secondary-subtle w-100" style={{ maxWidth: 960 }}>
          <div className="d-flex align-items-center justify-content-between p-3 border-bottom border-secondary-subtle">
            <h5 className="m-0">{title || "Reproductor"}</h5>
            <button className="btn btn-sm btn-outline-light" onClick={onClose}><i className="bi bi-x-lg" /></button>
          </div>
          <div className="p-2">
            <video src={url} controls playsInline className="w-100" style={{ borderRadius: 12, maxHeight: '70vh' }} />
          </div>
        </div>
      </div>
    </div>
  );
}
