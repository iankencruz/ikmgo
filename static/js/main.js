/* Mobile Sidebar Start */

function openSidebar() {
  const sidebar = document.getElementById("mobileSidebar");
  const backdrop = document.getElementById("mobileBackdrop");
  const sidebarPanel = sidebar.querySelector(".fixed.inset-0.flex");

  // Remove hidden so it's visible
  sidebar.classList.remove("hidden");

  // Slight delay to allow the DOM to apply hidden removal
  setTimeout(() => {
    // Slide in the panel
    sidebarPanel.classList.remove("-translate-x-full");
    sidebarPanel.classList.add("translate-x-0");
    // Fade in the backdrop
    backdrop.classList.remove("opacity-0", "pointer-events-none");
    backdrop.classList.add("opacity-100");
  }, 10);
}

function closeSidebar() {
  const sidebar = document.getElementById("mobileSidebar");
  const backdrop = document.getElementById("mobileBackdrop");
  const sidebarPanel = sidebar.querySelector(".fixed.inset-0.flex");

  // Slide out
  sidebarPanel.classList.remove("translate-x-0");
  sidebarPanel.classList.add("-translate-x-full");
  // Fade out backdrop
  backdrop.classList.remove("opacity-100");
  backdrop.classList.add("opacity-0", "pointer-events-none");

  // After 300ms (transition duration), hide the entire container
  setTimeout(() => {
    sidebar.classList.add("hidden");
  }, 300);
}

// ✅ Allow clicking backdrop to close
document.addEventListener("DOMContentLoaded", () => {
  const backdrop = document.getElementById("mobileBackdrop");
  if (backdrop) {
    backdrop.addEventListener("click", closeSidebar);
  }
});

/* Mobile Sidebar Start */


document.addEventListener("htmx:afterOnLoad", function(evt) {
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger-After-Settle");

  if (!trigger) return;

  const toastMap = {
    "show-toast": {
      variant: "success",
      heading: "Media Linked",
      subtitle: "Successfully attached to gallery or project"
    },
    "show-toast-unlinked": {
      variant: "warning",
      heading: "Media Unlinked",
      subtitle: "This media was removed from the gallery or project"
    }
  };

  const toast = toastMap[trigger];
  if (!toast) return;

  fetch(`/admin/toast?variant=${toast.variant}&heading=${encodeURIComponent(toast.heading)}&subtitle=${encodeURIComponent(toast.subtitle)}`)
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

