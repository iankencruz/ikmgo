// 🛡️ Global Sortable safeguard
if (window.Sortable && typeof Sortable.create === "function") {
  const originalCreate = Sortable.create;

  Sortable.create = function (el, options) {
    if (!el || !(el instanceof HTMLElement)) {
      console.warn("⛔ Sortable.create called with invalid element:", el);
      console.trace(); // shows where it was triggered
      return null;
    }
    return originalCreate.call(Sortable, el, options);
  };
}

// ========================
// 🔄 Sortable Grid Support
// ========================

function initSortableGrid() {
  requestAnimationFrame(() => {
    const outer = document.getElementById("sortableGrid");

    if (!outer) {
      //console.warn("⛔ #sortableGrid not found — Sortable init skipped");
      return;
    }

    const grid = outer.querySelector(".sortable");

    if (!grid || !(grid instanceof HTMLElement)) {
      //console.warn("⛔ .sortable not found or not an HTMLElement");
      return;
    }

    // Prevent duplicate init
    if (grid._sortableInstance) {
      grid._sortableInstance.destroy();
    }

    grid._sortableInstance = Sortable.create(grid, {
      animation: 150,
      handle: ".sortable-item",
      draggable: ".sortable-item",

      onEnd: function (evt) {
        const ids = [...grid.querySelectorAll(".sortable-item")].map((el) =>
          Number(el.dataset.id),
        );

        const galleryID = outer.dataset.gallery;
        const projectID = outer.dataset.project;

        const payload = { order: ids };
        if (galleryID) payload.gallery_id = Number(galleryID);
        if (projectID) payload.project_id = Number(projectID);

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

    //console.log("✅ Sortable initialized");
  });
}

// 🔁 Init Sortable after full load + HTMX swaps
document.addEventListener("DOMContentLoaded", initSortableGrid);

document.addEventListener("htmx:afterSwap", () => {
  setTimeout(() => {
    if (document.getElementById("sortableGrid")) {
      initSortableGrid();
    }
  }, 50);
});

// ===========================
// 🧭 Sidebar Mobile Handling
// ===========================
function openSidebar() {
  const sidebar = document.getElementById("mobileSidebar");
  const backdrop = document.getElementById("mobileBackdrop");
  const panel = sidebar?.querySelector(".fixed.inset-0.flex");

  sidebar?.classList.remove("hidden");
  setTimeout(() => {
    panel?.classList.remove("-translate-x-full");
    panel?.classList.add("translate-x-0");
    backdrop?.classList.remove("opacity-0", "pointer-events-none");
    backdrop?.classList.add("opacity-100");
  }, 10);
}

function closeSidebar() {
  const sidebar = document.getElementById("mobileSidebar");
  const backdrop = document.getElementById("mobileBackdrop");
  const panel = sidebar?.querySelector(".fixed.inset-0.flex");

  panel?.classList.remove("translate-x-0");
  panel?.classList.add("-translate-x-full");
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

// 🧼 Clean up modal after linking media
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

// 🍞 Toast trigger handler
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
    `/admin/toast?variant=${toast.variant}&heading=${encodeURIComponent(
      toast.heading,
    )}&subtitle=${encodeURIComponent(toast.subtitle)}`,
  )
    .then((res) => res.text())
    .then((html) => {
      const container = document.getElementById("toastContainer");
      if (!container) return;

      // 💥 Remove any existing toast before adding a new one
      container.innerHTML = "";

      container.insertAdjacentHTML("beforeend", html);

      const el = container.querySelector(".toast-fade");
      if (!el) return;

      // ✨ Animate in: fade + slide + bounce
      setTimeout(() => {
        el.classList.add("opacity-100", "translate-y-0", "scale-100");
      }, 50); // short delay to allow paint

      const fadeDuration = 700;
      const displayTime = parseInt(el.dataset.timeout) || 7000;

      // Fade out after display time
      setTimeout(() => {
        el.classList.remove("opacity-100");
        el.classList.add("opacity-0");
        setTimeout(() => el.remove(), fadeDuration);
      }, displayTime);
    });
});

// =====================
// 🗂️ Tab Switching
// =====================
function switchToTab(tabName, event = null) {
  if (event) event.preventDefault();

  document
    .querySelectorAll(".tab-pane")
    .forEach((pane) => pane.classList.add("hidden"));

  document.getElementById(tabName)?.classList.remove("hidden");

  document.querySelectorAll(".tab-link").forEach((link) => {
    link.classList.remove("border-indigo-500", "text-indigo-600");
    link.classList.add("border-transparent", "text-gray-500");
  });

  document.querySelectorAll(`.tab-link`).forEach((link) => {
    if (link.textContent.trim() === getTabLabel(tabName)) {
      link.classList.add("border-indigo-500", "text-indigo-600");
      link.classList.remove("border-transparent", "text-gray-500");
    }
  });
}

function getTabLabel(tabValue) {
  const map = { info: "Gallery Info", cover: "Cover Image" };
  return map[tabValue] || tabValue;
}

function toggleUserMenu(scope) {
  const menu = document.getElementById(`userMenu${scope}`);
  const button = document.getElementById(`userMenuButton${scope}`);

  if (!menu || !button) {
    console.warn(`⚠️ Menu or button not found for scope: ${scope}`);
    return;
  }

  menu.classList.toggle("hidden");

  // Close if clicked outside
  document.addEventListener("click", function onDocClick(e) {
    const clickedInsideMenu = menu.contains(e.target);
    const clickedButton = button.contains(e.target);

    if (!clickedInsideMenu && !clickedButton) {
      menu.classList.add("hidden");
      document.removeEventListener("click", onDocClick);
    }
  });
}

document.addEventListener("DOMContentLoaded", function () {
  const titleInput = document.getElementById("title");
  const slugInput = document.getElementById("slug");

  if (titleInput && slugInput) {
    titleInput.addEventListener("input", () => {
      const slug = titleInput.value
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9 -]/g, "")
        .replace(/\s+/g, "-")
        .replace(/-+/g, "-");

      slugInput.value = slug;
    });
  }
});

function openAboutImageModal() {
  const modal = document.getElementById("aboutImageModal");
  if (!modal) return;

  modal.classList.remove("hidden");

  // trigger HTMX fetch if not already loaded
  const content = modal.querySelector(".modal-content");
  if (!content.dataset.loaded) {
    htmx.ajax("GET", "/admin/settings/select-about-image", { target: content });
    content.dataset.loaded = "true";
  }
}
