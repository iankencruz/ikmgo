document.addEventListener("DOMContentLoaded", function () {
  const modal = document.getElementById("lightboxModal");
  const modalImg = document.getElementById("lightboxImg");
  const galleryImages = Array.from(document.querySelectorAll("#gallery img"));

  // ✅ Only run the event logic if these elements exist
  if (!modal || !modalImg || galleryImages.length === 0) {
    return; // Exit if required elements are missing
  }

  let currentIndex = 0;

  function openLightbox(index) {
    currentIndex = index;
    modal.classList.remove("hidden");
    modalImg.src = galleryImages[currentIndex].getAttribute("data-large");
  }

  function closeLightbox() {
    modal.classList.add("hidden");
  }

  function showNext() {
    currentIndex = (currentIndex + 1) % galleryImages.length;
    modalImg.src = galleryImages[currentIndex].getAttribute("data-large");
  }

  function showPrev() {
    currentIndex =
      (currentIndex - 1 + galleryImages.length) % galleryImages.length;
    modalImg.src = galleryImages[currentIndex].getAttribute("data-large");
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
});
