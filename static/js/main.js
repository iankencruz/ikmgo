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
