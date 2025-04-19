function initLightbox() {
  const modal = document.getElementById("lightboxModal");
  const modalImg = document.getElementById("lightboxImg");
  const galleryImages = Array.from(document.querySelectorAll("img[data-full]"));

  if (!modal || !modalImg || galleryImages.length === 0) {
    console.warn("⚠️ Lightbox: missing modal or no images to bind.");
    return;
  }

  let currentIndex = 0;

  function openLightbox(index) {
    currentIndex = index;
    const img = galleryImages[currentIndex];

    const fullResUrl = img.getAttribute("data-full");
    modalImg.src = fullResUrl;
    modal.classList.remove("hidden");

    //console.log(`✅ Opening lightbox: ${fullResUrl}`);
  }

  function closeLightbox() {
    modal.classList.add("hidden");
    modalImg.src = "";
    //console.log("✅ Lightbox closed.");
  }

  function showNext() {
    currentIndex = (currentIndex + 1) % galleryImages.length;
    openLightbox(currentIndex);
  }

  function showPrev() {
    currentIndex =
      (currentIndex - 1 + galleryImages.length) % galleryImages.length;
    openLightbox(currentIndex);
  }

  galleryImages.forEach((img, i) => {
    img.addEventListener("click", () => openLightbox(i));
  });

  document
    .getElementById("lightboxClose")
    ?.addEventListener("click", closeLightbox);
  document.getElementById("lightboxNext")?.addEventListener("click", showNext);
  document.getElementById("lightboxPrev")?.addEventListener("click", showPrev);

  modal.addEventListener("click", (e) => {
    if (e.target === modal) closeLightbox();
  });

  //console.log("✅ Lightbox bindings complete");
}

// 🧠 Run on initial page load
document.addEventListener("DOMContentLoaded", initLightbox);

// 🔁 Run after any HTMX swap
document.body.addEventListener("htmx:afterSwap", () => {
  initLightbox();
});
