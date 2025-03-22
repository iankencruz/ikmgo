document.addEventListener("DOMContentLoaded", function () {
  const modal = document.getElementById("lightboxModal");
  const modalImg = document.getElementById("lightboxImg");
  const galleryImages = Array.from(document.querySelectorAll("#gallery img"));

  if (!modal || !modalImg || galleryImages.length === 0) {
    console.error("❌ Lightbox elements missing or no images found.");
    return;
  }

  let currentIndex = 0;
  let currentGalleryID = null;

  function openLightbox(index) {
    currentIndex = index;
    modal.classList.remove("hidden");

    const fullResUrl = galleryImages[currentIndex].getAttribute("data-full");
    currentGalleryID = galleryImages[currentIndex].getAttribute("data-gallery");

    if (!fullResUrl || !currentGalleryID) {
      console.error("❌ Missing data attributes for image index:", index);
      return;
    }

    modalImg.src = fullResUrl;
    console.log(
      `✅ Opening lightbox for gallery ${currentGalleryID}, image ${fullResUrl}`,
    );
  }

  function closeLightbox() {
    modal.classList.add("hidden");
    console.log("✅ Lightbox closed.");
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

  console.log("✅ Lightbox initialized.");
});
