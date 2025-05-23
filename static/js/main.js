document.addEventListener("DOMContentLoaded", () => {
  const defaultTab = document
    .getElementById("upload-tab-upload")
    ?.classList.contains("block")
    ? "upload"
    : "existing";
  switchUploadTab(defaultTab);
});
// 🛡️ Global Sortable safeguard
if (window.Sortable && typeof Sortable.create === "function") {
  const originalCreate = Sortable.create;

  Sortable.create = function(el, options) {
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

      onEnd: function(evt) {
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
window.closeModal = function() {
  const modal = document.querySelector("#mediaModalContainer .fixed");
  if (modal) {
    modal.classList.add("opacity-0");
    setTimeout(() => {
      const container = document.getElementById("mediaModalContainer");
      if (container) container.innerHTML = "";
    }, 200);
  }
};

document.addEventListener("click", function(event) {
  const modal = document.querySelector("#mediaModalContainer .fixed");
  const content = document.getElementById("mediaModalContent");
  if (!modal || !content || modal.classList.contains("hidden")) return;

  const clickedInside = content.contains(event.target);
  const clickedCloseBtn = event.target.closest("[onclick^='closeModal']");
  if (!clickedInside && !clickedCloseBtn) {
    window.closeModal();
  }
});

window.switchUploadTab = function(tab, event = null) {
  if (event) event.preventDefault();

  // Hide all tab content sections
  document
    .querySelectorAll(".upload-tab-section")
    .forEach((el) => el.classList.add("hidden"));

  // Show the selected tab
  document.getElementById(`upload-tab-${tab}`)?.classList.remove("hidden");

  // Reset tab button styles
  document.querySelectorAll(".upload-tab-btn").forEach((btn) => {
    btn.classList.remove("border-amber-500", "text-amber-600", "border-b-2");
    btn.classList.add("border-transparent", "text-gray-500");

    if (btn.dataset.tab === tab) {
      btn.classList.add("border-amber-500", "text-amber-600", "border-b-2");
      btn.classList.remove("border-transparent", "text-gray-500");
    }
  });
};

// 🧼 Clean up modal after linking media
document.addEventListener("htmx:afterOnLoad", function(evt) {
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

document.addEventListener("htmx:afterOnLoad", function(evt) {
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger-After-Settle");
  if (!trigger) return;

  const toastMap = {
    "show-toast": {
      variant: "success",
      heading: "Media Linked",
      subtitle: "Successfully attached to gallery or project",
      path: "/admin/toast", // admin context
    },
    "show-toast-unlinked": {
      variant: "warning",
      heading: "Media Unlinked",
      subtitle: "This media was removed from the gallery or project",
      path: "/admin/toast",
    },
    "show-toast-contact": {
      variant: "success",
      heading: "Message Sent!",
      subtitle: "Your message has been submitted successfully.",
      path: "/toast", // public context
    },
  };

  const toast = toastMap[trigger];
  if (!toast) return;

  fetch(
    `${toast.path}?variant=${toast.variant}&heading=${encodeURIComponent(
      toast.heading
    )}&subtitle=${encodeURIComponent(toast.subtitle)}`
  )
    .then((res) => res.text())
    .then((html) => {
      const container = document.getElementById("toastContainer");
      if (!container) return;

      container.innerHTML = ""; // remove any existing toasts
      container.insertAdjacentHTML("beforeend", html);

      const el = container.querySelector(".toast-fade");
      if (!el) return;

      // Animate in
      setTimeout(() => {
        el.classList.add("opacity-100", "translate-y-0", "scale-100");
      }, 50);

      const fadeDuration = 700;
      const displayTime = parseInt(el.dataset.timeout) || 7000;

      // Auto-dismiss
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

  // Hide all tab content panels
  document.querySelectorAll(".tab-pane").forEach((pane) => {
    pane.classList.add("hidden");
  });

  // Show the selected tab pane
  document.getElementById(tabName)?.classList.remove("hidden");

  // Reset all tab links
  document.querySelectorAll(".tab-link").forEach((link) => {
    link.classList.remove("border-indigo-500", "text-indigo-600");
    link.classList.add("border-transparent", "text-gray-500");
  });

  // Activate the current tab link
  document.querySelectorAll(".tab-link").forEach((link) => {
    if (link.textContent.trim().toLowerCase() === tabName.toLowerCase()) {
      link.classList.remove("border-transparent", "text-gray-500");
      link.classList.add("border-indigo-500", "text-indigo-600");
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

document.addEventListener("DOMContentLoaded", function() {
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

function handleFileSelect(event) {
  const input = event.target;
  const previewGrid = document.getElementById("fileList");

  if (!input.files || input.files.length === 0) {
    previewGrid.classList.add("hidden");
    previewGrid.innerHTML = "";
    return;
  }

  previewGrid.classList.remove("hidden");
  previewGrid.innerHTML = "";

  Array.from(input.files).forEach((file) => {
    const container = document.createElement("div");
    container.className = "border rounded shadow-sm p-2 text-center";

    const name = document.createElement("p");
    name.className = "text-xs truncate mt-2 text-gray-700";
    name.textContent = file.name;

    if (file.type.startsWith("image/")) {
      const img = document.createElement("img");
      img.className = "w-full h-32 object-cover rounded";
      img.alt = file.name;

      const reader = new FileReader();
      reader.onload = (e) => {
        img.src = e.target.result;
      };
      reader.readAsDataURL(file);

      container.appendChild(img);
    } else {
      const icon = document.createElement("div");
      icon.textContent = "📁";
      icon.className = "text-4xl text-gray-400";
      container.appendChild(icon);
    }

    container.appendChild(name);
    previewGrid.appendChild(container);
  });
}

// Optional: Add visual border highlight on drag
document.addEventListener("DOMContentLoaded", () => {
  const dropzone = document.getElementById("dropzone");

  if (!dropzone) return;

  ["dragenter", "dragover"].forEach((evt) => {
    dropzone.addEventListener(evt, () => {
      dropzone.classList.add("border-amber-500", "bg-amber-50");
    });
  });

  ["dragleave", "drop"].forEach((evt) => {
    dropzone.addEventListener(evt, () => {
      dropzone.classList.remove("border-amber-500", "bg-amber-50");
    });
  });
});

// =====================
// 📤 Upload Controllers

let selectedFiles = [];
let uploadControllers = {};

function previewFiles(event) {
  const files = event.target.files;
  const fileList = document.getElementById("fileList");
  fileList.innerHTML = "";

  selectedFiles = {};

  if (files.length > 0) fileList.classList.remove("hidden");

  for (const file of files) {
    const fileId = crypto.randomUUID();

    selectedFiles[fileId] = file;

    const reader = new FileReader();
    reader.onload = function(e) {
      const block = document.createElement("div");
      block.id = `file-${fileId}`;
      block.className = "relative border rounded-lg p-2 shadow-sm bg-white";

      block.innerHTML = `
        <div class="flex flex-col gap-2">
          <div class="relative aspect-square overflow-hidden rounded border w-24 h-24">
            <img src="${e.target.result}" class="object-cover w-full h-full" />
            <button onclick="cancelUpload('${fileId}')" class="absolute top-1 right-1 text-sm text-white bg-red-500 rounded-full px-1">&times;</button>
          </div>
          <p class="text-xs text-gray-700 truncate">${file.name}</p>
          <p id="status-${fileId}" class="text-xs text-gray-500">Ready to upload</p>
          <div id="progress-container-${fileId}" class="hidden">
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div id="progress-${fileId}" class="bg-amber-500 h-2 rounded-full transition-all duration-300" style="width: 0%;"></div>
            </div>
          </div>
        </div>
      `;

      fileList.appendChild(block);
    };
    reader.readAsDataURL(file);
  }
}

function startUpload() {
  const projectId = document.getElementById("upload-project-id")?.value;
  const galleryId = document.getElementById("upload-gallery-id")?.value;

  //console.log("projectId:", projectId, "galleryId:", galleryId);

  Object.entries(selectedFiles).forEach(([fileId, file]) => {
    const formData = new FormData();
    formData.append("files[]", file);

    let uploadsRemaining = Object.keys(selectedFiles).length;

    if (projectId) formData.append("project_id", projectId);
    if (galleryId) formData.append("gallery_id", galleryId);

    // Show progress bar
    const progressContainer = document.getElementById(
      `progress-container-${fileId}`,
    );
    if (progressContainer) {
      progressContainer.classList.remove("hidden");
    }

    const status = document.getElementById(`status-${fileId}`);
    if (status) status.innerText = "Uploading...";

    const xhr = new XMLHttpRequest();
    xhr.open("POST", "/admin/media/upload", true);

    // Update progress bar
    xhr.upload.onprogress = function(e) {
      if (e.lengthComputable) {
        const percent = (e.loaded / e.total) * 100;
        const progressBar = document.getElementById(`progress-${fileId}`);
        if (progressBar) {
          progressBar.style.width = `${percent.toFixed(0)}%`;
        }
      }
    };

    // On upload complete
    xhr.onload = function() {
      if (status) {
        if (xhr.status === 200) {
          status.innerText = "Completed";

          //console.log("🚀 Response HTML:", xhr.responseText);
          const sortable = document.querySelector(".sortable");
          if (sortable && xhr.responseText.trim() !== "") {
            sortable.insertAdjacentHTML("beforeend", xhr.responseText);
          }
        } else {
          status.innerText = "Failed";
        }
        // Decrement and check if all are done
        uploadsRemaining--;
        if (uploadsRemaining === 0) {
          setTimeout(() => {
            closeModal();
          }, 500); // small delay feels smoother
        }
      }
    };

    //for (let [key, value] of formData.entries()) {
    //  console.log("🧾 FormData:", key, value);
    //}
    xhr.send(formData);
  });
}

function cancelUpload(fileId) {
  const xhr = uploadControllers[fileId];
  if (xhr) {
    xhr.abort();
    delete uploadControllers[fileId];
  }
}

document.addEventListener("htmx:afterOnLoad", function(evt) {
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger");

  if (trigger === "refresh-admin-grid") {
    const grid = document.getElementById("sortableGrid");
    const galleryID = grid?.dataset.gallery;
    const projectID = grid?.dataset.project;

    let url = "";
    if (galleryID) {
      url = `/admin/gallery/${galleryID}?partial=true`;
    } else if (projectID) {
      url = `/admin/project/edit/${projectID}?partial=true`;
    }

    if (url) {
      htmx.ajax("GET", url, { target: "#adminMediaGrid", swap: "innerHTML" });
    }
  }
});


// Clear contct form after submit

document.addEventListener("htmx:afterOnLoad", function(evt) {
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger");

  if (trigger === "form-submitted") {
    const form = document.getElementById("contact-form");
    if (form) {
      form.reset();
    }
  }
});

document.addEventListener("htmx:afterSwap", function(evt) {
  const firstError = document.querySelector(".field-error input, .field-error textarea");
  if (firstError) firstError.focus();
});

