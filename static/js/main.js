function initSortableGrid() {
  requestAnimationFrame(() => {
    const outer = document.getElementById("sortableGrid");
    const grid = outer?.querySelector(".sortable");
    if (!grid) {
      console.warn("❌ sortable inner grid not found");
      return;
    }

    // Destroy existing instance if needed
    if (grid._sortableInstance) {
      grid._sortableInstance.destroy();
    }

    console.log("✅ Initializing sortable on", grid);

    grid._sortableInstance = Sortable.create(grid, {
      animation: 150,
      handle: ".sortable-item",
      draggable: ".sortable-item",
      onEnd: function (evt) {
        const ids = [...grid.querySelectorAll(".sortable-item")].map((el) =>
          Number(el.dataset.id),
        );

        const payload = {
          order: ids,
        };

        const wrapper = grid.closest("#sortableGrid");
        const galleryID = wrapper?.dataset.gallery;
        const projectID = wrapper?.dataset.project;

        if (galleryID) payload.gallery_id = Number(galleryID);
        if (projectID) payload.project_id = Number(projectID);

        console.log("📦 Sending updated order:", payload);

        fetch("/admin/media/update-order-bulk", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "HX-Request": "true",
          },
          body: JSON.stringify(payload),
        });
      },
    });
  });
}

// 🧭 Sidebar Mobile Handling
// ==========================
function openSidebar() {
  const sidebar = document.getElementById("mobileSidebar");
  const backdrop = document.getElementById("mobileBackdrop");
  const sidebarPanel = sidebar?.querySelector(".fixed.inset-0.flex");

  sidebar?.classList.remove("hidden");
  setTimeout(() => {
    sidebarPanel?.classList.remove("-translate-x-full");
    sidebarPanel?.classList.add("translate-x-0");
    backdrop?.classList.remove("opacity-0", "pointer-events-none");
    backdrop?.classList.add("opacity-100");
  }, 10);
}

function closeSidebar() {
  const sidebar = document.getElementById("mobileSidebar");
  const backdrop = document.getElementById("mobileBackdrop");
  const sidebarPanel = sidebar?.querySelector(".fixed.inset-0.flex");

  sidebarPanel?.classList.remove("translate-x-0");
  sidebarPanel?.classList.add("-translate-x-full");
  backdrop?.classList.remove("opacity-100");
  backdrop?.classList.add("opacity-0", "pointer-events-none");

  setTimeout(() => {
    sidebar?.classList.add("hidden");
  }, 300);
}

document.addEventListener("DOMContentLoaded", () => {
  document
    .getElementById("mobileBackdrop")
    ?.addEventListener("click", closeSidebar);
});

// ==============================
// 🪄 Upload Modal + Tabs + Toast
// ==============================
window.closeModal = function () {
  const modal = document.querySelector("#mediaModalContainer .fixed");
  if (modal) {
    modal.classList.add("opacity-0");
    setTimeout(() => {
      const container = document.getElementById("mediaModalContainer");
      if (container) container.innerHTML = "";
    }, 200);
  }
};

document.addEventListener("click", function (event) {
  const modal = document.querySelector("#mediaModalContainer .fixed");
  const content = document.getElementById("mediaModalContent");
  if (!modal || !content || modal.classList.contains("hidden")) return;

  const clickedInside = content.contains(event.target);
  const clickedCloseBtn = event.target.closest("[onclick^='closeModal']");
  if (!clickedInside && !clickedCloseBtn) {
    window.closeModal();
  }
});

window.switchUploadTab = function (tab) {
  document
    .querySelectorAll(".upload-tab-section")
    .forEach((el) => el.classList.add("hidden"));
  document
    .querySelectorAll(".upload-tab-btn")
    .forEach((el) => el.classList.remove("active"));

  const target = document.getElementById(`upload-tab-${tab}`);
  if (target) {
    target.classList.remove("hidden");
    event.target.classList.add("active");
  }
};

document.addEventListener("htmx:afterOnLoad", function (evt) {
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger");
  if (trigger?.startsWith("media-attached-")) {
    const mediaID = trigger.replace("media-attached-", "");
    document.getElementById(`media-${mediaID}`)?.remove();
  }

  const isExistingTabActive = document
    .getElementById("upload-tab-existing")
    ?.classList.contains("block");

  if (!isExistingTabActive) return;

  const buttonsLeft = document.querySelectorAll("#upload-tab-existing button");
  if (buttonsLeft.length === 0) {
    window.closeModal();
  }
});

document.addEventListener("htmx:afterOnLoad", function (evt) {
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger-After-Settle");
  if (!trigger) return;

  const toastMap = {
    "show-toast": {
      variant: "success",
      heading: "Media Linked",
      subtitle: "Successfully attached to gallery or project",
    },
    "show-toast-unlinked": {
      variant: "warning",
      heading: "Media Unlinked",
      subtitle: "This media was removed from the gallery or project",
    },
  };

  const toast = toastMap[trigger];
  if (!toast) return;

  fetch(
    `/admin/toast?variant=${toast.variant}&heading=${encodeURIComponent(toast.heading)}&subtitle=${encodeURIComponent(toast.subtitle)}`,
  )
    .then((res) => res.text())
    .then((html) => {
      document.body.insertAdjacentHTML("beforeend", html);

      const el = document.querySelector(".toast-fade");
      if (!el) return;

      const fadeDuration = 700;
      const displayTime = parseInt(el.dataset.timeout) || 7000;

      setTimeout(() => {
        el.classList.remove("opacity-100");
        el.classList.add("opacity-0");
        setTimeout(() => el.remove(), fadeDuration);
      }, displayTime);
    });
});

document.addEventListener("htmx:afterSwap", function (event) {
  console.log("🔥 htmx:afterSwap fired for:", event.target.id);
  if (event.target.id === "sortableGrid") {
    console.log("🎯 Reinitializing sortable");
    initSortableGrid();
  }
});
